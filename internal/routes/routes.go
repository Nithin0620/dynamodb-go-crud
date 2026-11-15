package routes

import (
	"net/http"

	ServerConfig "github.com/Nithin0620/dynamodb-go-crud/config"
	AuthHandler "github.com/Nithin0620/dynamodb-go-crud/internal/handlers/auth"
	HealthHandler "github.com/Nithin0620/dynamodb-go-crud/internal/handlers/health"
	OrderHandler "github.com/Nithin0620/dynamodb-go-crud/internal/handlers/order"
	ProductHandler "github.com/Nithin0620/dynamodb-go-crud/internal/handlers/product"
	"github.com/Nithin0620/dynamodb-go-crud/internal/repository/adapter"
	"github.com/Nithin0620/dynamodb-go-crud/internal/routes/middlewares"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Router struct {
	config *Config
	router *chi.Mux
}

func NewRouter() *Router {
	return &Router{
		config: NewConfig().SetTimeout(ServerConfig.GetConfig().Timeout),
		router: chi.NewRouter(),
	}
}

func (r *Router) SetRouters(repository adapter.Interface) *chi.Mux {
	r.setConfigsRouters()

	// Public Health
	r.RouterHealth(repository)

	// Auth routes (register, login)
	r.RouterAuth(repository)

	// Public Product browsing
	r.RouterProduct(repository)

	// User Order routes (checkout puzzle & place order)
	r.RouterOrders(repository)

	// Admin / Store Owner routes (manage own inventory & store orders)
	r.RouterAdmin(repository)

	// Serve React Frontend Production Build (from frontend/dist if available)
	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	r.router.Handle("/*", fileServer)

	return r.router
}

func (r *Router) setConfigsRouters() {
	r.EnableCORS()
	r.EnableLogger()
	r.EnableTimeout()
	r.EnableRecover()
	r.EnableRequestID()
	r.EnableRealIP()
}

func (r *Router) RouterHealth(repository adapter.Interface) {
	handler := HealthHandler.NewHandler(repository)
	r.router.Route("/health", func(route chi.Router) {
		route.Get("/", handler.Get)
		route.Options("/", handler.Options)
	})
}

func (r *Router) RouterAuth(repository adapter.Interface) {
	handler := AuthHandler.NewHandler(repository)
	r.router.Route("/auth", func(route chi.Router) {
		route.Post("/register", handler.Register)
		route.Post("/login", handler.Login)
	})
}

func (r *Router) RouterProduct(repository adapter.Interface) {
	handler := ProductHandler.NewHandler(repository)
	r.router.Route("/products", func(route chi.Router) {
		route.Get("/", handler.Get)
		route.Get("/{ID}", handler.Get)
		route.Options("/", handler.Options)
	})
}

func (r *Router) RouterOrders(repository adapter.Interface) {
	handler := OrderHandler.NewHandler(repository)
	r.router.Route("/orders", func(route chi.Router) {
		route.Use(middlewares.AuthMiddleware)
		route.Post("/checkout-challenge", handler.RequestChallenge)
		route.Post("/confirm", handler.ConfirmOrder)
		route.Get("/my-orders", handler.GetUserOrders)
	})
}

func (r *Router) RouterAdmin(repository adapter.Interface) {
	productHandler := ProductHandler.NewHandler(repository)
	orderHandler := OrderHandler.NewHandler(repository)

	r.router.Route("/admin", func(route chi.Router) {
		route.Use(middlewares.AuthMiddleware)
		route.Use(middlewares.RequireRole("admin"))

		// Inventory management for Store Owner
		route.Get("/products", productHandler.GetMyProducts)
		route.Post("/products", productHandler.Post)
		route.Put("/products/{ID}", productHandler.Put)
		route.Delete("/products/{ID}", productHandler.Delete)

		// Order management for Store Owner
		route.Get("/orders", orderHandler.GetStoreOwnerOrders)
		route.Put("/orders/{id}/status", orderHandler.UpdateOrderStatus)
	})
}

func (r *Router) EnableLogger() *Router {
	r.router.Use(middleware.Logger)
	return r
}

func (r *Router) EnableTimeout() *Router {
	r.router.Use(middleware.Timeout(r.config.GetTimeout()))
	return r
}

func (r *Router) EnableCORS() *Router {
	r.router.Use(r.config.Cors)
	return r
}

func (r *Router) EnableRecover() *Router {
	r.router.Use(middleware.Recoverer)
	return r
}

func (r *Router) EnableRequestID() *Router {
	r.router.Use(middleware.RequestID)
	return r
}

func (r *Router) EnableRealIP() *Router {
	r.router.Use(middleware.RealIP)
	return r
}
