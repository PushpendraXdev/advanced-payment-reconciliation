package gateway

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)
type CreateTransactionRequest struct{
	 OrderID       int     `json:"order_id"`
    Amount        float64 `json:"amount"`
    ModeOfPayment string  `json:"mode_of_payment"`
}



type Handler struct {
    Pool *pgxpool.Pool
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	 var req CreateTransactionRequest
	err:=c.ShouldBindJSON(&req)
	if err!=nil{
	c.JSON(400, gin.H{"error": err.Error()})
    return
	}
	var newID int
	err=h.Pool.QueryRow(context.Background(),
    "INSERT INTO internal_transaction (order_id,amount, mode_of_payment) VALUES ($1, $2, $3) RETURNING id", req.OrderID,req.Amount,req.ModeOfPayment).Scan(&newID)
	if err!=nil{
		c.JSON(500,gin.H{"error":err.Error()})
		return
	}
	c.JSON(201, gin.H{"newid":newID,"message":"transaction created"})
	
}