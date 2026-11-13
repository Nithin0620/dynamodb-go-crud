package routes

import(
	"github.com/go-chi/chi/"
	"github.com/go-chi/chi/middleware"
	// "github.com"
	ServerConfig "github.com/Nithin0620/dynamodb-go-crud/config"
	HealthHandler "github.com/Nithin0620/dynamodb-go-crud/internal/handlers/health"
	ProductHandler "github.com/Nithin0620/dynamodb-go-crud/internal/handlers/product"
)

type Router Struct {
	config *Config
	router *chi.Mux
}

func NewRouter() *Router{
	return &Router{
		config: newConfig().setTimeout(serviceConfig.GetConfig().Timeout),
		router: chi.NewRouter(),
	}
}

func (r *Router) SetRouters(repository adapter.Interface) *chi.Mux {
	r.setConfigsRouters()
	r.RouterHealth(repository)
	r.RouterProduct(repository)
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
	handler := HealthHandler.NewHealthHandler(repository)

	r.router.Route("/health",func(route chi.Router){
		route.Get("/",handler.Get)
		route.Put("/",handler.Put)
		route.Post("/",handler.Post)
		route.Delete("/",handler.Delete)
		route.Options("/",handler.Options)
	})
}

func (r *Router) RouterProduct(repository adapter.Interface) {
	handler := ProductHandler.NewProductHandler(repository)

	r.router.Route("/product",func(route chi.Router){
		route.Get("/",handler.Get)
		route.Put("/{ID}",handler.Put)
		route.Post("/",handler.Post)
		route.Delete("/{ID}",handler.Delete)
		route.Options("/",handler.Options)
	})
}

func (r *Router) EnableLogger() *Router{ 
	r.router.Use(middleware.Logger)
	return r
}
func (r *Router) EnableTimeout() *Router{
	r.router.Use(middleware.Timeout(r.config.GetTimeout()))
	return r
}

func (r *Router) EnableCORS() *Router{
	r.router.Use(r.config.cors)

	return r
}

func (r *Router) EnableRecover() *Router{
	r.router.Use(middleware.Recoverer)

	return r
}

func (r *Router) EnableRequestID() *Router{
	r.router.Use(middleware.RequestID)

	return r
}

func (r *Router) EnableRealIP() *Router{
	r.router.Use(middleware.RealIP)

	return r
}