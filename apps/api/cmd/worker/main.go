package main

import (
	"log"
	"os"

	"github.com/hibiken/asynq"
	"github.com/tylor/goaipj/internal/config"
	"github.com/tylor/goaipj/internal/worker"
)

func main() {
	cfg, err := config.Load(config.Options{EnvFile: ".env"})
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := log.New(os.Stdout, "worker ", log.LstdFlags|log.LUTC)
	logger.Printf("config summary: %+v", cfg.Summary())
	mux := asynq.NewServeMux()
	worker.NewRegistry().RegisterAll(worker.NewAsynqRegistrar(mux))
	logger.Println("worker bootstrap complete")
}
