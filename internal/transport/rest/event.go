package rest

import (
	"custom-banking/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Event struct {
	eventService EventService
}

func NewEvent(eventService EventService) Event {
	return Event{eventService: eventService}
}

func (t *Event) InjectRoutes(r *gin.Engine, middlewares ...gin.HandlerFunc) {
	events := r.Group("/event").Use(middlewares...)
	{
		events.GET("/", t.getEventList)
	}
}

func (t Event) getEventList(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting current user error", err))
		return
	}

	domainList, err := t.eventService.GetEventList(ctx, userID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("getting events list error", err))
		return
	}

	var list []models.Event
	for _, event := range domainList {
		list = append(list, models.Event{
			ID:       event.ID,
			UserID:   event.UserID,
			Type:     event.Type,
			Message:  event.Message,
			Metadata: event.Metadata,
			DateTime: event.DateTime,
		})
	}

	ctx.JSON(http.StatusOK, list)
}
