package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tylor/goaipj/internal/config"
)

func main() {
	cfg, err := config.Load(config.Options{EnvFile: ".env"})
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := log.New(os.Stdout, "api ", log.LstdFlags|log.LUTC)
	logger.Printf("config summary: %+v", cfg.Summary())

	server := newDevelopmentServer(logger)

	addr := fmt.Sprintf(":%s", cfg.APIPort)
	logger.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("serve http: %v", err)
	}
}
