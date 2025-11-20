package routes

import (
	"net/http"
	"net/url" // <--- Import added

	"github.com/MarcosAndradeV/go-ecommerce/internal/handlers"
	"github.com/MarcosAndradeV/go-ecommerce/internal/service" // Import necessário
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Middleware Simplificado: Apenas verifica se tem o cookie
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("sessao_loja")
		if err != nil {
			// Sem cookie = Redireciona para login com next
			nextURL := r.URL.RequestURI()
			http.Redirect(w, r, "/login?msg=faca_login&next="+url.QueryEscape(nextURL), http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(authH *handlers.AuthHandler, storeH *handlers.StoreHandler, authS *service.AuthService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// --- ROTAS PÚBLICAS ---
	r.Get("/", storeH.HomeHandler)
	r.Get("/product/{id}", storeH.ProductDetailHandler)

	// --- AUTH ---
	r.Get("/register", authH.RegisterPageHandler)
	r.Post("/register", authH.RegisterPostHandler)
	r.Get("/login", authH.LoginPageHandler)
	r.Post("/do-login", authH.LoginPostHandler)
	r.Get("/logout", authH.LogoutHandler)

	// --- ROTAS PROTEGIDAS (Usa o Middleware) ---
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Get("/dashboard", authH.DashboardHandler)
		r.Get("/cart", storeH.ViewCartHandler)
		r.Get("/add-to-cart", storeH.AddToCartHandler)
		r.Get("/remove-from-cart", storeH.RemoveFromCartHandler) // <--- Nova rota
		r.Get("/checkout", storeH.CheckoutPageHandler)
		r.Post("/checkout", storeH.CheckoutPageHandler) // <--- Permitir POST para seleção
		r.Post("/payment", storeH.PaymentPageHandler)   // <--- Nova rota de pagamento
		r.Post("/purchase", storeH.PurchaseHandler)
	})

	// --- ADMIN ---
	r.Route("/admin", func(r chi.Router) {
		r.Get("/dashboard", storeH.AdminDashboardHandler)
		r.Post("/create", storeH.AdminCreateProductHandler)
		r.Get("/edit/product/{product_id}", storeH.EditProductFormHandler)
		r.Post("/edit/product", storeH.EditProductHandler)
	})

	return r
}
