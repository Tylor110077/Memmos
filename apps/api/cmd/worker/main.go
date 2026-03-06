package main

import (
	"log"
	"os"

	"github.com/tylor/goaipj/internal/config"
)

func main() {
	cfg, err := config.Load(config.Options{EnvFile: ".env"})
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := log.New(os.Stdout, "worker ", log.LstdFlags|log.LUTC)
	logger.Printf("config summary: %+v", cfg.Summary())
	logger.Println("worker bootstrap complete")
}
