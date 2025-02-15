package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	asl "github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/router"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/config"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/database"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/server"
)

func initDatabase(ctx context.Context, cfg *config.Config) (*database.DB, error) {
	log := logger.GetLogger(ctx)

	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	log.Info("Database connection established")
	return db, nil
}

func main() {
	// Initialize global logger
	log := logger.New(os.Stdout)
	// Add to context
	ctx := context.WithValue(context.Background(), types.CtxKeyLogger, log)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Errorf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Initialize database if config is provided
	var db *database.DB
	if config.GlobalFlags.DBConfig != "" {
		dbCfg, err := config.Load(config.GlobalFlags.DBConfig, &config.Config{})
		if err != nil {
			log.Errorf("Failed to load database configuration: %v", err)
			os.Exit(1)
		}
		cfg.Database = dbCfg.Database
	}

	if cfg.Database.Host != "" {
		db, err = initDatabase(ctx, cfg)
		if err != nil {
			log.Errorf("Failed to initialize database: %v", err)
			os.Exit(1)
		}
		defer db.Close()
	}

	r := router.New(ctx, db)

	libConfig := &asl.ASLConfig{
		LoggingEnabled: cfg.ASLConfig.LoggingEnabled,
		LogLevel:       int32(cfg.ASLConfig.LogLevel),
	}
	err = asl.ASLinit(libConfig)
	if err != nil {
		log.Errorf("Error initializing ASL: %v", err)
	}

	// Configure ASL endpoint
	endpointConfig := &asl.EndpointConfig{
		MutualAuthentication: cfg.Endpoint.MutualAuthentication,
		NoEncryption:         cfg.Endpoint.NoEncryption,
		ASLKeyExchangeMethod: asl.ASLKeyExchangeMethod(cfg.Endpoint.ASLKeyExchangeMethod),
		PreSharedKey: asl.PreSharedKey{
			Enable: false,
		},
		DeviceCertificateChain: asl.DeviceCertificateChain{Path: cfg.TLS.Certs},
		PrivateKey: asl.PrivateKey{
			Path: cfg.TLS.PrivateKey,
		},
		RootCertificate: asl.RootCertificate{Path: cfg.CA.Certs},
		KeylogFile:      cfg.Endpoint.KeylogFile,
		PKCS11: asl.PKCS11ASL{
			Path: cfg.PKCS11.EntityModule.Path,
			Pin:  cfg.PKCS11.EntityModule.Pin,
		},
	}

	// Create server configuration
	serverConfig := &server.ASLServerConfig{
		Logger:         log,
		Address:        cfg.Server.ListenAddr,
		EndpointConfig: endpointConfig,
	}

	// Create and start server
	srv, err := server.ASLServer(serverConfig, r, ctx)
	if err != nil {
		log.Errorf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServeASL(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Failed to start server: %v", err)
		}
	}()

	log.Infof("Server started on %s", serverConfig.Address)

	// Wait for interrupt signal
	<-done
	log.Info("Server stopping...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server shutdown failed: %v", err)
	}

	log.Info("Server stopped gracefully")
}
