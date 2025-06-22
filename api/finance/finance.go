package finance

import (
	"fmt"
	"sync"

	"github.com/danmuck/dps_http/lib/storage"
)

// get - Transaction and expenses by year, month, day
// post - add income or expense for a specific date
// put - update income or expense for a specific date
// delete - remove income or expense for a specific date
// props ->
// - year
// - month
// - day
// - type (income/expense)
// - amount
// - description
// - category (e.g., food, transport, etc.)

type Portfolio struct {
	ID          string
	Name        string
	Description string
	Assets      []string // List of asset IDs
}

type FinanceService struct {
	version  string
	endpoint string
	running  bool
	buckets  []*storage.Bucket

	//        YEAR -> MONTH -> DAY -> txn
	txns map[string]map[string]map[string]*txn
	mu   sync.Mutex
}

func (svc *FinanceService) String() string {
	if svc == nil {
		return "FinanceService: <nil>"
	}
	svc.mu.Lock()
	defer svc.mu.Unlock()
	return `
	FinanceService:
		Version: ` + svc.version + `
		Endpoint: ` + svc.endpoint + `
		Running: ` + fmt.Sprintf("%t", svc.running) + `
	`
}

func NewFinanceService(store ...storage.Bucket) *FinanceService {
	return &FinanceService{
		version:  "1.0.0",
		endpoint: "/finance",
		running:  false,
		buckets:  make([]*storage.Bucket, 0),

		txns: make(map[string]map[string]map[string]*txn),
	}
}
