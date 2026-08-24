package repository

import (
	"context"
	"time"

	"github.com/couchbase/gocb/v2"
)

type APIKey struct {
	ID         string    `json:"id"`
	OwnerID    string    `json:"owner_id"`
	Name       string    `json:"name"`
	Hash       string    `json:"hash"`
	Status     string    `json:"status"`
	RateLimit  int       `json:"rate_limit"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at,omitempty"`
}

type APIKeyRepository struct {
	CB *Couchbase
}

func (r *APIKeyRepository) Create(ctx context.Context, rec APIKey) error {
	return r.CB.Keyspace("api_keys").Upsert(rec.ID, rec, &gocb.UpsertOptions{Context: ctx})
}

func (r *APIKeyRepository) FindByHash(ctx context.Context, hash string) (APIKey, error) {
	var rec APIKey
	// This query expects the index from migrations/001_indexes.sql.
	q := "SELECT a.* FROM `api_monetization`.`app`.`api_keys` a WHERE a.hash = $hash AND a.status = 'ACTIVE' LIMIT 1"
	rows, err := r.CB.Cluster.Query(q, &gocb.QueryOptions{
		NamedParameters: map[string]any{"hash": hash}, Context: ctx,
	})
	if err != nil {
		return rec, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Row(&rec); err != nil {
			return rec, err
		}
		return rec, nil
	}
	if err := rows.Err(); err != nil {
		return rec, err
	}
	return rec, gocb.ErrDocumentNotFound
}
