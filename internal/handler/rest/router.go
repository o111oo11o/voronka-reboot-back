package rest

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"

	"voronka/internal/handler/rest/middleware"
	"voronka/internal/handler/ws"
	"voronka/internal/service"
)

//go:embed static/admin.html
var adminHTML []byte

func NewRouter(
	auth service.AuthService,
	users service.UserService,
	events service.EventService,
	merch service.MerchService,
	orders service.OrderService,
	hub *ws.Hub,
	adminToken string,
	uploadsDir string,
) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.Status(200) })

	// Serve uploaded files and admin panel
	r.Static("/uploads", uploadsDir)
	r.GET("/admin", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", adminHTML)
	})

	api := r.Group("/api/v1")
	{
		// Auth (public)
		ah := newAuthHandler(auth)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", ah.register)
			authGroup.POST("/verify", ah.verify)
			authGroup.POST("/refresh", ah.refresh)
			authGroup.POST("/confirm", ah.confirmRegistration)
		}

		// Users (auth protected)
		uh := newUserHandler(users)
		u := api.Group("/users", middleware.Auth(auth))
		{
			u.GET("", uh.list)
			u.GET("/:id", uh.getByID)
			u.PUT("/:id", uh.update)
			u.DELETE("/:id", uh.delete)
		}

		// Events (public read)
		eh := newEventHandler(events)
		ev := api.Group("/events")
		{
			ev.GET("", eh.list)
			ev.GET("/:id", eh.getByID)
		}

		// Merch (public read)
		mh := newMerchHandler(merch, orders)
		mi := api.Group("/merch")
		{
			mi.GET("", mh.listItems)
			mi.GET("/:id", mh.getItem)
		}

		// Orders (authenticated write)
		api.POST("/orders", middleware.Auth(auth), mh.placeOrder)

		// Admin — all routes require X-Admin-Token
		uploadH := newUploadHandler(uploadsDir)
		admin := api.Group("/admin", middleware.Admin(adminToken))
		{
			admin.POST("/upload", uploadH.upload)

			admin.GET("/users", uh.list)
			admin.GET("/merch", mh.listAllItems)

			adminEvents := admin.Group("/events")
			{
				adminEvents.POST("", eh.create)
				adminEvents.PUT("/:id", eh.update)
				adminEvents.DELETE("/:id", eh.delete)
			}

			adminMerch := admin.Group("/merch")
			{
				adminMerch.POST("", mh.createItem)
				adminMerch.PUT("/:id", mh.updateItem)
				adminMerch.DELETE("/:id", mh.deleteItem)
			}

			adminOrders := admin.Group("/orders")
			{
				adminOrders.GET("", mh.listOrders)
				adminOrders.PUT("/:id/status", mh.updateOrderStatus)
			}
		}
	}

	// WebSocket
	chatHandler := ws.NewChatHandler(hub)
	r.GET("/ws/chat", chatHandler.ServeWS)

	return r
}
