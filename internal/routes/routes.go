package routes

import (
	"net/http"

	"github.com/MarcosAndradeV/go-ecommerce/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// func SetupRoutes(r *chi.Mux, db *mongo.Database) {

// 	r.Use(middleware.Logger)
// 	r.Use(middleware.Recoverer)

// 	fileServer := http.FileServer(http.Dir("./static"))
// 	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

// 	sr := repository.NewStoreRepository(db)
// 	ss := service.NewStoreService(sr)
// 	sh := handlers.NewStoreHandler(ss)

// 	// // Loja (público)
// 	r.Get("/", sh.HomeHandler)

// 	r.Get("/product{id}", sh.ProductDetailHandler)
// 	// //Carrinho
// 	// r.Get("/checkout", h.CheckoutPageHandler)
// 	// r.Post("/purchase", h.PurchaseHandler)
// 	// r.Get("/sucess", func(w http.ResponseWriter, r *http.Request) {
// 	// 	handlers.RenderTemplate(w, r, "sucess.html", nil)
// 	// })
// 	// //ADMIM
// 	// r.Get("/login", h.LoginPage)
// 	// r.Get("/do-login", h.PerformLogin)
// 	// r.Get("/logout", h.PerformLogout)

// 	// r.Route("/admin", func(r chi.Router){
// 	// 	r.Get("/", h.AdminDashboard)
// 	// 	r.Get("/create", h.AdminCreateProduct)
// 	// })
// }

// NewRouter recebe os handlers já instanciados e devolve o roteador configurado
func NewRouter(authH *handlers.AuthHandler, storeH *handlers.StoreHandler) *chi.Mux {
	r := chi.NewRouter()

	// --- Middlewares Globais ---
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(middleware.RealIP) // Útil se for pra produção

	// --- Arquivos Estáticos (CSS/JS/Imagens) ---
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// --- ROTAS DA LOJA (Públicas) ---
	r.Get("/", storeH.HomeHandler)
	r.Get("/product/{id}", storeH.ProductDetailHandler)
	r.Get("/checkout", storeH.CheckoutPageHandler)
	r.Post("/purchase", storeH.PurchaseHandler)

	// --- ROTAS DE AUTENTICAÇÃO ---

	// Registro
	r.Get("/register", authH.RegisterPageHandler)
	r.Post("/register", authH.RegisterPostHandler)

	// Login / Logout
	r.Get("/login", authH.LoginPageHandler)
	r.Post("/do-login", authH.LoginPostHandler)
	r.Get("/logout", authH.LogoutHandler)

	// Área do Cliente (Protegida por verificação interna no handler)
	r.Get("/dashboard", authH.DashboardHandler)

	// --- ÁREA ADMIN (Opcional: Agrupar rotas) ---
	r.Route("/admin", func(r chi.Router) {
		// Aqui você poderia adicionar um middleware específico se quisesse
		// r.Use(AdminMiddleware)

		// Como seu AdminHandler atual é hardcoded e verifica cookie,
		// você pode apontar direto para o handler se tiver criado um.
		// Exemplo: r.Get("/", adminH.DashboardHandler)
	})

	return r
}
