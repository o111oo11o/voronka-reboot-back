package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"voronka/internal/domain"
	"voronka/internal/service"
)

type authHandler struct {
	auth service.AuthService
}

func newAuthHandler(auth service.AuthService) *authHandler {
	return &authHandler{auth: auth}
}

// POST /api/v1/auth/register
func (h *authHandler) register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.auth.Register(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// POST /api/v1/auth/verify  (existing users — submit the 4-digit code)
func (h *authHandler) verify(c *gin.Context) {
	var req domain.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.auth.VerifyCode(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, pair)
}

// POST /api/v1/auth/confirm — browser polls this with the deeplink token after returning from Telegram
func (h *authHandler) confirmRegistration(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	pair, err := h.auth.PollRegistration(c.Request.Context(), req.Token)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, pair)
}

// POST /api/v1/auth/refresh
func (h *authHandler) refresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, pair)
}
