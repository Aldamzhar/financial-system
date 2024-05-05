package main

import (
	"database/sql"
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

func updateTransaction(c *gin.Context) {
    var updatedTransaction Transaction
    transactionId := c.Param("id")
    if err := c.ShouldBindJSON(&updatedTransaction); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    err := updateTransactionInDB(transactionId, updatedTransaction)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"success": true})
}

func deleteTransaction(c *gin.Context) {
    transactionId := c.Param("id")
    err := deleteTransactionFromDB(transactionId)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"success": true})
}


func updateTransactionInDB(transactionId string, transaction Transaction) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    var oldTransaction Transaction
    err = tx.QueryRow("SELECT value, account_id, group_type, account2_id FROM transactions WHERE id = $1", transactionId).Scan(&oldTransaction.Value, &oldTransaction.AccountID, &oldTransaction.GroupType, &oldTransaction.Account2ID)
    if err != nil {
        return err
    }

    _, err = tx.Exec("UPDATE transactions SET value = $1, account_id = $2, group_type = $3, account2_id = $4, transaction_date = $5 WHERE id = $6",
        transaction.Value, transaction.AccountID, transaction.GroupType, transaction.Account2ID, transaction.TransactionDate, transactionId)
    if err != nil {
        return err
    }

    switch oldTransaction.GroupType {
    case "income":
        _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", oldTransaction.Value, oldTransaction.AccountID)
        if err != nil {
            return err
        }
    case "transfer":
        _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", oldTransaction.Value, oldTransaction.AccountID)
        if err != nil {
            return err
        }
        _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", oldTransaction.Value, oldTransaction.Account2ID)
        if err != nil {
            return err
        }
    case "outcome":
        _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", oldTransaction.Value, oldTransaction.AccountID)
        if err != nil {
            return err
        }
    }

    switch transaction.GroupType {
    case "income":
        _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", transaction.Value, transaction.AccountID)
        if err != nil {
            return err
        }
    case "transfer":
        _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", transaction.Value, transaction.AccountID)
        if err != nil {
            return err
        }
        _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", transaction.Value, transaction.Account2ID)
        if err != nil {
            return err
        }
    case "outcome":
        _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", transaction.Value, transaction.AccountID)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}




func deleteTransactionFromDB(transactionId string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    var value float64
	var accountId int
    var account2Id sql.NullInt64
    var groupType string
    err = tx.QueryRow("SELECT value, account_id, group_type, account2_id FROM transactions WHERE id = $1", transactionId).Scan(&value, &accountId, &groupType, &account2Id)
    if err != nil {
        return err
    }

    _, err = tx.Exec("DELETE FROM transactions WHERE id = $1", transactionId)
    if err != nil {
        return err
    }

    if groupType == "transfer" {
        _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", value, accountId)
        if err != nil {
            return err
        }

        _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", value, account2Id)
        if err != nil {
            return err
        }
    } else if groupType == "income" {
        _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", value, accountId)
        if err != nil {
            return err
        }
    } else if groupType == "outcome" {
        _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", value, accountId)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
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
