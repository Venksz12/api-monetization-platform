package ledger

import (
	"context"
	"errors"

	"github.com/yourusername/api-monetization-platform/internal/repository"
)

type Service struct {
	Repo *repository.WalletRepository
}

type MutationResult struct {
	Entry          LedgerEntry
	Wallet         Wallet
	AlreadyApplied bool
}

func (s *Service) Credit(ctx context.Context, walletID, ownerID, currency string, amount int64, idem, reference string) (MutationResult, error) {
	if amount <= 0 {
		return MutationResult{}, ErrInvalidAmount
	}
	return s.Repo.Mutate(ctx, walletID, ownerID, currency, amount, "CREDIT", idem, reference)
}

func (s *Service) Debit(ctx context.Context, walletID, ownerID, currency string, amount int64, idem, reference string) (MutationResult, error) {
	if amount <= 0 {
		return MutationResult{}, ErrInvalidAmount
	}
	return s.Repo.Mutate(ctx, walletID, ownerID, currency, -amount, "DEBIT", idem, reference)
}

func IsInsufficient(err error) bool {
	return errors.Is(err, ErrInsufficientFunds)
}
