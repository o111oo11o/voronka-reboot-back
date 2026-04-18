package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"voronka/internal/domain"
	"voronka/internal/service"
)

type userHandler struct {
	users service.UserService
}

func newUserHandler(users service.UserService) *userHandler {
	return &userHandler{users: users}
}

func (h *userHandler) create(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.users.Create(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *userHandler) getByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *userHandler) update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.users.Update(c.Request.Context(), id, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *userHandler) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.users.Delete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *userHandler) list(c *gin.Context) {
	users, err := h.users.List(c.Request.Context(), domain.UserFilter{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrBadRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
