//go:build worker

// Command whatsapp-worker consumes the `whatsapp.inbound` Kafka topic and runs
// the WhatsApp orchestration pipeline (identity → parse → domain services →
// reply). It shares the expense-service dependency graph with the HTTP server but
// runs as its own process. Build with: go build -tags worker -o bin/whatsapp-worker ./cmd
//
// It is a separate main via the `worker` build tag so the cmd package keeps a
// single binary per build without colliding main() symbols.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kharchibook/expense-service/config"
	"github.com/kharchibook/expense-service/pkg/di"
	"github.com/kharchibook/expense-service/pkg/infrastructure/msgqueuerepo"
	"github.com/kharchibook/expense-service/third_party/platlogger"
)

func main() {
	if err := runWorker(); err != nil {
		platlogger.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func runWorker() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logLevel := "info"
	if cfg.App.Env == "dev" {
		logLevel = "debug"
	}
	platlogger.Init(cfg.App.Name+"-whatsapp-worker", logLevel)
	platlogger.Info("starting whatsapp-worker", "env", cfg.App.Env)

	if !cfg.MsgQueue.Enabled {
		return fmt.Errorf("MsgQueue.Enabled is false — the worker needs Kafka; set MSGQUEUE_ENABLED=true")
	}

	app, err := di.InitializeApp(cfg)
	if err != nil {
		return fmt.Errorf("initialize app: %w", err)
	}
	defer func() { _ = app.Close() }()

	consumer, err := msgqueuerepo.NewKafkaConsumer(
		cfg.MsgQueue.Brokers, cfg.MsgQueue.ConsumerGroup, cfg.MsgQueue.InboundTopic,
	)
	if err != nil {
		return fmt.Errorf("init kafka consumer: %w", err)
	}
	defer func() { _ = consumer.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the consume loop; stop on SIGINT/SIGTERM by cancelling ctx.
	runErr := make(chan error, 1)
	go func() {
		platlogger.Info("whatsapp-worker consuming", "topic", cfg.MsgQueue.InboundTopic)
		runErr <- consumer.Run(ctx, app.WhatsAppService().Handle)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-runErr:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("consumer: %w", err)
		}
	case sig := <-stop:
		platlogger.Info("shutdown signal received", "signal", sig.String())
		cancel()
		<-runErr // wait for the loop to unwind
	}
	platlogger.Info("whatsapp-worker stopped cleanly")
	return nil
}
