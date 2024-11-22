package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/playmixer/medal-of-discipline/internal/adapters/api/rest"
	"github.com/playmixer/medal-of-discipline/internal/adapters/session/cookie"
	"github.com/playmixer/medal-of-discipline/internal/adapters/storage/database"
	"github.com/playmixer/medal-of-discipline/internal/adapters/ui/web"
	"github.com/playmixer/medal-of-discipline/internal/core/config"
	"github.com/playmixer/medal-of-discipline/internal/core/discipline"
	"github.com/playmixer/medal-of-discipline/internal/logger"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()
	cfg, err := config.Init()
	if err != nil {
		return fmt.Errorf("failed initialize config: %w", err)
	}

	lgr, err := logger.New(logger.SetLevel(cfg.LogLevel))
	if err != nil {
		return fmt.Errorf("failed inittialize logger: %w", err)
	}

	store, err := database.New(ctx, cfg.Store.DSN)
	if err != nil {
		return fmt.Errorf("failed inittialize storage: %w", err)
	}

	disc, err := discipline.New(ctx, store, discipline.SetLogger(lgr))
	if err != nil {
		return fmt.Errorf("failed initialize discipline: %w", err)
	}

	ui, err := web.New()
	if err != nil {
		return fmt.Errorf("failed initialize web interface: %w", err)
	}

	sess, err := cookie.New([]byte(cfg.Rest.SecretKey))
	if err != nil {
		return fmt.Errorf("failed inititalize session store: %w", err)
	}

	srv := rest.New(
		disc,
		ui,
		sess,
		rest.SetAddress(cfg.Rest.Address),
		rest.SetLogger(lgr),
		rest.SetSecretKey([]byte(cfg.Rest.SecretKey)),
	)
	go func() {
		if err := srv.Run(); err != nil {
			lgr.Error("stop run server", zap.Error(err))
		}
	}()
	<-ctx.Done()
	lgr.Info("Stopping...")
	ctx, stop := context.WithTimeout(context.Background(), time.Second*2)
	defer stop()

	<-ctx.Done()
	lgr.Info("Server stoped")

	return nil
}
