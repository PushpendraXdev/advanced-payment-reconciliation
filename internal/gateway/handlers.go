package gateway

import (
	"context"
	"strings"
	"time"

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
    "INSERT INTO internal_transaction (order_id,amount, mode_of_payment, status_of_payment) VALUES ($1, $2, $3, $4) RETURNING id", req.OrderID,req.Amount,req.ModeOfPayment, "pending").Scan(&newID)
	if err!=nil{
		c.JSON(500,gin.H{"error":err.Error()})
		return
	}
	c.JSON(201, gin.H{"newid":newID,"message":"transaction created"})
	
}

type ListTransactions struct{
	Id int `json:"id"`
	 OrderID       int     `json:"order_id"`
    Amount        float64 `json:"amount"`
    ModeOfPayment string  `json:"mode_of_payment"`
	StatusOfPayment string `json:"status_of_payment"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *Handler) ListTransactions(c *gin.Context){
     rows,err:=h.Pool.Query(context.Background(),"SELECT id, order_id, amount, mode_of_payment, status_of_payment, created_at FROM internal_transaction")	

	  if err!=nil{
		c.JSON(500,gin.H{"error":err.Error()})
		return
	  }

	 defer rows.Close() 
	  var listcount []ListTransactions
	  for rows.Next(){
          var t ListTransactions
		  err:=rows.Scan(&t.Id, &t.OrderID, &t.Amount, &t.ModeOfPayment, &t.StatusOfPayment, &t.CreatedAt)
		  if err!=nil{
			c.JSON(500,gin.H{"error":err.Error()})
			return
		  }
		  listcount = append(listcount, t)
	  }
	  c.JSON(200, listcount)
	}
	type GatewayTransactionRequest struct{
    PaymentID     string  `json:"payment_id"`
    UserID        int     `json:"user_id"`
    Amount        float64 `json:"amount"`
    UTR           string  `json:"utr"`
    ModeOfPayment string  `json:"mode_of_payment"`
    GatewayName   string  `json:"gateway_name"`
}
func (h *Handler) HandleWebhook(c *gin.Context){
    var gtr GatewayTransactionRequest
	err:=c.ShouldBindJSON(&gtr)
	if err!=nil{
		c.JSON(400,gin.H{"error":err.Error()})
		return
	}

    var newId int
	err=h.Pool.QueryRow(context.Background(),"INSERT INTO gateway_transactions (payment_id, user_id,amount, utr, mode_of_payment,gateway_name) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id",gtr.PaymentID,gtr.UserID,gtr.Amount,gtr.UTR,gtr.ModeOfPayment,gtr.GatewayName).Scan(&newId)
if err!=nil{
        if strings.Contains(err.Error(),"duplicate key"){
			c.JSON(200,gin.H{"already processed":"skipping"})
			return
		}
		c.JSON(500,gin.H{"error":err.Error()})
		return
}
c.JSON(201,gin.H{"id":newId,"message":"gateway transaction recorded"})


}