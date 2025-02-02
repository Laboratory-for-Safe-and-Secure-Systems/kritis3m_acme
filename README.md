# KRITIS3M ACME Server

This is a simple ACME server that implements the ACME protocol as defined in
[RFC 8555](https://tools.ietf.org/html/rfc8555) with custom TLS Layer and
custom PKI.

- TLS Layer: [ASL](https://github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_asl)
- Go ASL Wrapper: [Go-ASL](https://github.com/Laboratory-for-Safe-and-Secure-Systems/go-asl)
- PKI: [PKI](https://github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_pki)

## Future Work

- [ ] Add ASL support
- [ ] Request Authentication
  - [ ] JWS Signature Verification (RFC 8555 Section 6.2)
  - [ ] Nonce Handling (RFC 8555 Section 6.5) in successful and error responses
- [ ] Account Handling
- [ ] Order Handling
- [ ] Challenge Handling (Mainly for IP based hosts)
  - [ ] HTTP-01
  - [ ] TLS-ALPN-01 (Maybe)
- [ ] Certificate Issuance
  - [ ] Validation
  - [ ] Issuance
