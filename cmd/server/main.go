package main

import (
	"log"

	"github.com/PushpendraXdev/advanced-payment-reconciliation/internal/db"
	"github.com/PushpendraXdev/advanced-payment-reconciliation/internal/gateway"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := db.NewPool()
	if err!=nil{
		log.Fatal(err)
	}
	defer pool.Close()
	r := gin.Default()
	r.GET("/health", func(ctx *gin.Context) { ctx.JSON(200, "welcome to project") })
	 handler := &gateway.Handler{Pool: pool}
   r.POST("/transactions/internal", handler.CreateTransaction)
   
	r.Run(":8080")
  
}
