package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/google/uuid"

	"github.com/yourusername/api-monetization-platform/internal/ledger"
)

type WalletRepository struct {
	CB *Couchbase
}

func (r *WalletRepository) Get(ctx context.Context, walletID string) (ledger.Wallet, error) {
	var w ledger.Wallet
	res, err := r.CB.Keyspace("wallets").Get(walletID, &gocb.GetOptions{Context: ctx})
	if err != nil {
		return w, err
	}
	return w, res.Content(&w)
}

func (r *WalletRepository) Mutate(ctx context.Context, walletID, ownerID, currency string, delta int64, typ, idem, reference string) (ledger.MutationResult, error) {
	if idem == "" {
		return ledger.MutationResult{}, errors.New("idempotency key required")
	}

	var out ledger.MutationResult
	err := r.CB.Cluster.Transactions().Run(func(ctx *gocb.TransactionAttemptContext) error {
		idemKey := fmt.Sprintf("ledger-idem::wallet::%s", idem)
		var existing ledger.LedgerEntry
		getExisting, err := ctx.Get(r.CB.Keyspace("idempotency"), idemKey)
		if err == nil {
			if err := getExisting.Content(&existing); err != nil {
				return err
			}
			getWallet, werr := ctx.Get(r.CB.Keyspace("wallets"), walletID)
			if werr != nil {
				return werr
			}
			var w ledger.Wallet
			if werr := getWallet.Content(&w); werr != nil {
				return werr
			}
			out = ledger.MutationResult{Entry: existing, Wallet: w, AlreadyApplied: true}
			return nil
		}
		if err != nil && !errors.Is(err, gocb.ErrDocumentNotFound) {
			return err
		}

		var w ledger.Wallet
		getWallet, err := ctx.Get(r.CB.Keyspace("wallets"), walletID)
		if errors.Is(err, gocb.ErrDocumentNotFound) {
			if delta < 0 {
				return ledger.ErrInsufficientFunds
			}
			w = ledger.Wallet{
				ID: walletID, OwnerID: ownerID, Currency: currency,
				Balance: 0, Version: 0, UpdatedAt: time.Now().UTC(),
			}
		} else if err != nil {
			return err
		} else if err := getWallet.Content(&w); err != nil {
			return err
		}

		if currency != "" && w.Currency != "" && w.Currency != currency {
			return fmt.Errorf("currency mismatch: wallet=%s request=%s", w.Currency, currency)
		}
		if w.Currency == "" {
			w.Currency = currency
		}
		if delta < 0 && w.Balance < -delta {
			return ledger.ErrInsufficientFunds
		}

		w.Balance += delta
		w.Version++
		w.UpdatedAt = time.Now().UTC()

		entry := ledger.LedgerEntry{
			ID:             "ledger::" + uuid.NewString(),
			WalletID:       walletID,
			Type:           typ,
			AmountMinor:    delta,
			Currency:       w.Currency,
			Reference:      reference,
			IdempotencyKey: idem,
			CreatedAt:      time.Now().UTC(),
		}

		if err := ctx.Replace(getWallet, w); err != nil {
			if errors.Is(err, gocb.ErrDocumentNotFound) {
				if err := ctx.Insert(r.CB.Keyspace("wallets"), walletID, w); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		if err := ctx.Insert(r.CB.Keyspace("ledger"), entry.ID, entry); err != nil {
			return err
		}
		if err := ctx.Insert(r.CB.Keyspace("idempotency"), idemKey, entry); err != nil {
			return err
		}

		out = ledger.MutationResult{Entry: entry, Wallet: w}
		return nil
	}, nil)

	return out, err
}
