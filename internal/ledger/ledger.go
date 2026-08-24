package ledger

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type Entry struct {
	ID             string    `json:"id"`
	WalletID       string    `json:"wallet_id"`
	Type           string    `json:"type"`
	AmountMinor    int64     `json:"amount_minor"`
	Currency       string    `json:"currency"`
	Reference      string    `json:"reference"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type Wallet struct {
	ID         string `json:"id"`
	Balance    int64  `json:"balance_minor"`
	Currency   string `json:"currency"`
}

type Service struct {
	mu      sync.Mutex
	wallets map[string]*Wallet
	entries map[string]Entry
}

func New() *Service {
	return &Service{
		wallets: map[string]*Wallet{},
		entries: map[string]Entry{},
	}
}

func (s *Service) Get(walletID string) Wallet {
	s.mu.Lock()
	defer s.mu.Unlock()
	w := s.wallets[walletID]
	if w == nil {
		w = &Wallet{ID: walletID, Currency: "INR"}
		s.wallets[walletID] = w
	}
	return *w
}

func (s *Service) Credit(walletID string, amount int64, currency, key, reference string) (Entry, Wallet, bool, error) {
	return s.mutate(walletID, amount, currency, key, reference, "CREDIT")
}

func (s *Service) Debit(walletID string, amount int64, currency, key, reference string) (Entry, Wallet, bool, error) {
	return s.mutate(walletID, -amount, currency, key, reference, "DEBIT")
}

func (s *Service) mutate(walletID string, delta int64, currency, key, reference, typ string) (Entry, Wallet, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.entries[key]; ok {
		return e, *s.wallets[walletID], true, nil
	}
	if delta <= 0 && typ == "DEBIT" && s.wallets[walletID] != nil && s.wallets[walletID].Balance < -delta {
		return Entry{}, Wallet{}, false, ErrInsufficientFunds
	}

	w := s.wallets[walletID]
	if w == nil {
		w = &Wallet{ID: walletID, Currency: currency}
		s.wallets[walletID] = w
	}
	if currency != "" {
		w.Currency = currency
	}
	w.Balance += delta

	e := Entry{
		ID: uuid.NewString(), WalletID: walletID, Type: typ,
		AmountMinor: delta, Currency: w.Currency, Reference: reference,
		IdempotencyKey: key, CreatedAt: time.Now().UTC(),
	}
	s.entries[key] = e
	return e, *w, false, nil
}
