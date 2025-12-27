package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"Pay2Go/internal/domain/entities"
	"Pay2Go/internal/domain/errors"
)

// PartnerRepository implements ports.PartnerRepository for PostgreSQL
type PartnerRepository struct {
	db *sql.DB
}

// NewPartnerRepository creates a new PostgreSQL partner repository
func NewPartnerRepository(db *sql.DB) *PartnerRepository {
	return &PartnerRepository{db: db}
}

// Create creates a new partner
func (r *PartnerRepository) Create(ctx context.Context, partner *entities.Partner) error {
	query := `
		INSERT INTO partners (
			id, name, email, api_key_hash, api_key_prefix, is_active,
			rate_limit_per_minute, webhook_url, webhook_secret, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	metadataJSON, _ := json.Marshal(partner.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		partner.ID,
		partner.Name,
		partner.Email,
		partner.APIKeyHash,
		partner.APIKeyPrefix,
		partner.IsActive,
		partner.RateLimitPerMinute,
		partner.WebhookURL,
		partner.WebhookSecret,
		metadataJSON,
		partner.CreatedAt,
		partner.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create partner: %w", err)
	}

	return nil
}

// GetByID retrieves a partner by ID
func (r *PartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Partner, error) {
	query := `
		SELECT id, name, email, api_key_hash, api_key_prefix, is_active,
			   rate_limit_per_minute, webhook_url, webhook_secret, metadata,
			   created_at, updated_at
		FROM partners
		WHERE id = $1 AND deleted_at IS NULL
	`

	var partner entities.Partner
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&partner.ID,
		&partner.Name,
		&partner.Email,
		&partner.APIKeyHash,
		&partner.APIKeyPrefix,
		&partner.IsActive,
		&partner.RateLimitPerMinute,
		&partner.WebhookURL,
		&partner.WebhookSecret,
		&metadataJSON,
		&partner.CreatedAt,
		&partner.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrPartnerNotFound
		}
		return nil, fmt.Errorf("failed to get partner: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &partner.Metadata)
	}

	return &partner, nil
}

// GetByEmail retrieves a partner by email
func (r *PartnerRepository) GetByEmail(ctx context.Context, email string) (*entities.Partner, error) {
	query := `
		SELECT id FROM partners
		WHERE email = $1 AND deleted_at IS NULL
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, email).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrPartnerNotFound
		}
		return nil, fmt.Errorf("failed to get partner by email: %w", err)
	}

	return r.GetByID(ctx, id)
}

// GetByAPIKeyPrefix retrieves a partner by API key prefix
func (r *PartnerRepository) GetByAPIKeyPrefix(ctx context.Context, prefix string) (*entities.Partner, error) {
	query := `
		SELECT id FROM partners
		WHERE api_key_prefix = $1 AND deleted_at IS NULL
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, prefix).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrPartnerNotFound
		}
		return nil, fmt.Errorf("failed to get partner by API key prefix: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Update updates an existing partner
func (r *PartnerRepository) Update(ctx context.Context, partner *entities.Partner) error {
	query := `
		UPDATE partners SET
			name = $1,
			email = $2,
			is_active = $3,
			rate_limit_per_minute = $4,
			webhook_url = $5,
			webhook_secret = $6,
			updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.ExecContext(ctx, query,
		partner.Name,
		partner.Email,
		partner.IsActive,
		partner.RateLimitPerMinute,
		partner.WebhookURL,
		partner.WebhookSecret,
		partner.UpdatedAt,
		partner.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update partner: %w", err)
	}

	return nil
}

// List retrieves all partners with pagination
func (r *PartnerRepository) List(ctx context.Context, limit, offset int) ([]*entities.Partner, error) {
	query := `
		SELECT id FROM partners
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list partners: %w", err)
	}
	defer rows.Close()

	var partners []*entities.Partner
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		partner, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}

		partners = append(partners, partner)
	}

	return partners, nil
}
