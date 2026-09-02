package postgresql

import (
	"context"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func exec(ctx context.Context, db *pgxpool.Pool, query string, args ...any) error {
	cmdTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
func Update(ctx context.Context, db *pgxpool.Pool, query Query) error {
	return exec(ctx, db, query.query, query.args)
}

func Find[T any](ctx context.Context, db *pgxpool.Pool, query Query) (*T, error) {
	rows, err := db.Query(ctx, query.query, query.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func Create(ctx context.Context, db *pgxpool.Pool, query Query)error{
	_,err:= db.Exec(ctx, query.query, query.args...)
	return err
}

// assumed "id" is a valid column of type pgtype.UUID and is returned
func CreateAndReturnID(ctx context.Context, db *pgxpool.Pool, query Query) (uuid.UUID, error) {
	var pgID pgtype.UUID
	err := db.QueryRow(ctx, query.query+" RETURNING id", query.args...).Scan(&pgID)
	if err != nil {
		return uuid.Nil(), err
	}
	return uuid.UUID(pgID.Bytes), nil
}
