package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBalanceDelta_Income(t *testing.T) {
	delta := CalculateBalanceDelta(TxTypeIncome, 50000)
	assert.Equal(t, int64(50000), delta)
}

func TestCalculateBalanceDelta_Expense(t *testing.T) {
	delta := CalculateBalanceDelta(TxTypeExpense, 30000)
	assert.Equal(t, int64(-30000), delta)
}

func TestCalculateBalanceDelta_Adjustment(t *testing.T) {
	delta := CalculateBalanceDelta(TxTypeAdjustment, 100000)
	assert.Equal(t, int64(100000), delta)
}

func TestCalculateBalanceDelta_Transfer(t *testing.T) {
	delta := CalculateBalanceDelta(TxTypeTransfer, 25000)
	assert.Equal(t, int64(0), delta)
}

func TestCalculateBalanceDelta_UnknownType(t *testing.T) {
	delta := CalculateBalanceDelta(TransactionType("INVALID"), 10000)
	assert.Equal(t, int64(0), delta)
}
