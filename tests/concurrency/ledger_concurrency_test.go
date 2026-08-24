package concurrency

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/api-monetization-platform/internal/ledger"
)

type fakeRepo struct {
	mu      sync.Mutex
	wallets map[string]ledger.Wallet
	seen    map[string]ledger.LedgerEntry
	nextID  int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{wallets: map[string]ledger.Wallet{}, seen: map[string]ledger.LedgerEntry{}}
}

func (r *fakeRepo) mutate(walletID, ownerID, currency string, delta int64, typ, idem, ref string) (ledger.MutationResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if e, ok := r.seen[idem]; ok {
		return ledger.MutationResult{Entry: e, Wallet: r.wallets[walletID], AlreadyApplied: true}, nil
	}
	w := r.wallets[walletID]
	if w.ID == "" {
		w = ledger.Wallet{ID: walletID, OwnerID: ownerID, Currency: currency}
	}
	if delta < 0 && w.Balance < -delta {
		return ledger.MutationResult{}, ledger.ErrInsufficientFunds
	}
	w.Balance += delta
	w.Version++
	w.UpdatedAt = time.Now().UTC()
	r.nextID++
	e := ledger.LedgerEntry{
		ID: "ledger-test", WalletID: walletID, Type: typ, AmountMinor: delta,
		Currency: currency, Reference: ref, IdempotencyKey: idem, CreatedAt: time.Now().UTC(),
	}
	r.wallets[walletID] = w
	r.seen[idem] = e
	return ledger.MutationResult{Entry: e, Wallet: w}, nil
}

func TestLedgerConcurrencyAndIdempotency(t *testing.T) {
	repo := newFakeRepo()
	const initial int64 = 100_000

	repo.wallets["w1"] = ledger.Wallet{ID: "w1", OwnerID: "u1", Currency: "INR", Balance: initial}

	var wg sync.WaitGroup
	var credits, debits int64
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			amount := int64(100 + i)
			if i%2 == 0 {
				if _, err := repo.mutate("w1", "u1", "INR", amount, "CREDIT", "credit-"+itoa(i), "test"); err != nil {
					t.Errorf("credit: %v", err)
				}
				mu.Lock()
				credits += amount
				mu.Unlock()
			} else {
				if _, err := repo.mutate("w1", "u1", "INR", -amount, "DEBIT", "debit-"+itoa(i), "test"); err != nil {
					t.Errorf("debit: %v", err)
				}
				mu.Lock()
				debits += amount
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	expected := initial + credits - debits
	got := repo.wallets["w1"].Balance
	if got != expected {
		t.Fatalf("reconciliation failed: initial=%d credits=%d debits=%d final=%d", initial, credits, debits, got)
	}

	before := repo.wallets["w1"].Balance
	first, err := repo.mutate("w1", "u1", "INR", -500, "DEBIT", "replay-key", "test")
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.mutate("w1", "u1", "INR", -500, "DEBIT", "replay-key", "test")
	if err != nil {
		t.Fatal(err)
	}
	if !second.AlreadyApplied {
		t.Fatal("expected idempotent replay")
	}
	if first.Entry.ID != second.Entry.ID {
		t.Fatal("replay returned different transaction")
	}
	if repo.wallets["w1"].Balance != before-500 {
		t.Fatal("idempotent replay mutated balance twice")
	}
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	b := make([]byte, 0, 12)
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	return string(b)
}

var _ = context.Background
