package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

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
