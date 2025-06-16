package finance

import (
	"sync"

	"github.com/danmuck/dps_http/storage"
)

type FinanceService struct {
	version  string
	endpoint string
	running  bool
	buckets  []*storage.Bucket

	income   map[string]*Income
	expenses map[string]*Expense

	mu sync.Mutex
}

func NewFinanceService(store ...storage.Bucket) *FinanceService {
	return &FinanceService{
		version:  "1.0.0",
		endpoint: "/finance",
		running:  false,
		buckets:  make([]*storage.Bucket, 0),

		income:   make(map[string]*Income),
		expenses: make(map[string]*Expense),
	}
}
