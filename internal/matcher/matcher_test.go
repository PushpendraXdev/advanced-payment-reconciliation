package matcher

import (
	"testing"
)

func TestMatchTransactions(t *testing.T) {
	tests := []struct {
		name        string
		internalTxn []InternalTxn
		gatewayTxn  []GatewayTxn
		expected    int
	}{

		{
			name:        "normal exact match",
			internalTxn: []InternalTxn{{ID: 1, Amount: 500.00}},
			gatewayTxn:  []GatewayTxn{{ID: 1, Amount: 500.00}},
			expected:    1,
		},
		{
			name:        "no match different amounts",
			internalTxn: []InternalTxn{{ID: 1, Amount: 500.00}},
			gatewayTxn:  []GatewayTxn{{ID: 1, Amount: 600.00}},
			expected:    0,
		}, {
			name:        "duplicate gateway payment",
			internalTxn: []InternalTxn{{ID: 1, Amount: 500.00}},
			gatewayTxn:  []GatewayTxn{{ID: 1, Amount: 500.00}, {ID: 2, Amount: 500.00}},
			expected:    1,
		},
		{
			name:        "empty lists",
			internalTxn: []InternalTxn{},
			gatewayTxn:  []GatewayTxn{},
			expected:    0,
		},
		{
			name:        "partial match - some match some dont",
			internalTxn: []InternalTxn{{ID: 1, Amount: 500.00}, {ID: 2, Amount: 700.00}},
			gatewayTxn:  []GatewayTxn{{ID: 1, Amount: 500.00}},
			expected:    1,
		},
		{
			name:        "zero amount matches",
			internalTxn: []InternalTxn{{ID: 1, Amount: 0}},
			gatewayTxn:  []GatewayTxn{{ID: 1, Amount: 0}},
			expected:    1,
		},
		{
			name:        "negative amount still matches if equal",
			internalTxn: []InternalTxn{{ID: 1, Amount: -100}},
			gatewayTxn:  []GatewayTxn{{ID: 1, Amount: -100}},
			expected:    1,
		},
		{
			name: "multiple internal, multiple gateway match",
			internalTxn: []InternalTxn{{ID:1,Amount: 500.00},{ID:9,Amount: 800.00},{ID:19,Amount: 1900.00}},
			gatewayTxn: []GatewayTxn{{ID:1,Amount: 500.00},{ID:9,Amount: 800.00},{ID:19,Amount: 1900.00}},
			expected: 3,

		},
	}

	for _, cases := range tests {
		t.Run(cases.name, func(t *testing.T) {
			result := MatchTransactions(cases.internalTxn, cases.gatewayTxn)
			if len(result) != cases.expected {
				t.Errorf("expected %d matches, got %d match", cases.expected, len(result))
			}
		})
	}
}
