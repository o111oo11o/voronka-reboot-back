package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"voronka/internal/domain"
	"voronka/internal/service"
)

type merchHandler struct {
	merch  service.MerchService
	orders service.OrderService
}

func newMerchHandler(merch service.MerchService, orders service.OrderService) *merchHandler {
	return &merchHandler{merch: merch, orders: orders}
}

// --- Public ---

func (h *merchHandler) listItems(c *gin.Context) {
	items, err := h.merch.ListItems(c.Request.Context(), domain.MerchItemFilter{
		InStockOnly: true,
		Limit:       50,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *merchHandler) getItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, err := h.merch.GetItem(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *merchHandler) placeOrder(c *gin.Context) {
	var req domain.PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.orders.PlaceOrder(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, order)
}

// --- Admin ---

func (h *merchHandler) listAllItems(c *gin.Context) {
	items, err := h.merch.ListItems(c.Request.Context(), domain.MerchItemFilter{
		InStockOnly: false,
		Limit:       200,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *merchHandler) createItem(c *gin.Context) {
	var req domain.CreateMerchItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.merch.CreateItem(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *merchHandler) updateItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req domain.UpdateMerchItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.merch.UpdateItem(c.Request.Context(), id, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *merchHandler) deleteItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.merch.DeleteItem(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *merchHandler) listOrders(c *gin.Context) {
	orders, err := h.orders.ListOrders(c.Request.Context(), domain.OrderFilter{Limit: 50})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *merchHandler) updateOrderStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req domain.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.orders.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}
