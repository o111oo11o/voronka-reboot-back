package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"voronka/internal/domain"
	"voronka/internal/service"
)

type eventHandler struct {
	events service.EventService
}

func newEventHandler(events service.EventService) *eventHandler {
	return &eventHandler{events: events}
}

func (h *eventHandler) list(c *gin.Context) {
	events, err := h.events.List(c.Request.Context(), domain.EventFilter{Limit: 50})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *eventHandler) getByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	event, err := h.events.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *eventHandler) create(c *gin.Context) {
	var req domain.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, err := h.events.Create(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *eventHandler) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req domain.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, err := h.events.Update(c.Request.Context(), id, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *eventHandler) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.events.Delete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
