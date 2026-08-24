package main

import (
	"context"
	"errors"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	walletv1 "github.com/yourusername/api-monetization-platform/proto"
	"github.com/yourusername/api-monetization-platform/internal/config"
	"github.com/yourusername/api-monetization-platform/internal/ledger"
	"github.com/yourusername/api-monetization-platform/internal/observability"
)

type server struct {
	walletv1.UnimplementedWalletServiceServer
	ledger *ledger.Service
}

func main() {
	cfg := config.Load()
	log := observability.NewLogger()
	defer log.Sync()

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("listen", zap.Error(err))
	}

	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(4<<20),
		grpc.MaxSendMsgSize(4<<20),
	)
	walletv1.RegisterWalletServiceServer(s, &server{ledger: ledger.New()})

	log.Info("wallet grpc started", zap.String("addr", cfg.GRPCAddr))
	if err := s.Serve(lis); err != nil {
		log.Fatal("grpc serve", zap.Error(err))
	}
}

func (s *server) GetBalance(_ context.Context, in *walletv1.GetBalanceRequest) (*walletv1.GetBalanceResponse, error) {
	if in.GetWalletId() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id required")
	}
	w := s.ledger.Get(in.GetWalletId())
	return &walletv1.GetBalanceResponse{
		WalletId: w.ID, BalanceMinor: w.Balance, Currency: w.Currency,
	}, nil
}

func (s *server) Credit(_ context.Context, in *walletv1.CreditRequest) (*walletv1.MutationResponse, error) {
	if in.GetWalletId() == "" || in.GetAmountMinor() <= 0 || in.GetIdempotencyKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id, positive amount and idempotency_key required")
	}
	e, w, duplicate, err := s.ledger.Credit(
		in.WalletId, in.AmountMinor, in.Currency, in.IdempotencyKey, in.Reference,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.MutationResponse{
		TransactionId: e.ID, WalletId: w.ID, BalanceMinor: w.Balance,
		Currency: w.Currency, AlreadyApplied: duplicate,
	}, nil
}

func (s *server) Debit(_ context.Context, in *walletv1.DebitRequest) (*walletv1.MutationResponse, error) {
	if in.GetWalletId() == "" || in.GetAmountMinor() <= 0 || in.GetIdempotencyKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id, positive amount and idempotency_key required")
	}
	e, w, duplicate, err := s.ledger.Debit(
		in.WalletId, in.AmountMinor, in.Currency, in.IdempotencyKey, in.Reference,
	)
	if errors.Is(err, ledger.ErrInsufficientFunds) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &walletv1.MutationResponse{
		TransactionId: e.ID, WalletId: w.ID, BalanceMinor: w.Balance,
		Currency: w.Currency, AlreadyApplied: duplicate,
	}, nil
}
