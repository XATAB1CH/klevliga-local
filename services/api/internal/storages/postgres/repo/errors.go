package repo

import (
	"errors"
	"fmt"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	"github.com/XATAB1CH/klev-liga/pkg/pg"
)

// wrapDBError нормализует ошибку БД в доменную, сохраняя исходную в цепочке,
// чтобы вышестоящие слои могли уточнить причину (например, constraint name).
func wrapDBError(op string, err error) error {
	if err == nil {
		return nil
	}

	err = pg.ParseError(err)

	switch {
	case errors.Is(err, pg.ErrNoRows):
		return fmt.Errorf("%s: %w", op, apperr.ErrNotFound)
	case errors.Is(err, pg.ErrUniqueViolation):
		return fmt.Errorf("%s: %w: %w", op, apperr.ErrConflict, err)
	case errors.Is(err, pg.ErrForeignKeyViolation),
		errors.Is(err, pg.ErrNotNullViolation),
		errors.Is(err, pg.ErrCheckViolation):
		return fmt.Errorf("%s: %w: %w", op, apperr.ErrValidation, err)
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
