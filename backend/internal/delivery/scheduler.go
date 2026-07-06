package delivery

import (
	"log"
	"time"
)

func StartPeriodicSync(dbPath string, interval time.Duration, logger *log.Logger) {
	if interval <= 0 {
		logger.Printf("delivery periodic sync disabled (interval <= 0)")
		return
	}
	if !CredentialsConfigured() {
		logger.Printf("delivery periodic sync disabled: DELIVERY_SAPIENS_DB_PASSWORD not set")
		return
	}

	logger.Printf("delivery periodic sync enabled: every %s", interval)

	go func() {
		run := func() {
			message, err := RunSync(dbPath)
			if err != nil {
				logger.Printf("delivery sync error: %v", err)
				return
			}
			logger.Printf("delivery sync ok: %s", message)
		}

		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
}
