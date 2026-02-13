package handlers

import (
	"net/http"

	"service.data/external/storage/adx"
	"service.data/internal/core/observability/logger"
)

// Handler is the main handler struct for the application.
type Handler struct {
	logger logger.Logger
	adx    *adx.Client
}

// New creates a new Handler instance.
func New(logger logger.Logger, adx *adx.Client) *Handler {
	return &Handler{
		logger: logger,
		adx:    adx,
	}
}

// isHtmx checks if the request was made by HTMX.
func isHtmx(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
