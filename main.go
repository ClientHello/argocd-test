package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	_httpPortEnvVar   = "HTTP_PORT"
	_secretsDirEnvVar = "SECRETS_DIR"
)

var (
	logger  zerolog.Logger
	secrets []secret
)

type secretsDTO struct {
	Secrets []secret `json:"secrets"`
}

type secret struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

func main() {
	logger = log.Logger
	httpPortStr := lookupEnv(_httpPortEnvVar)
	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("http_port", httpPortStr).
			Msg("HTTP Port set in environment variable is not a number")
	}

	secretsDir := lookupEnv(_secretsDirEnvVar)

	if err := filepath.WalkDir(secretsDir, logSecrets); err != nil {
		logger.Fatal().Err(err).Msg("Failed to read filesystem")
	}

	http.HandleFunc("/secret", secretHandler)

	listenAddr := fmt.Sprintf(":%d", httpPort)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Warn().
			Err(err).
			Msg("HTTP server stopped")
	}
}

func lookupEnv(envVarName string) string {
	envVarVal, ok := os.LookupEnv(envVarName)
	if !ok {
		logger.Fatal().
			Str("env_var_name", envVarName).
			Msg("Environment variable is not set")
	}

	return envVarVal
}

func logSecrets(path string, d fs.DirEntry, err error) error {
	if d == nil {
		return fmt.Errorf("path doesn't exist or can't be read: %s", path)
	}

	if d.IsDir() {
		return nil
	}

	childLogger := logger.With().Str("path", path).Logger()
	secretVal, err := os.ReadFile(path)
	if err != nil {
		childLogger.Warn().Err(err).Msg("Failed to read file from disk")
	}

	secretValStr := string(secretVal)
	childLogger.Debug().Str("secret_value", secretValStr).Msg("Read secret from disk")
	s := secret{
		Path:  path,
		Value: secretValStr,
	}

	secrets = append(secrets, s)

	return nil
}

func secretHandler(resp http.ResponseWriter, req *http.Request) {
	s := secretsDTO{
		Secrets: secrets,
	}

	sJSON, err := json.Marshal(s)
	if err != nil {
		logger.Error().
			Err(err).
			Any("secrets", s).
			Msg("Failed to marshal secrets as JSON")
	}

	if numBytesWritten, err := resp.Write(sJSON); err != nil {
		logger.Error().
			Err(err).
			Str("secrets_json", string(sJSON)).
			Int("num_response_bytes_written", numBytesWritten).
			Msg("Failed to write secrets JSON to HTTP response")
	}
}
