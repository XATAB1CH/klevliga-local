package pg

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNoRows               = errors.New("no rows in result set")
	ErrUniqueViolation      = errors.New("unique violation")
	ErrForeignKeyViolation  = errors.New("foreign key violation")
	ErrNotNullViolation     = errors.New("not null violation")
	ErrCheckViolation       = errors.New("check violation")
	ErrSerializationFailure = errors.New("serialization failure")
	ErrDeadlockDetected     = errors.New("deadlock detected")
)

func ParseError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w", ErrNoRows)
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	var kind error

	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		kind = ErrUniqueViolation
	case pgerrcode.ForeignKeyViolation:
		kind = ErrForeignKeyViolation
	case pgerrcode.NotNullViolation:
		kind = ErrNotNullViolation
	case pgerrcode.CheckViolation:
		kind = ErrCheckViolation
	case pgerrcode.SerializationFailure:
		kind = ErrSerializationFailure
	case pgerrcode.DeadlockDetected:
		kind = ErrDeadlockDetected
	default:
		return err
	}

	if pgErr.Detail == "" {
		return fmt.Errorf("%w", kind)
	}

	return fmt.Errorf("%w: %s", kind, pgErr.Detail)
}
