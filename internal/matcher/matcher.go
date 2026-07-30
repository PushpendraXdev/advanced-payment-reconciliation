package matcher

import (
	"context"
	"strconv"
	"time"

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
	internalTxn, err := m.FetchPendingInternal(ctx)
	if err != nil {
		return err
	}
	gatewayTxn, err := m.FetchUnlinkedGateway(ctx)
	if err != nil {
		return err
	}
	results := MatchTransactions(internalTxn, gatewayTxn)
	for _, r := range results {
		_, err := m.Pool.Exec(ctx, "UPDATE internal_transaction SET status_of_payment='pending_approval' WHERE id=$1", r.InternalID)
		if err != nil {
			return err
		}
		_, err = m.Pool.Exec(ctx, "UPDATE gateway_transactions SET internal_transaction_id=$1 WHERE id=$2", r.InternalID, r.GatewayID)
		if err != nil {
			return err
		}
		
		var matchID int
		err = m.Pool.QueryRow(ctx, "INSERT INTO matches (internal_transaction_id, gateway_transaction_id) VALUES ($1, $2) RETURNING id", r.InternalID, r.GatewayID).Scan(&matchID)
		if err != nil {
			return err
		}

		_, err = m.Pool.Exec(ctx, "INSERT INTO audit_logs (matched_id, internal_transaction_id, gateway_transaction_id, action, status) VALUES ($1, $2, $3, $4, $5)", matchID, r.InternalID, r.GatewayID, "matched", "pending_approval")
		if err != nil {
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

func (m *Matcher) ReconcileHandler(c *gin.Context) {
	err := m.RunReconciliation(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "reconcile success"})
}

type PendingApproval struct {
	Id            int       `json:"id"`
	InternalTxnId int       `json:"internal_transaction_id"`
	GatewayTxnId  int       `json:"gateway_transaction_id"`
	MatchedAt     time.Time `json:"matched_at"`
}

func (m *Matcher) ListPendngApprovals(c *gin.Context) {
	rows, err := m.Pool.Query(c.Request.Context(), "SELECT id,internal_transaction_id,gateway_transaction_id,matched_at FROM matches WHERE status ='pending_approval'")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []PendingApproval
	for rows.Next() {
		var pm PendingApproval
		err = rows.Scan(&pm.Id, &pm.InternalTxnId, &pm.GatewayTxnId, &pm.MatchedAt)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		list = append(list, pm)
	}
	c.JSON(200, list)
}

func (m *Matcher) ApproveMatch(c *gin.Context) {
	idParam := c.Param("id")
	matchID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid match id"})
		return
	}

	var internalID int
	err = m.Pool.QueryRow(c.Request.Context(), "SELECT internal_transaction_id FROM matches WHERE id=$1", matchID).Scan(&internalID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, err = m.Pool.Exec(c.Request.Context(), "UPDATE matches SET status='approved', approved_at=NOW() WHERE id=$1", matchID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, err = m.Pool.Exec(c.Request.Context(), "UPDATE internal_transaction SET status_of_payment='matched' WHERE id=$1", internalID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "match approved"})
}

func (m *Matcher) ApproveBatch(c *gin.Context) {
	_, err := m.Pool.Exec(c.Request.Context(), "UPDATE matches SET status='approved', approved_at=NOW() WHERE status='pending_approval'")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	_, err = m.Pool.Exec(c.Request.Context(), "UPDATE internal_transaction SET status_of_payment='matched' WHERE id IN (SELECT internal_transaction_id FROM matches WHERE status='approved')")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "batch approved"})
}

func (m *Matcher) RejectMatch(c *gin.Context) {
	idParam := c.Param("id")
	matchId, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var internalId int
	var gatewayId int
	err = m.Pool.QueryRow(c.Request.Context(), "SELECT internal_transaction_id, gateway_transaction_id FROM matches WHERE id=$1", matchId).Scan(&internalId, &gatewayId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, err = m.Pool.Exec(c.Request.Context(), "UPDATE matches SET status='rejected' WHERE id=$1", matchId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, err = m.Pool.Exec(c.Request.Context(), "UPDATE internal_transaction SET status_of_payment='pending' WHERE id=$1", internalId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, err = m.Pool.Exec(c.Request.Context(), "UPDATE gateway_transactions SET internal_transaction_id=NULL WHERE id=$1", gatewayId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "reject match done"})
}
