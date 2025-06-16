package finance

import (
	"fmt"
	"time"
)

type txnT int

const (
	IncomeT = iota
	ExpenseT
	TransferT
	InvestmentT
	LoanT
	OtherT
	UnknownT      = -1
	IncomeStr     = "income"
	ExpenseStr    = "expense"
	TransferStr   = "transfer"
	InvestmentStr = "investment"
	LoanStr       = "loan"
	OtherStr      = "other"
	UnknownStr    = "unknown"
)

func (t txnT) String() string {
	return [...]string{
		IncomeStr,
		ExpenseStr,
		TransferStr,
		InvestmentStr,
		LoanStr,
		OtherStr,
		UnknownStr,
	}[t]
}

// base transaction structure
type txn struct {
	t   txnT
	amt float64
	src string
	cat string

	note string
	ts   int64
}

func (t *txn) String() string {
	return fmt.Sprintf("%s %f %s %s %s %d", t.t, t.amt, t.src, t.cat, t.note, t.ts)
}

func (t *txn) SetType(ty txnT) {
	t.t = ty
}

func (t *txn) SetAmount(amt float64) {
	t.amt = amt
}

func (t *txn) SetSource(source string) {
	t.src = source
}

func (t *txn) SetCategory(category string) {
	t.cat = category
}

func (t *txn) SetNote(note string) {
	t.note = note
}

func NewTransaction(amt float64, source, category, note, t string) *txn {
	return &txn{
		t:    txnT(UnknownT),
		amt:  amt,
		src:  source,
		cat:  category,
		note: note,
		ts:   time.Now().UnixNano(),
	}
}

func NewIncome(amt float64, source, category, note string) *txn {
	return NewTransaction(amt, source, category, note, "in")
}

func NewExpense(amt float64, source, category, note string) *txn {
	return NewTransaction(amt, source, category, note, "out")

}
