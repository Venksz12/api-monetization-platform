package repository

import (
	"context"
	"time"

	"github.com/couchbase/gocb/v2"
)

type Invoice struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Period    string    `json:"period"`
	Subtotal  int64     `json:"subtotal_minor"`
	Tax       int64     `json:"tax_minor"`
	Total     int64     `json:"total_minor"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InvoiceRepository struct {
	CB *Couchbase
}

func (r *InvoiceRepository) Upsert(ctx context.Context, inv Invoice) error {
	_, err := r.CB.Keyspace("invoices").Upsert(inv.ID, inv, &gocb.UpsertOptions{Context: ctx})
	return err
}

func (r *InvoiceRepository) Get(ctx context.Context, id string) (Invoice, error) {
	var inv Invoice
	res, err := r.CB.Keyspace("invoices").Get(id, &gocb.GetOptions{Context: ctx})
	if err != nil {
		return inv, err
	}
	return inv, res.Content(&inv)
}
