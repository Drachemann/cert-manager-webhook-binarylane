package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/drachemann/cert-manager-webhook-binarylane/pkg/solver"
)

func main() {
	groupName := os.Getenv("GROUP_NAME")
	if groupName == "" {
		groupName = "acme.binarylane.com"
	}

	healthzPort := os.Getenv("HEALTHZ_PORT")
	if healthzPort == "" {
		healthzPort = "6080"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go serveHealthz(ctx, healthzPort)

	cmd.RunWebhookServer(groupName, &solver.Solver{})
}

func serveHealthz(ctx context.Context, port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "healthz server error: %v\n", err)
	}
}
