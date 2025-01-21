package services

import (
	"errors"
	"fmt"
	"log"
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
			log.Println("[TRANSACTION] Processing bill: ", bill)
			log.Println("[TRANSACTION] Payers: ", bill.Payers)

			num_payers := float32(len(bill.Payers))
			if bill.IncludeOwner {
				num_payers += 1
			}

			transactionAmt := bill.Amount / num_payers
			for _, payers := range bill.Payers {
				transaction := model.Transaction{
					Consolidation: *consolidatedBill,
					Payer:         payers,
					PayerID:       payers.ID,
					Payee:         bill.Owner,
					PayeeID:       bill.OwnerID,
					Amount:        transactionAmt,
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
	log.Println("[TRANSACTION] Transactions: ", transactions)

	consolidatedTransactions := consolidateTransactions(&transactions, consolidatedBill)
	log.Println("[TRANSACTION] Consolidated Transactions: ", transactions)
	if err := ts.DB.Create(&consolidatedTransactions).Error; err != nil {
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

func (ts *TransactionService) SettleTransaction(transactionId string, userId string) error {
	db := ts.DB.Table("transactions")
	var transaction model.Transaction

	if err := db.First(&transaction, transactionId).Error; err != nil {
		return err
	}

	if transaction.IsPaid {
		return errors.New("Transaction already settled")
	}

	if fmt.Sprint(transaction.PayerID) != userId {
		return errors.New("Invalid payer")
	}

	transaction.IsPaid = true
	if err := ts.DB.Save(&transaction).Error; err != nil {
		return err
	}

	return nil
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
			log.Println("[DEBUG] Graph before: ", graph)
			visited[transaction.PayerID] = true
			hasCycle = removeCycle(transaction.PayerID, graph, visited)
			resetVisited(visited)
			log.Println("[DEBUG] Graph after: ", graph)
		}
	}

	var newTransactions []model.Transaction
	for startNode, edges := range graph {
		for _, edge := range edges {
			transaction := model.Transaction{
				Consolidation: *consolidatedBill,
				PayerID:       startNode,
				PayeeID:       edge.userId,
				Amount:        edge.amount,
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

func removeCycle(startNode uint, graph map[uint][]edge, visited map[uint]bool) float32 {
	log.Println("[DEBUG] Current node: ", startNode)

	neighbors := graph[startNode]
	for i, neighbor := range neighbors {
		// cycle detected
		if isVisited := visited[neighbor.userId]; isVisited {
			// remove the edge
			neighbors = append(neighbors[:i], neighbors[i+1:]...)
			graph[startNode] = neighbors

			// return amount to deduct
			return neighbor.amount
		}

		visited[neighbor.userId] = true
		amtToDeduct := removeCycle(neighbor.userId, graph, visited)
		// no cycle -> nothing to deduct
		if amtToDeduct == -1 {
			continue
		}

		if (neighbor.amount - amtToDeduct) > 0 {
			// deduct the amount of the cycle's edge
			neighbors[i] = edge{
				userId: neighbor.userId,
				amount: neighbor.amount - amtToDeduct,
			}
		} else {
			// -ve -> remove this edge and add it in opposite direction
			neighbors = append(neighbors[:i], neighbors[i+1:]...)
			graph[startNode] = neighbors

			newEndNode := edge{
				userId: startNode,
				amount: amtToDeduct - neighbor.amount,
			}
			graph[neighbor.userId] = append(graph[neighbor.userId], newEndNode)
		}

		// bubble the amount to deduct back to path
		return amtToDeduct
	}

	return -1
}
