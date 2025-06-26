package services

import (
	"errors"
	"math"
	"sync"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/repository"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/sirupsen/logrus"
)

var (
	ErrTransactionAlreadySettled = errors.New("transaction already settled")
	ErrInvalidPayer              = errors.New("invalid payer")
)

type TransactionService struct {
	transactionRepo repository.TransactionRepository
	billRepo        repository.BillRepository
	logger          *logrus.Entry
}

type edge struct {
	userId uint
	amount float32
}

// NOTE: used var instead of func to enable mocking in tests
var NewTransactionService = func(
	transactionRepo repository.TransactionRepository,
	billRepo repository.BillRepository,
	logger *logrus.Logger,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		billRepo:        billRepo,
		logger:          utils.AddServiceField(logger, "TransactionService"),
	}
}

func (ts *TransactionService) GenerateTransactions(bills *[]model.Bill, consolidatedBill *model.Consolidation) (*[]model.Transaction, error) {
	var wg sync.WaitGroup
	txChan := make(chan *model.Transaction)

	for _, bill := range *bills {
		wg.Add(1)
		go func(bill *model.Bill) {
			defer wg.Done()
			ts.logger.Infof("Processing bill: %s with amount %f\n", bill.Name, bill.Amount)
			ts.logger.Info("Payers: ", len(bill.Payers))

			num_payers := float32(len(bill.Payers))
			if bill.IncludeOwner {
				num_payers += 1
			}

			// hack to round numbers to 2 decimal places
			transactionAmt := float32(math.Floor(float64(bill.Amount/num_payers)*100) / 100)
			for _, payers := range bill.Payers {
				transaction := model.Transaction{
					ConsolidationID: consolidatedBill.ID,
					Payer:           payers,
					PayerID:         payers.ID,
					Payee:           bill.Owner,
					PayeeID:         bill.OwnerID,
					Amount:          transactionAmt,
				}
				txChan <- &transaction
			}
		}(&bill)
	}

	go func() {
		wg.Wait()
		close(txChan)
	}()

	var transactions []model.Transaction
	// BLOCKING until all goroutines finish
	for transaction := range txChan {
		transactions = append(transactions, *transaction)
	}

	for _, transaction := range transactions {
		ts.logger.Debugf("Before %d -> %d : %f\n", transaction.PayerID, transaction.PayeeID, transaction.Amount)
	}

	consolidatedTransactions := ts.consolidateTransactions(&transactions, consolidatedBill)
	for _, transaction := range *consolidatedTransactions {
		ts.logger.Debugf("After %d -> %d : %f\n", transaction.PayerID, transaction.PayeeID, transaction.Amount)
	}

	return consolidatedTransactions, nil
}

func (ts *TransactionService) GetTransactionsByUser(isPaid bool, userId string) (*[]model.Transaction, error) {
	return ts.transactionRepo.FindByUser(isPaid, userId)
}

func (ts *TransactionService) SettleTransaction(transactionId string, userId string) (*model.Transaction, error) {
	transaction, err := ts.transactionRepo.FindByID(transactionId)
	if err != nil {
		return nil, err
	}

	if transaction.IsPaid {
		return nil, ErrTransactionAlreadySettled
	}

	if utils.UIntToString(transaction.PayerID) != userId {
		return nil, ErrInvalidPayer
	}

	transaction.IsPaid = true
	if err := ts.transactionRepo.Update(transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (ts *TransactionService) consolidateTransactions(transactions *[]model.Transaction, consolidatedBill *model.Consolidation) *[]model.Transaction {
	graph := make(map[uint][]edge)
	visited := make(map[uint]bool)

	// construct adjacency list and init visited set
	for _, transaction := range *transactions {
		startNode := transaction.PayerID
		endNode := edge{
			userId: transaction.PayeeID,
			amount: transaction.Amount,
		}

		graph[startNode] = append(graph[startNode], endNode)
		visited[startNode] = false
	}

	var hasCycle float32
	for _, transaction := range *transactions {
		// trigger the do-while loop
		hasCycle = 1
		for hasCycle != -1 {
			// 0 -> Start from source node
			visited[transaction.PayerID] = true
			hasCycle, _ = ts.removeCycle(transaction.PayerID, graph, visited)
			ts.resetVisited(visited)
		}
	}

	var newTransactions []model.Transaction
	for startNode, edges := range graph {
		for _, edge := range edges {
			transaction := model.Transaction{
				ConsolidationID: consolidatedBill.ID,
				PayerID:         startNode,
				PayeeID:         edge.userId,
				Amount:          edge.amount,
			}
			newTransactions = append(newTransactions, transaction)
		}
	}

	return &newTransactions
}

func (ts *TransactionService) resetVisited(visited map[uint]bool) {
	for key := range visited {
		visited[key] = false
	}
}

func (ts *TransactionService) removeCycle(startNode uint, graph map[uint][]edge, visited map[uint]bool) (float32, uint) {
	neighbors := graph[startNode]
	for i, neighbor := range neighbors {
		ts.logger.Debug("Current node ", startNode, " with neighbor: ", neighbor.userId)

		// cycle detected
		if isVisited := visited[neighbor.userId]; isVisited {
			ts.logger.Debug("Cycle detected: ", startNode, " -> ", neighbor.userId)

			// remove the edge
			neighbors = append(neighbors[:i], neighbors[i+1:]...)
			graph[startNode] = neighbors

			// return amount to deduct
			return neighbor.amount, neighbor.userId
		}

		visited[neighbor.userId] = true
		amtToDeduct, stopAt := ts.removeCycle(neighbor.userId, graph, visited)
		visited[neighbor.userId] = false

		// no cycle -> nothing to deduct
		if stopAt == 0 {
			continue
		}

		if (neighbor.amount - amtToDeduct) > 0 {
			// deduct the amount of the cycle's edge
			neighbors[i] = edge{
				userId: neighbor.userId,
				amount: neighbor.amount - amtToDeduct,
			}
		} else if (neighbor.amount - amtToDeduct) < 0 {
			// -ve -> remove this edge and add it in opposite direction
			neighbors = append(neighbors[:i], neighbors[i+1:]...)
			graph[startNode] = neighbors

			newEndNode := edge{
				userId: startNode,
				amount: amtToDeduct - neighbor.amount,
			}
			graph[neighbor.userId] = append(graph[neighbor.userId], newEndNode)
		} else {
			// 0 -> remove the edge
			neighbors = append(neighbors[:i], neighbors[i+1:]...)
			graph[startNode] = neighbors
		}

		// cycle is resolved
		// still return amount to deduct for retry in parent call
		if stopAt == startNode {
			return amtToDeduct, 0
		} else {
			return amtToDeduct, stopAt
		}
	}

	return -1, 0
}
