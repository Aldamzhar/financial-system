package main

import (
	"time"
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
)

func createAccount(c *gin.Context) {
	var newAccount Account
	if err := c.ShouldBindJSON(&newAccount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sqlStatement := `INSERT INTO accounts (name, balance) VALUES ($1, $2) RETURNING id`
	if err := db.QueryRow(sqlStatement, newAccount.Name, newAccount.Balance).Scan(&newAccount.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newAccount)
}

func getAllAccounts(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, balance FROM accounts")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.ID, &acc.Name, &acc.Balance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		accounts = append(accounts, acc)
	}

	c.JSON(http.StatusOK, accounts)
}

func createTransaction(c *gin.Context) {
	var newTransaction Transaction
	if err := c.ShouldBindJSON(&newTransaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newTransaction.TransactionDate = time.Now()

	fmt.Println("transaction created");

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO transactions (value, account_id, group_type, account2_id, transaction_date) VALUES ($1, $2, $3, $4, $5) RETURNING id")
	if err != nil {
		fmt.Println("Error preparing statement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	var id int
	fmt.Println("Transaction Details:",
    "Value:", newTransaction.Value,
    "AccountID:", newTransaction.AccountID,
    "GroupType:", newTransaction.GroupType,
    "Account2ID:", newTransaction.Account2ID,
    "TransactionDate:", newTransaction.TransactionDate)
	err = stmt.QueryRow(newTransaction.Value, newTransaction.AccountID, newTransaction.GroupType, newTransaction.Account2ID, newTransaction.TransactionDate).Scan(&id)
	if err != nil {
		fmt.Println("Error executing query: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newTransaction.ID = id

	if newTransaction.GroupType == "transfer" && newTransaction.Account2ID != nil {
		_, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", newTransaction.Value, newTransaction.AccountID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
		}

		_, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", newTransaction.Value, *newTransaction.Account2ID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if newTransaction.GroupType == "income" {
		_, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", newTransaction.Value, newTransaction.AccountID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if newTransaction.GroupType == "outcome" {
		_, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", newTransaction.Value, newTransaction.AccountID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newTransaction)
}

func getAccountTransactions(c *gin.Context) {
	accountID := c.Param("id")

	rows, err := db.Query("SELECT id, value, account_id, group_type, account2_id, transaction_date FROM transactions WHERE account_id = $1", accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		var tr Transaction
		if err := rows.Scan(&tr.ID, &tr.Value, &tr.AccountID, &tr.GroupType, &tr.Account2ID, &tr.TransactionDate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		transactions = append(transactions, tr)
	}

	c.JSON(http.StatusOK, transactions)
}
