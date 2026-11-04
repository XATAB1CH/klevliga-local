package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
)

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.forum.ListCategories(r.Context())
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusOK, dto.CategoriesFromDomain(categories))
}

func (h *Handler) listTopics(w http.ResponseWriter, r *http.Request) {
	categoryID, err := optionalUUIDQuery(r, "category_id")
	if err != nil {
		writeError(w, r, err)

		return
	}

	cursorAt, cursorID, err := timeCursorParam(r)
	if err != nil {
		writeError(w, r, err)

		return
	}

	limit := limitParam(r)

	topics, err := h.forum.ListTopics(r.Context(), categoryID, cursorAt, cursorID, limit+1)
	if err != nil {
		writeError(w, r, err)

		return
	}

	var next *string
	if len(topics) > int(limit) {
		last := topics[int(limit)]
		next = strPtr(encodeTimeCursor(last.LastActivityAt, last.ID))
		topics = topics[:limit]
	}

	writeJSON(w, http.StatusOK, dto.TopicPage{Items: dto.TopicsFromDomain(topics), NextCursor: next})
}

func (h *Handler) createTopic(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)

		return
	}

	var req dto.NewTopicRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	if req.CategoryID == uuid.Nil || len(req.Title) < 5 || req.Body == "" {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	topic, err := h.forum.CreateTopic(r.Context(), req.CategoryID, user.ID, req.Title, req.Body)
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusCreated, dto.TopicFromDomain(topic))
}

func (h *Handler) getTopic(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	topic, err := h.forum.GetTopic(r.Context(), id)
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusOK, dto.TopicFromDomain(topic))
}

func (h *Handler) listPosts(w http.ResponseWriter, r *http.Request) {
	topicID, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	cursorAt, cursorID, err := timeCursorParam(r)
	if err != nil {
		writeError(w, r, err)

		return
	}

	limit := limitParam(r)

	posts, err := h.forum.ListPosts(r.Context(), topicID, cursorAt, cursorID, limit+1)
	if err != nil {
		writeError(w, r, err)

		return
	}

	var next *string
	if len(posts) > int(limit) {
		last := posts[int(limit)]
		next = strPtr(encodeTimeCursor(last.CreatedAt, last.ID))
		posts = posts[:limit]
	}

	writeJSON(w, http.StatusOK, dto.PostPage{Items: dto.PostsFromDomain(posts), NextCursor: next})
}

func (h *Handler) reply(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)

		return
	}

	topicID, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	var req dto.ReplyRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	if req.Body == "" {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	post, err := h.forum.CreatePost(r.Context(), topicID, user.ID, req.Body, req.ParentPostID)
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusCreated, dto.PostFromDomain(post))
}
