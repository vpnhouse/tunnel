package sentry

import (
	"github.com/getsentry/sentry-go"
)

// Config for embedding into services' configurations
type Config struct {
	DSN string `json:"dsn"`
	Env string `json:"environment"`
}

func ConfigureGlobal(config Config, release string) error {
	if len(config.DSN) == 0 {
		return nil
	}

	cfg := sentry.ClientOptions{
		Dsn:              config.DSN,
		Debug:            false,
		AttachStacktrace: true,
		SampleRate:       1.0,
		Release:          release,
		Environment:      config.Env,
		Integrations: func(integrations []sentry.Integration) []sentry.Integration {
			use := make([]sentry.Integration, 0, len(integrations))
			for _, in := range integrations {
				// exclude the "Modules" integration from defaults
				if in.Name() == "Modules" {
					continue
				}
				use = append(use, in)
			}
			return use
		},
	}

	return sentry.Init(cfg)
}
