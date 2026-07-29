package main

import (
	"context"
	"log"
	"time"

	"github.com/PushpendraXdev/advanced-payment-reconciliation/internal/db"
	"github.com/PushpendraXdev/advanced-payment-reconciliation/internal/gateway"
	"github.com/PushpendraXdev/advanced-payment-reconciliation/internal/matcher"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := db.NewPool()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	r := gin.Default()

	r.GET("/health", func(ctx* gin.Context) { ctx.JSON(200, "welcome to project") })
	handler := &gateway.Handler{Pool: pool}
	r.POST("/transactions/internal", handler.CreateTransaction)
	r.GET("/transactions", handler.ListTransactions)
	r.POST("/gateway/webhook", handler.HandleWebhook)
	matcher := &matcher.Matcher{Pool: pool}
	r.POST("/reconcile/run", matcher.ReconcileHandler)
    
	go func() {
		timer:=time.NewTicker(30*time.Second)
		for{
			<-timer.C
		err:=matcher.RunReconciliation(context.Background())
		if err!=nil{
			log.Println("reconcile automate error:",err)
		
		}
	}
    }()

     r.GET("/matches/pending-approval",matcher.ListPendngApprovals)
	 r.POST("/matches/:id/approve",matcher.ApproveMatch)
	r.Run(":8080")

}
