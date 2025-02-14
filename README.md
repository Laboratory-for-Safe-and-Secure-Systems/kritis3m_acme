# KRITIS3M ACME Server

This is a simple ACME server that implements the ACME protocol as defined in
[RFC 8555](https://tools.ietf.org/html/rfc8555) with custom TLS Layer and
custom PKI.

- TLS Layer: [ASL](https://github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_asl)
- Go ASL Wrapper: [Go-ASL](https://github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl)
- PKI: [PKI](https://github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_pki)

## Features

- [x] ASL support
- [x] Directory endpoint
- [x] Health check endpoint
- [x] Basic CORS support
- [x] Structured logging
- [x] Configuration management
- [x] Graceful shutdown

## Work in Progress

- [ ] Request Authentication
  - [ ] JWS Signature Verification (RFC 8555 Section 6.2)
  - [x] Nonce Handling (RFC 8555 Section 6.5)
- [ ] Account Handling
  - [x] Account creation endpoint
  - [ ] Account update
  - [ ] Key change
- [ ] Order Handling
  - [x] Order creation endpoint
  - [ ] Order retrieval
  - [ ] Order finalization
- [ ] Challenge Handling (Mainly for IP based hosts)
  - [ ] HTTP-01
  - [ ] TLS-ALPN-01 (Maybe)
- [ ] Certificate Issuance
  - [ ] CSR Validation
  - [ ] Certificate Generation
  - [ ] Certificate Revocation

## Configuration

The server can be configured using a JSON configuration file. See `config.json` for an example configuration.

Key configuration options:
- Server settings (port, host)
- ASL configuration
- TLS/Certificate settings
- Logging options

## Building and Running

```bash
# Build the server
go build -o acme-server ./cmd/acme-server

# Run with config file
./acme-server -config config.json

# Enable debug mode
./acme-server -config config.json -debug
```

## License

MIT License - See LICENSE file for details.
