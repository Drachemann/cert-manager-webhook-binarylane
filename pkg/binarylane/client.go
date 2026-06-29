package binarylane

import "context"

type Client interface {
	CreateRecord(ctx context.Context, domain string, record Record) (*Record, error)
	DeleteRecord(ctx context.Context, domain string, recordID int) error
	GetRecord(ctx context.Context, domain string, recordID int) (*Record, error)
}
