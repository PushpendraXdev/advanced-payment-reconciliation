package matcher

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InternalTxn struct {
	ID     int
	Amount float64
}

type GatewayTxn struct {
	ID     int
	Amount float64
}

type MatchResult struct {
	InternalID int
	GatewayID  int
}


func MatchTransactions(internalTxns []InternalTxn, gatewayTxns []GatewayTxn) []MatchResult {
	var results []MatchResult

	gatewayMap := make(map[int64][]GatewayTxn)
	for _, g := range gatewayTxns {
		key := int64(g.Amount * 100)
		gatewayMap[key] = append(gatewayMap[key], g)

	}
	for _, i := range internalTxns {
		key := int64(i.Amount * 100)
		if matches, found := gatewayMap[key]; found && len(matches) > 0 {
			matchedtxn := matches[0]
			results = append(results, MatchResult{
				InternalID: i.ID,
				GatewayID:  matchedtxn.ID,
			})
			gatewayMap[key] = matches[1:]
		}
	}

	return results
}

type Matcher struct {
	Pool *pgxpool.Pool
}

func (m *Matcher) RunReconciliation(ctx context.Context) error {
    internalTxn,err:=m.FetchPendingInternal(ctx)
	if err!=nil{
		return err
	}
	gatewayTxn,err:=m.FetchUnlinkedGateway(ctx)
	if err!=nil{
		return err
	}
	results:=MatchTransactions(internalTxn,gatewayTxn)
	for _,r:=range results{
		_,err:=m.Pool.Exec(ctx,"UPDATE internal_transaction SET status_of_payment='matched' WHERE id=$1",r.InternalID)
		if err!=nil{
			return err
		}
		_,err=m.Pool.Exec(ctx,"UPDATE gateway_transactions SET internal_transaction_id=$1 WHERE id=$2",r.InternalID,r.GatewayID)
		if err!=nil{
			return err
		}
	}
	return nil
}

func (m *Matcher) FetchPendingInternal(ctx context.Context) ([]InternalTxn, error) {
	rows, err := m.Pool.Query(ctx, "SELECT id, amount FROM internal_transaction WHERE status_of_payment='pending'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var listcount []InternalTxn
	for rows.Next() {
		var t InternalTxn
		err := rows.Scan(&t.ID, &t.Amount)
		if err != nil {
			return nil, err
		}
		listcount = append(listcount, t)
	}
	return listcount, nil

}
func (m *Matcher) FetchUnlinkedGateway(ctx context.Context) ([]GatewayTxn, error) {
	rows, err := m.Pool.Query(ctx, "SELECT id, amount FROM gateway_transactions WHERE internal_transaction_id IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var listcount []GatewayTxn
	for rows.Next() {
		var t GatewayTxn
		err := rows.Scan(&t.ID, &t.Amount)
		if err != nil {
			return nil, err
		}
		listcount = append(listcount, t)
	}
	return listcount, nil
}

func (m *Matcher) ReconcileHandler(c *gin.Context){
    err:=m.RunReconciliation(c.Request.Context())
	if err!=nil{
		c.JSON(500,gin.H{"error":err.Error()})
	    return
	}
	c.JSON(200,gin.H{"message":"reconcile success"})
}