package domain

import (
	"time"

	"github.com/google/uuid"
)

type MerchItem struct {
	ID          uuid.UUID
	Name        string
	Description string
	PriceCents  int64
	Stock       int
	ImageURLs   []string // ordered list of image URLs
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateMerchItemRequest struct {
	Name        string
	Description string
	PriceCents  int64
	Stock       int
	ImageURLs   []string
}

type UpdateMerchItemRequest struct {
	Name        *string
	Description *string
	PriceCents  *int64
	Stock       *int // used for restock / adjustment
	ImageURLs   *[]string
}

type MerchItemFilter struct {
	InStockOnly bool
	Limit       int
	Offset      int
}

// -----------------------------------------------------------------------

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusFulfilled OrderStatus = "fulfilled"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID            uuid.UUID
	CustomerName  string
	CustomerEmail string
	UserID        *uuid.UUID
	Status        OrderStatus
	TotalCents    int64
	Items         []OrderItem
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OrderItem struct {
	ID               uuid.UUID
	OrderID          uuid.UUID
	MerchItemID      uuid.UUID
	Quantity         int
	PriceAtTimeCents int64
}

type OrderLineItem struct {
	MerchItemID uuid.UUID
	Quantity    int
}

type PlaceOrderRequest struct {
	CustomerName  string
	CustomerEmail string
	UserID        *uuid.UUID
	Items         []OrderLineItem
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus
}

type OrderFilter struct {
	Status *OrderStatus
	Limit  int
	Offset int
}
