package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)


type Config struct {
	REST_Port string 
	GRPC_Port string
	DSN string
}

func LoadConfig(cfg *Config) error {
	// Get the path of the current file (config.go)
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Load .env from the same directory as config.go
	envPath := filepath.Join(dir, ".env")

	err := godotenv.Load(envPath)
	if err != nil {
		return err
	}

	// ACCESS ENV varialbes

	dsn := os.Getenv("DSN")
	if dsn == "" {
		return errors.New("could not find DSN in env variables")
	} 


	// Initiliaze the config struct 
	cfg.DSN = dsn
	cfg.REST_Port = os.Getenv("REST_PORT")
	cfg.GRPC_Port = os.Getenv("GRPC_PORT")


	return nil
}