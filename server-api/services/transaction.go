package services

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/RowenTey/JustJio/model"

	"gorm.io/gorm"
)

type TransactionService struct {
	DB *gorm.DB
}

type edge struct {
	userId uint
	amount float32
}

func (ts *TransactionService) GenerateTransactions(consolidatedBill *model.Consolidation) error {
	billsDb := ts.DB.Table("bills")
	var bills []model.Bill

	if err := billsDb.
		Where("consolidation_id = ?", (*consolidatedBill).ID).
		Preload("Payers").
		Find(&bills).Error; err != nil {
		return err
	}

	var wg sync.WaitGroup
	txChan := make(chan *model.Transaction)

	for _, bill := range bills {
		wg.Add(1)
		go func(bill *model.Bill) {
			defer wg.Done()
			log.Printf("[TRANSACTION] Processing bill: %s with amount %f\n", bill.Name, bill.Amount)
			log.Println("[TRANSACTION] Payers: ", len(bill.Payers))

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
		log.Printf("[TRANSACTION] Before %d -> %d : %f\n", transaction.PayerID, transaction.PayeeID, transaction.Amount)
	}

	consolidatedTransactions := consolidateTransactions(&transactions, consolidatedBill)
	for _, transaction := range *consolidatedTransactions {
		log.Printf("[TRANSACTION] After %d -> %d : %f\n", transaction.PayerID, transaction.PayeeID, transaction.Amount)
	}
	if err := ts.DB.Omit("Consolidation").Create(&consolidatedTransactions).Error; err != nil {
		return err
	}

	return nil
}

func (ts *TransactionService) GetTransactionsByUser(isPaid bool, userId string) (*[]model.Transaction, error) {
	db := ts.DB.Table("transactions")
	var transactions []model.Transaction

	// TODO: Implement pagination
	if err := db.
		Where("is_paid = ? AND (payee_id = ? OR payer_id = ?)", isPaid, userId, userId).
		Preload("Payee").
		Preload("Payer").
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	return &transactions, nil
}

func (ts *TransactionService) SettleTransaction(transactionId string, userId string) (*model.Transaction, error) {
	db := ts.DB.Table("transactions")
	var transaction model.Transaction

	if err := db.First(&transaction, transactionId).Error; err != nil {
		return nil, err
	}

	if transaction.IsPaid {
		return nil, errors.New("transaction already settled")
	}

	if fmt.Sprint(transaction.PayerID) != userId {
		return nil, errors.New("invalid payer")
	}

	transaction.IsPaid = true
	if err := ts.DB.Save(&transaction).Error; err != nil {
		return nil, err
	}

	return &transaction, nil
}

func consolidateTransactions(transactions *[]model.Transaction, consolidatedBill *model.Consolidation) *[]model.Transaction {
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
			hasCycle, _ = removeCycle(transaction.PayerID, graph, visited)
			resetVisited(visited)
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

func resetVisited(visited map[uint]bool) {
	for key := range visited {
		visited[key] = false
	}
}

func removeCycle(startNode uint, graph map[uint][]edge, visited map[uint]bool) (float32, uint) {
	neighbors := graph[startNode]
	for i, neighbor := range neighbors {
		log.Println("[DEBUG] Current node ", startNode, " with neighbor: ", neighbor.userId)

		// cycle detected
		if isVisited := visited[neighbor.userId]; isVisited {
			log.Println("[DEBUG] Cycle detected: ", startNode, " -> ", neighbor.userId)

			// remove the edge
			neighbors = append(neighbors[:i], neighbors[i+1:]...)
			graph[startNode] = neighbors

			// return amount to deduct
			return neighbor.amount, neighbor.userId
		}

		visited[neighbor.userId] = true
		amtToDeduct, stopAt := removeCycle(neighbor.userId, graph, visited)
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
