package netx

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"scpi-sim1440/internal/core"
	"syscall"
	"time"
)

func HandleServer(device core.Device) {
	srv := http.Server{
		Addr: ":8080", //todo задавать через конфиг
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("failed to start server: %v", err)
			}
		}
	}()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()

	ctxTimeout, c := context.WithTimeout(ctx, 5*time.Second)
	defer c()

	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Fatalf("failed to shutdown server: %v", err)
	}
}
