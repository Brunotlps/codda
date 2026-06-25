package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/Brunotlps/codda/internal/adapters/http"
	"github.com/Brunotlps/codda/internal/adapters/memory"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/config"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	repo := memory.NewOrderRepository()

	createOrder := application.NewCreateOrderUseCase(repo)
	findOrder := application.NewFindOrderByIDUseCase(repo)
	listOrders := application.NewListOrdersUseCase(repo)
	markPaid := application.NewMarkOrderAsPaidUseCase(repo)
	markCancelled := application.NewMarkOrderAsCancelledUseCase(repo)
	markShipped := application.NewMarkOrderAsShippedUseCase(repo)

	handler := http.NewHandler(createOrder, findOrder, listOrders, markPaid, markCancelled, markShipped)
	router := http.NewRouter(handler)
	server := http.NewServer(cfg.HTTPAddr, router)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("server started on %s", cfg.HTTPAddr)
		if err := server.Start(); err != nil {
			log.Printf("server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("received shutdown signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}
