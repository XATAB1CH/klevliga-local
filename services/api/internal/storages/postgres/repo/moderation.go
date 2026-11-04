package repo

import (
	"context"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/moderation"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

func (s *Store) CreateAbuseReport(
	ctx context.Context,
	reporterID uuid.UUID,
	targetType string,
	targetID uuid.UUID,
	reason string,
) (*moderation.AbuseReport, error) {
	row, err := s.queries(ctx).CreateAbuseReport(ctx, sqlc.CreateAbuseReportParams{
		ReporterID: reporterID,
		TargetType: targetType,
		TargetID:   targetID,
		Reason:     reason,
	})
	if err != nil {
		return nil, wrapDBError("create abuse report", err)
	}

	return mapAbuseReport(
		row.ID, reporterID, targetType, targetID, reason, row.Status, row.CreatedAt,
	), nil
}
