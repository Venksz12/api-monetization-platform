package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
)

type UsageRollup struct {
	ID            string    `json:"id"`
	OwnerID       string    `json:"owner_id"`
	Period        string    `json:"period"`
	TotalUnits    int64     `json:"total_units"`
	TotalRequests int64     `json:"total_requests"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UsageRepository struct {
	CB *Couchbase
}

func (r *UsageRepository) Add(ctx context.Context, ownerID, period string, units int64) error {
	id := fmt.Sprintf("usage::%s::%s", ownerID, period)
	_, err := r.CB.Keyspace("usage").MutateIn(id, []gocb.MutateInSpec{
		gocb.UpsertSpec("owner_id", ownerID, nil),
		gocb.UpsertSpec("period", period, nil),
		gocb.IncrementSpec("total_units", units, &gocb.MutateInSpecOptions{CreatePath: true}),
		gocb.IncrementSpec("total_requests", 1, &gocb.MutateInSpecOptions{CreatePath: true}),
		gocb.UpsertSpec("updated_at", time.Now().UTC(), nil),
	}, &gocb.MutateInOptions{Context: ctx, StoreSemantic: gocb.StoreSemanticsUpsert})
	return err
}

func (r *UsageRepository) Get(ctx context.Context, ownerID, period string) (UsageRollup, error) {
	var u UsageRollup
	res, err := r.CB.Keyspace("usage").Get(fmt.Sprintf("usage::%s::%s", ownerID, period), &gocb.GetOptions{Context: ctx})
	if err != nil {
		return u, err
	}
	return u, res.Content(&u)
}
