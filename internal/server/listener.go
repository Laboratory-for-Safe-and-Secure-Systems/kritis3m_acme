package server

import (
	"fmt"
	"log"
	"net"
	"os"

	asl "github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl/lib"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
)

type ASLServerConfig struct {
	Logger         *logger.Logger
	Address        string
	EndpointConfig *asl.EndpointConfig
}

func NewASLListener(config *ASLServerConfig) (*lib.ASLListener, error) {
	if config == nil {
		return nil, fmt.Errorf("server config is required")
	}

	var logger *log.Logger
	if config.Logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		logger = log.New(config.Logger, "", 0)
	}

	// Create TCP Listener
	tcpListener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %v", err)
	}

	// Setup ASL Endpoint
	endpoint := asl.ASLsetupServerEndpoint(config.EndpointConfig)
	if endpoint == nil {
		tcpListener.Close()
		return nil, fmt.Errorf("failed to setup ASL endpoint")
	}

	// Create ASL Listener
	aslListener := &lib.ASLListener{
		Logger:      logger,
		TcpListener: tcpListener.(*net.TCPListener),
		Endpoint:    endpoint,
	}

	return aslListener, nil
}
