package server

import (
	"fmt"
	"net"

	asl "github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl/listener"
	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/logger"
)

type ASLServerConfig struct {
	Logger         *logger.Logger
	Address        string
	EndpointConfig *asl.EndpointConfig
}

func NewASLListener(config *ASLServerConfig) (*listener.ASLListener, error) {
	if config == nil {
		return nil, fmt.Errorf("server config is required")
	}

	// Create TCP Listener
	netListener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %v", err)
	}

	// Setup ASL Endpoint
	endpoint := asl.ASLsetupServerEndpoint(config.EndpointConfig)
	if endpoint == nil {
		netListener.Close()
		return nil, fmt.Errorf("failed to setup ASL endpoint")
	}

	// Create ASL Listener
	aslListener := &listener.ASLListener{
		Logger:   config.Logger,
		Endpoint: endpoint,
		Listener: netListener,
	}

	return aslListener, nil
}
