// Package water описывает каталог водоёмов.
package water

import (
	"time"

	"github.com/google/uuid"
)

type Water struct {
	ID          uuid.UUID
	Slug        string
	Name        string
	Kind        string
	RegionID    *uuid.UUID
	Description *string
	CreatedAt   time.Time
}

const (
	KindRiver     = "river"
	KindLake      = "lake"
	KindReservoir = "reservoir"
	KindPond      = "pond"
	KindOther     = "other"
)
