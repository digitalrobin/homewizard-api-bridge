package main

import (
	"errors"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	store, err := newTokenStore(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	client := newHomeWizardClient(cfg, store)
	usage, err := newUsageStore(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Addr:              cfg.BindAddr,
		Handler:           newServer(client, usage).routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("homewizard-bridge listening on %s", cfg.BindAddr)
	log.Printf("HomeWizard target: %s", cfg.HomeWizardHost)
	log.Printf("token ready: %t", store.Get() != "")

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
