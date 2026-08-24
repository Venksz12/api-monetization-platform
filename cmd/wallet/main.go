package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/yourusername/api-monetization-platform/internal/config"
	"github.com/yourusername/api-monetization-platform/internal/ledger"
	"github.com/yourusername/api-monetization-platform/internal/observability"
	"github.com/yourusername/api-monetization-platform/internal/repository"
	walletv1 "github.com/yourusername/api-monetization-platform/proto"
	walletservice "github.com/yourusername/api-monetization-platform/services/wallet"
)

func main() {
	cfg := config.Load()
	log, _ := zap.NewProduction()
	defer log.Sync()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := observability.InitTracer(ctx, "api-monetization-wallet")
	if err != nil {
		log.Fatal("tracer", zap.Error(err))
	}
	defer shutdownTracer(context.Background())

	cb, err := repository.Connect(ctx, cfg.CBConnStr, cfg.CBUsername, cfg.CBPassword, cfg.CBBucket, cfg.CBScope)
	if err != nil {
		log.Fatal("couchbase", zap.Error(err))
	}
	defer cb.Close()

	lis, err := net.Listen("tcp", cfg.WalletGRPCAddr)
	if err != nil {
		log.Fatal("listen", zap.Error(err))
	}

	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	walletv1.RegisterWalletServiceServer(s, &walletservice.Server{Ledger: &ledger.Service{Repo: &repository.WalletRepository{CB: cb}}})

	go func() {
		log.Info("wallet listening", zap.String("addr", cfg.WalletGRPCAddr))
		if err := s.Serve(lis); err != nil {
			log.Fatal("grpc", zap.Error(err))
		}
	}()

	<-ctx.Done()
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() { s.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-stopCtx.Done():
		s.Stop()
	}
}
