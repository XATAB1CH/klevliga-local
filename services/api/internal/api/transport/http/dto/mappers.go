package dto

import (
	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/domains/forum"
	"github.com/XATAB1CH/klev-liga/internal/domains/report"
	"github.com/XATAB1CH/klev-liga/internal/domains/social"
	"github.com/XATAB1CH/klev-liga/internal/domains/water"
)

func CategoriesFromDomain(cs []*forum.Category) []*Category {
	out := make([]*Category, 0, len(cs))
	for _, c := range cs {
		out = append(out, CategoryFromDomain(c))
	}

	return out
}

func TopicsFromDomain(ts []*forum.Topic) []*Topic {
	out := make([]*Topic, 0, len(ts))
	for _, t := range ts {
		out = append(out, TopicFromDomain(t))
	}

	return out
}

func PostsFromDomain(ps []*forum.Post) []*Post {
	out := make([]*Post, 0, len(ps))
	for _, p := range ps {
		out = append(out, PostFromDomain(p))
	}

	return out
}

func ReportsFromDomain(rs []*report.Report) []*Report {
	out := make([]*Report, 0, len(rs))
	for _, r := range rs {
		out = append(out, ReportFromDomain(r))
	}

	return out
}

func WatersFromDomain(ws []*water.Water) []*Water {
	out := make([]*Water, 0, len(ws))
	for _, w := range ws {
		out = append(out, WaterFromDomain(w))
	}

	return out
}

func UsersFromDomain(us []*auth.User) []*User {
	out := make([]*User, 0, len(us))
	for _, u := range us {
		out = append(out, UserFromDomain(u))
	}

	return out
}

func NotificationsFromDomain(ns []*social.Notification) []*Notification {
	out := make([]*Notification, 0, len(ns))
	for _, n := range ns {
		out = append(out, NotificationFromDomain(n))
	}

	return out
}
