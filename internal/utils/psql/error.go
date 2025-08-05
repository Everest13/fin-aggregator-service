package psql

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	UniqueViolation     = "23505"
	ForeignKeyViolation = "23503"
)

func MapPostgresError(msg string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return status.Error(codes.NotFound, msg)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case UniqueViolation:
			return status.Error(codes.AlreadyExists, msg)
		case ForeignKeyViolation:
			return status.Error(codes.FailedPrecondition, msg)
		default:
			return status.Error(codes.Internal, msg)
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, msg)
	}

	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, msg)
	}

	return status.Error(codes.Internal, msg)
}
