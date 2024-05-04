package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Static("/front", "./front")

	dataSourceName := "host=db user=postgres dbname=financial_system password=postgres sslmode=disable"
	initDB(dataSourceName)

	r.GET("/", func(c *gin.Context) {
        c.File("./front/index.html")
    })

	r.POST("/accounts", createAccount)
	r.GET("/accounts", getAllAccounts)
	r.POST("/transactions", createTransaction)
	r.GET("/accounts/:id/transactions", getAccountTransactions)

	r.Run()
}
