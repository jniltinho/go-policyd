// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: queries.sql

package db

import (
	"context"
	"database/sql"
)

const createEvent = `-- name: CreateEvent :execresult
INSERT INTO events (sasl_username,sender,client_address,recipient_count) VALUES (?,?,?,?)
`

type CreateEventParams struct {
	SaslUsername   string
	Sender         string
	ClientAddress  string
	RecipientCount sql.NullInt32
}

func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, createEvent,
		arg.SaslUsername,
		arg.Sender,
		arg.ClientAddress,
		arg.RecipientCount,
	)
}
