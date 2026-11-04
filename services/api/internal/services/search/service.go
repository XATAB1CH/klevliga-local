// Package search — поиск по публичным материалам.
package search

import (
	"context"

	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/domains/forum"
	"github.com/XATAB1CH/klev-liga/internal/domains/report"
	"github.com/XATAB1CH/klev-liga/internal/domains/water"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

const (
	TypeAll     = "all"
	TypeTopics  = "topics"
	TypeReports = "reports"
	TypeWaters  = "waters"
	TypeUsers   = "users"
)

type Results struct {
	Topics  []*forum.Topic
	Reports []*report.Report
	Waters  []*water.Water
	Users   []*auth.User
}

type Service struct {
	store *repo.Store
}

func NewService(store *repo.Store) *Service {
	return &Service{store: store}
}

// Search выполняет поиск по выбранным типам (или по всем).
func (s *Service) Search(ctx context.Context, q string, types map[string]bool, limit int32) (*Results, error) {
	run := func(t string) bool { return len(types) == 0 || types[t] }

	res := &Results{
		Topics:  []*forum.Topic{},
		Reports: []*report.Report{},
		Waters:  []*water.Water{},
		Users:   []*auth.User{},
	}

	if run(TypeTopics) {
		topics, err := s.store.SearchTopics(ctx, q, limit)
		if err != nil {
			return nil, err
		}

		res.Topics = topics
	}

	if run(TypeReports) {
		reports, err := s.store.SearchReports(ctx, q, limit)
		if err != nil {
			return nil, err
		}

		res.Reports = reports
	}

	if run(TypeWaters) {
		waters, err := s.store.SearchWaters(ctx, q, limit)
		if err != nil {
			return nil, err
		}

		res.Waters = waters
	}

	if run(TypeUsers) {
		users, err := s.store.SearchUsers(ctx, q, limit)
		if err != nil {
			return nil, err
		}

		res.Users = users
	}

	return res, nil
}
