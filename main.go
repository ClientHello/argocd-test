package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

func main() {
	logger = log.Logger
	secretsDir, ok := os.LookupEnv("SECRETS_DIR")
	if !ok {
		logger.Fatal().Msg("SECRETS_DIR environment variable is not set")
	}

	if err := filepath.WalkDir(secretsDir, logSecrets); err != nil {
		logger.Fatal().Err(err).Msg("Failed to read filesystem")
	}
}

func logSecrets(path string, d fs.DirEntry, err error) error {
	if d.IsDir() {
		return nil
	}

	childLogger := logger.With().Str("path", path).Logger()
	f, err := os.ReadFile(path)
	if err != nil {
		childLogger.Warn().Err(err).Msg("Failed to read file from disk")
	}

	childLogger.Debug().Str("secret_value", string(f)).Msg("Read secret from disk")
	return nil
}
