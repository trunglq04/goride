package http

import (
	"log"
	"net/http"
	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	"github.com/trunglq04/goride/shared/types"

	"github.com/gin-gonic/gin"
)

type HttpHandler struct {
	Service domain.TripService
}

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (h *HttpHandler) HandleTripPreview(c *gin.Context) {
	var reqBody previewTripRequest
	if err := c.ShouldBindBodyWithJSON(&reqBody); err != nil {
		c.JSON(http.StatusInternalServerError, "Failed to parse JSON data")
		return
	}

	t, err := h.Service.GetRoute(c, &reqBody.Pickup, &reqBody.Destination, true)
	if err != nil {
		log.Println(err)
	}

	c.JSON(http.StatusOK, t)
}
