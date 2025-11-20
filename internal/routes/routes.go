package routes

import (
	"net/http"

	"github.com/MarcosAndradeV/go-ecommerce/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

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
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		})
		r.Get("/dashboard", storeH.AdminDashboardHandler)
		r.Post("/create", storeH.AdminCreateProductHandler)
	})

	return r
}
