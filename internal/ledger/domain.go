package ledger

import (
	"errors"
	"time"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("amount must be positive")
)

type Wallet struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Currency  string    `json:"currency"`
	Balance   int64     `json:"balance_minor"`
	Version   uint64    `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LedgerEntry struct {
	ID             string    `json:"id"`
	WalletID       string    `json:"wallet_id"`
	Type           string    `json:"type"`
	AmountMinor    int64     `json:"amount_minor"`
	Currency       string    `json:"currency"`
	Reference      string    `json:"reference"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}
