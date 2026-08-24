package wallet

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yourusername/api-monetization-platform/internal/ledger"
	walletv1 "github.com/yourusername/api-monetization-platform/proto"
)

type Server struct {
	walletv1.UnimplementedWalletServiceServer
	Ledger *ledger.Service
}

func (s *Server) GetBalance(ctx context.Context, req *walletv1.GetBalanceRequest) (*walletv1.GetBalanceResponse, error) {
	if req.GetWalletId() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id required")
	}
	w, err := s.Ledger.Repo.Get(ctx, req.GetWalletId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "wallet not found")
	}
	return &walletv1.GetBalanceResponse{WalletId: w.ID, BalanceMinor: w.Balance, Currency: w.Currency}, nil
}

func (s *Server) Credit(ctx context.Context, req *walletv1.CreditRequest) (*walletv1.MutationResponse, error) {
	if req.GetWalletId() == "" || req.GetAmountMinor() <= 0 || req.GetIdempotencyKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id, amount and idempotency_key required")
	}
	res, err := s.Ledger.Credit(ctx, req.WalletId, "", req.Currency, req.AmountMinor, req.IdempotencyKey, req.Reference)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mutation(res), nil
}

func (s *Server) Debit(ctx context.Context, req *walletv1.DebitRequest) (*walletv1.MutationResponse, error) {
	if req.GetWalletId() == "" || req.GetAmountMinor() <= 0 || req.GetIdempotencyKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id, amount and idempotency_key required")
	}
	res, err := s.Ledger.Debit(ctx, req.WalletId, "", req.Currency, req.AmountMinor, req.IdempotencyKey, req.Reference)
	if errors.Is(err, ledger.ErrInsufficientFunds) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mutation(res), nil
}

func mutation(res ledger.MutationResult) *walletv1.MutationResponse {
	return &walletv1.MutationResponse{
		TransactionId: res.Entry.ID, WalletId: res.Wallet.ID,
		BalanceMinor: res.Wallet.Balance, Currency: res.Wallet.Currency,
		AlreadyApplied: res.AlreadyApplied,
	}
}
