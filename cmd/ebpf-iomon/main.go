package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mansam/ebpf-iomon/pkg/config"
	"github.com/mansam/ebpf-iomon/pkg/device"
	bpf "github.com/mansam/ebpf-iomon/pkg/ebpf"
	"github.com/mansam/ebpf-iomon/pkg/metrics"
)

func main() {
	cfg := config.Parse()
	log := setupLogger(cfg.LogLevel)

	if err := cfg.Validate(); err != nil {
		log.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	log.Info("starting ebpf-iomon",
		"node", cfg.NodeName,
		"block", cfg.EnableBlock,
		"nfs", cfg.EnableNFS,
		"nfsKprobe", cfg.EnableNFSKprobe,
		"metricsPort", cfg.MetricsPort,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	resolver := device.NewResolver(
		cfg.NodeName,
		cfg.ProcPath,
		time.Duration(cfg.ScanInterval)*time.Second,
		log,
	)
	go resolver.Run(ctx)

	programs, err := bpf.LoadAndAttach(cfg.EnableBlock, cfg.EnableNFS, cfg.EnableNFSKprobe, cfg.BlockMapSize, cfg.NFSMapSize, cfg.NFSKprobeMapSize, log)
	if err != nil {
		log.Error("loading eBPF programs", "error", err)
		os.Exit(1)
	}
	defer programs.Close()

	collector := metrics.NewCollector(programs.BlockHists, programs.NfsHists, programs.NfsKprobeHists, programs.BlockActive, programs.NFSActive, programs.NFSKprobeActive, resolver, cfg.NodeName, log)
	prometheus.MustRegister(collector)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("metrics server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case sig := <-sigCh:
		log.Info("received signal, shutting down", "signal", sig)
	case err := <-errCh:
		log.Error("metrics server error, shutting down", "error", err)
	}

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
}

func setupLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
