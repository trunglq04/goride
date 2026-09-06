package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	"github.com/trunglq04/goride/shared/types"
)

type HttpHandler struct {
	Service domain.TripService
}

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (h *HttpHandler) HandleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode("Failed to parse JSON data")
		return
	}

	t, err := h.Service.GetRoute(r.Context(), &reqBody.Pickup, &reqBody.Destination, true)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to get route", "user_id", reqBody.UserID, "err", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(t)
}
