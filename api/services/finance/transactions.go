package finance

import (
	"fmt"
	"time"
)

type Transaction struct {
	Type     string
	Amount   float64
	Source   string
	Category string

	Note      string
	Timestamp int64
}

func (t *Transaction) String() string {
	return fmt.Sprintf("%s %f %s %s %s %d", t.Type, t.Amount, t.Source, t.Category, t.Note, t.Timestamp)
}

func (t *Transaction) SetType(ty string) {
	t.Type = ty
}

func (t *Transaction) SetAmount(amt float64) {
	t.Amount = amt
}

func (t *Transaction) SetSource(source string) {
	t.Source = source
}

func (t *Transaction) SetCategory(category string) {
	t.Category = category
}

func (t *Transaction) SetNote(note string) {
	t.Note = note
}

func NewTransaction(amt float64, source, category, note, t string) *Transaction {
	return &Transaction{
		Type:      t,
		Amount:    amt,
		Source:    source,
		Category:  category,
		Note:      note,
		Timestamp: time.Now().UnixNano(),
	}
}

type Income struct {
	*Transaction
}

func NewIncome(amt float64, source, category, note string) *Income {
	return &Income{
		Transaction: NewTransaction(amt, source, category, note, "in"),
	}
}

type Expense struct {
	*Transaction
}

func NewExpense(amt float64, source, category, note string) *Expense {
	return &Expense{
		Transaction: NewTransaction(amt, source, category, note, "out"),
	}
}
