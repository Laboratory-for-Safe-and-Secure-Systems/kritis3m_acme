package database

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
)

// Account-related queries
const (
	createAccountQuery = `
		INSERT INTO accounts (id, key, contact, status, terms_agreed, created_at, initial_ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	getAccountQuery = `
		SELECT id, key, contact, status, terms_agreed, created_at, initial_ip
		FROM accounts
		WHERE id = $1
		AND status != 'deactivated'`

	updateAccountQuery = `
		UPDATE accounts
		SET contact = $2, status = $3
		WHERE id = $1
		RETURNING id`
)

// CreateAccount creates a new account in the database
func (db *DB) CreateAccount(ctx context.Context, account *types.Account) error {
	keyJSON, err := json.Marshal(account.Key)
	if err != nil {
		return fmt.Errorf("error marshaling account key: %w", err)
	}

	contactJSON, err := json.Marshal(account.Contact)
	if err != nil {
		return fmt.Errorf("error marshaling contact info: %w", err)
	}

	return db.Transaction(ctx, func(tx *sql.Tx) error {
		var id string
		err := tx.QueryRowContext(ctx, createAccountQuery,
			account.ID,
			keyJSON,
			contactJSON,
			account.Status,
			account.TermsOfServiceAgreed,
			account.CreatedAt,
			account.InitialIP,
		).Scan(&id)

		if err != nil {
			return fmt.Errorf("error creating account: %w", err)
		}

		return nil
	})
}

// GetAccount retrieves an account from the database
func (db *DB) GetAccount(ctx context.Context, id string) (*types.Account, error) {
	var account types.Account
	var keyJSON, contactJSON []byte

	err := db.QueryRowContext(ctx, getAccountQuery, id).Scan(
		&account.ID,
		&keyJSON,
		&contactJSON,
		&account.Status,
		&account.TermsOfServiceAgreed,
		&account.CreatedAt,
		&account.InitialIP,
	)

	if err == sql.ErrNoRows {
		return nil, &types.Problem{
			Type:   "urn:ietf:params:acme:error:accountDoesNotExist",
			Detail: fmt.Sprintf("account %s does not exist", id),
			Status: http.StatusNotFound,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error querying account: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(keyJSON, &account.Key); err != nil {
		return nil, fmt.Errorf("error unmarshaling account key: %w", err)
	}
	if err := json.Unmarshal(contactJSON, &account.Contact); err != nil {
		return nil, fmt.Errorf("error unmarshaling contact info: %w", err)
	}

	return &account, nil
}

// UpdateAccount updates an existing account in the database
func (db *DB) UpdateAccount(ctx context.Context, account *types.Account) error {
	contactJSON, err := json.Marshal(account.Contact)
	if err != nil {
		return fmt.Errorf("error marshaling contact info: %w", err)
	}

	return db.Transaction(ctx, func(tx *sql.Tx) error {
		var id string
		err := tx.QueryRowContext(ctx, updateAccountQuery,
			account.ID,
			contactJSON,
			account.Status,
		).Scan(&id)

		if err == sql.ErrNoRows {
			return fmt.Errorf("account not found: %s", account.ID)
		}
		if err != nil {
			return fmt.Errorf("error updating account: %w", err)
		}

		return nil
	})
}

// CreateOrder creates a new order and its authorizations in the database
func (db *DB) CreateOrder(ctx context.Context, order *types.Order, authzs []*types.Authorization) error {
	return db.Transaction(ctx, func(tx *sql.Tx) error {
		// Marshal identifiers to JSON
		identifiersJSON, err := json.Marshal(order.Identifiers)
		if err != nil {
			return fmt.Errorf("error marshaling identifiers: %w", err)
		}

		// Create order
		var orderID string
		now := time.Now()
		err = tx.QueryRowContext(ctx, createOrderQuery,
			order.ID,
			order.AccountID,
			order.Status,
			order.ExpiresAt.Time,
			order.NotBefore.Time,
			order.NotAfter.Time,
			identifiersJSON,
			order.Finalize,
			now,
		).Scan(&orderID)

		if err != nil {
			return fmt.Errorf("error creating order: %w", err)
		}

		// Create authorizations using the same transaction
		for _, authz := range authzs {
			authz.OrderID = orderID // Set the order ID for the authorization
			if err := createAuthorizationTx(ctx, tx, authz); err != nil {
				return fmt.Errorf("error creating authorization: %w", err)
			}

			// For each authorization, create its challenges
			// Create HTTP-01 challenge
			httpChallenge := &types.Challenge{
				Type:            "http-01",
				Status:          types.ChallengeStatusPending,
				Token:           generateToken(),
				AuthorizationID: authz.ID,
			}

			// Create TLS-ALPN-01 challenge
			tlsChallenge := &types.Challenge{
				Type:            "tls-alpn-01",
				Status:          types.ChallengeStatusPending,
				Token:           generateToken(),
				AuthorizationID: authz.ID,
			}

			// Insert challenges
			if err := insertChallenge(ctx, tx, httpChallenge); err != nil {
				return fmt.Errorf("failed to create http challenge: %w", err)
			}
			if err := insertChallenge(ctx, tx, tlsChallenge); err != nil {
				return fmt.Errorf("failed to create tls challenge: %w", err)
			}
		}

		return nil
	})
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func insertChallenge(ctx context.Context, tx *sql.Tx, challenge *types.Challenge) error {
	query := `
		INSERT INTO challenges (id, authorization_id, type, status, token, url)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.ExecContext(ctx, query,
		fmt.Sprintf("chall_%d", time.Now().UnixNano()),
		challenge.AuthorizationID,
		challenge.Type,
		challenge.Status,
		challenge.Token,
		challenge.URL,
	)
	return err
}

const (
	createAuthorizationQuery = `
		INSERT INTO authorizations (id, order_id, status, expires_at, identifier, wildcard)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	createOrderQuery = `
		INSERT INTO orders (
			id, account_id, status, expires_at, not_before, not_after, 
			identifiers, finalize, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id`

	getOrderQuery = `
		SELECT id, account_id, status, expires_at, not_before, not_after,
			   identifiers, finalize, certificate_id, created_at, updated_at
		FROM orders
		WHERE id = $1`

	updateOrderQuery = `
		UPDATE orders
		SET status = $2, certificate_id = $3, updated_at = $4
		WHERE id = $1
		RETURNING id`
)

// CreateAuthorization stores an authorization in the database
func (db *DB) CreateAuthorization(ctx context.Context, authz *types.Authorization) error {
	identifierJSON, err := json.Marshal(authz.Identifier)
	if err != nil {
		return fmt.Errorf("error marshaling identifier: %w", err)
	}

	return db.Transaction(ctx, func(tx *sql.Tx) error {
		var id string
		err := tx.QueryRowContext(ctx, createAuthorizationQuery,
			authz.ID,
			authz.OrderID, // Ensure OrderID is part of Authorization struct
			authz.Status,
			authz.Expires.Time,
			identifierJSON,
			authz.Wildcard,
		).Scan(&id)

		if err != nil {
			return fmt.Errorf("error creating authorization: %w", err)
		}

		return nil
	})
}

// createAuthorizationTx inserts an authorization using the provided transaction.
func createAuthorizationTx(ctx context.Context, tx *sql.Tx, authz *types.Authorization) error {
	identifierJSON, err := json.Marshal(authz.Identifier)
	if err != nil {
		return fmt.Errorf("error marshaling identifier: %w", err)
	}

	var id string
	err = tx.QueryRowContext(ctx, createAuthorizationQuery,
		authz.ID,
		authz.OrderID,
		authz.Status,
		authz.Expires.Time,
		identifierJSON,
		authz.Wildcard,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("error creating authorization: %w", err)
	}

	return nil
}

func (db *DB) GetOrder(ctx context.Context, id string) (*types.Order, error) {
	var order types.Order
	var identifiersJSON []byte
	var certificateID sql.NullString
	var notBefore, notAfter sql.NullTime

	// First get the order details
	err := db.QueryRowContext(ctx, getOrderQuery, id).Scan(
		&order.ID,
		&order.AccountID,
		&order.Status,
		&order.ExpiresAt.Time,
		&notBefore.Time,
		&notAfter.Time,
		&identifiersJSON,
		&order.Finalize,
		&certificateID,
		&order.CreatedAt.Time,
		&order.UpdatedAt.Time,
	)

	if err == sql.ErrNoRows {
		return nil, &types.Problem{
			Type:   "urn:ietf:params:acme:error:orderDoesNotExist",
			Detail: fmt.Sprintf("order %s does not exist", id),
			Status: http.StatusNotFound,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error querying order: %w", err)
	}

	// Get the authorizations for this order
	authzQuery := `
		SELECT id FROM authorizations 
		WHERE order_id = $1
	`
	rows, err := db.QueryContext(ctx, authzQuery, order.ID)
	if err != nil {
		return nil, fmt.Errorf("error querying authorizations: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice for authorizations
	order.Authorizations = make([]string, 0)

	// Build the authorization URLs
	baseURL := "https://localhost:8443" // TODO: Get this from config
	for rows.Next() {
		var authzID string
		if err := rows.Scan(&authzID); err != nil {
			return nil, fmt.Errorf("error scanning authorization: %w", err)
		}
		authzURL := fmt.Sprintf("%s/authz/%s", baseURL, authzID)
		order.Authorizations = append(order.Authorizations, authzURL)
	}

	// Rest of the existing code...
	if err := json.Unmarshal(identifiersJSON, &order.Identifiers); err != nil {
		return nil, fmt.Errorf("error unmarshaling identifiers: %w", err)
	}

	if notBefore.Valid {
		order.NotBefore = types.Time{Time: notBefore.Time}
	}
	if notAfter.Valid {
		order.NotAfter = types.Time{Time: notAfter.Time}
	}
	if certificateID.Valid {
		order.CertificateID = certificateID.String
	}

	return &order, nil
}

func (db *DB) UpdateOrder(ctx context.Context, order *types.Order) error {
	return db.Transaction(ctx, func(tx *sql.Tx) error {
		var id string
		err := tx.QueryRowContext(ctx, updateOrderQuery,
			order.ID,
			order.Status,
			order.CertificateID,
			order.UpdatedAt.Time,
		).Scan(&id)

		if err == sql.ErrNoRows {
			return &types.Problem{
				Type:   "urn:ietf:params:acme:error:orderDoesNotExist",
				Detail: fmt.Sprintf("order %s does not exist", order.ID),
				Status: http.StatusNotFound,
			}
		}
		if err != nil {
			return fmt.Errorf("error updating order: %w", err)
		}

		return nil
	})
}

// GetAuthorization retrieves an authorization (and its associated challenges)
// from the database by its ID.
func (db *DB) GetAuthorization(ctx context.Context, id string) (*types.Authorization, error) {
	query := `
        SELECT id, order_id, status, expires_at, identifier, wildcard, created_at, updated_at
        FROM authorizations
        WHERE id = $1
    `
	row := db.QueryRowContext(ctx, query, id)
	var authz types.Authorization
	var identifierJSON []byte
	var expiresAt time.Time
	if err := row.Scan(&authz.ID, &authz.OrderID, &authz.Status, &expiresAt, &identifierJSON, &authz.Wildcard, &authz.CreatedAt, &authz.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("authorization not found")
		}
		return nil, fmt.Errorf("error querying authorization: %w", err)
	}
	authz.Expires = &types.Time{Time: expiresAt}
	if err := json.Unmarshal(identifierJSON, &authz.Identifier); err != nil {
		return nil, fmt.Errorf("error parsing identifier: %w", err)
	}

	// Retrieve associated challenges.
	challengesQuery := `
        SELECT id, type, status, token
        FROM challenges
        WHERE authorization_id = $1
    `
	rows, err := db.QueryContext(ctx, challengesQuery, authz.ID)
	if err != nil {
		return nil, fmt.Errorf("error querying challenges: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var chID, chType, chStatus, token string
		if err := rows.Scan(&chID, &chType, &chStatus, &token); err != nil {
			continue
		}
		challenge := types.Challenge{
			Type:   chType,
			Status: types.ChallengeStatus(chStatus),
			Token:  token,
			URL:    "", // The handler will set the URL based on the request.
		}
		authz.Challenges = append(authz.Challenges, challenge)
	}
	return &authz, nil
}

// GetChallenge retrieves a challenge from the database by its ID.
func (db *DB) GetChallenge(ctx context.Context, id string) (*types.Challenge, error) {
	query := `
        SELECT id, authorization_id, type, url, status, token, validated
        FROM challenges 
        WHERE token = $1
    `
	var c types.Challenge
	var validated sql.NullTime
	if err := db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.AuthorizationID,
		&c.Type,
		&c.URL,
		&c.Status,
		&c.Token,
		&validated,
	); err != nil {
		return nil, fmt.Errorf("error getting challenge: %w", err)
	}
	if validated.Valid {
		c.Validated = &types.Time{Time: validated.Time}
	}
	return &c, nil
}

// UpdateChallengeStatus updates the status of a challenge in the database.
func (db *DB) UpdateChallengeStatus(ctx context.Context, id string, status string) error {
	query := `
        UPDATE challenges
        SET status = $1, updated_at = CURRENT_TIMESTAMP
        WHERE token = $2
    `
	res, err := db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("error updating challenge status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error fetching rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("challenge not found")
	}
	return nil
}

// GetAuthorizationsByOrder retrieves all authorizations associated with a given order ID.
func (db *DB) GetAuthorizationsByOrder(ctx context.Context, orderID string) ([]*types.Authorization, error) {
	query := `
        SELECT id, order_id, status, expires_at, identifier, wildcard, created_at, updated_at
        FROM authorizations
        WHERE order_id = $1
    `
	rows, err := db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("error querying authorizations: %w", err)
	}
	defer rows.Close()

	var authzs []*types.Authorization
	for rows.Next() {
		var authz types.Authorization
		var expiresAt time.Time
		var identifierJSON []byte
		if err := rows.Scan(&authz.ID, &authz.OrderID, &authz.Status, &expiresAt, &identifierJSON, &authz.Wildcard, &authz.CreatedAt, &authz.UpdatedAt); err != nil {
			continue
		}
		authz.Expires = &types.Time{Time: expiresAt}
		if err := json.Unmarshal(identifierJSON, &authz.Identifier); err != nil {
			continue
		}
		authzs = append(authzs, &authz)
	}
	return authzs, nil
}

func (db *DB) GetChallengesByAuthorization(ctx context.Context, authzID string) ([]types.Challenge, error) {
	query := `
        SELECT id, type, url, status, token, validated
        FROM challenges
        WHERE authorization_id = $1
    `
	rows, err := db.QueryContext(ctx, query, authzID)
	if err != nil {
		return nil, fmt.Errorf("error querying challenges: %w", err)
	}
	defer rows.Close()

	var challenges []types.Challenge
	for rows.Next() {
		var c types.Challenge
		var validated sql.NullTime
		if err := rows.Scan(&c.ID, &c.Type, &c.URL, &c.Status, &c.Token, &validated); err != nil {
			return nil, fmt.Errorf("error scanning challenge: %w", err)
		}
		if validated.Valid {
			c.Validated = &types.Time{Time: validated.Time}
		}
		challenges = append(challenges, c)
	}
	return challenges, nil
}

func (db *DB) UpdateAuthorizationStatus(ctx context.Context, authzID string, status string) error {
	query := `
        UPDATE authorizations 
        SET status = $2, updated_at = NOW()
        WHERE id = $1
        RETURNING id
    `
	var id string
	err := db.QueryRowContext(ctx, query, authzID, status).Scan(&id)
	if err != nil {
		return fmt.Errorf("error updating authorization status: %w", err)
	}
	return nil
}

func (db *DB) GetCertificate(ctx context.Context, id string) (*types.Certificate, error) {
	query := `
		SELECT id, order_id, certificate, chain, revoked, revocation_reason, revoked_at
		FROM certificates
		WHERE id = $1
	`
	var cert types.Certificate
	if err := db.QueryRowContext(ctx, query, id).Scan(
		&cert.ID,
		&cert.OrderID,
		&cert.Certificate,
		&cert.Revoked,
		&cert.RevocationReason,
		&cert.RevokedAt,
	); err != nil {
		return nil, fmt.Errorf("error getting certificate: %w", err)
	}
	return &cert, nil
}

func (db *DB) CreateCertificate(ctx context.Context, cert *types.Certificate) error {
	query := `
		INSERT INTO certificates (id, order_id, certificate, revoked, revocation_reason, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.ExecContext(ctx, query,
		cert.ID,
		cert.OrderID,
		cert.Certificate,
		cert.Revoked,
		cert.RevocationReason,
		cert.RevokedAt.Time,
	)
	if err != nil {
		return fmt.Errorf("error creating certificate: %w", err)
	}
	return nil
}

func (db *DB) UpdateAuthorization(ctx context.Context, authz *types.Authorization) error {
	query := `
		UPDATE authorizations
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`
	var id string
	err := db.QueryRowContext(ctx, query, authz.ID, authz.Status).Scan(&id)
	if err != nil {
		return fmt.Errorf("error updating authorization: %w", err)
	}
	return nil
}
