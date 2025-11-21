package handlers

import (
	"net/http"
	"time"

	"github.com/MarcosAndradeV/go-ecommerce/internal/service"
)

type AuthHandler struct {
	Service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{Service: s}
}

// --- LOGIN ---

func (h *AuthHandler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	data := map[string]any{
		"Next": next,
	}
	RenderTemplate(w, r, "login.html", data)
}

func (h *AuthHandler) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	next := r.FormValue("next") // <--- Captura o next

	// 1. Admin Hardcoded
	if email == "admin" && password == "admin123" {
		http.SetCookie(w, &http.Cookie{
			Name:  "sessao_admin",
			Value: "true",
			Path:  "/", // <--- OBRIGATÓRIO
		})
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	// 2. Cliente
	user, err := h.Service.AuthenticateUser(email, password)
	if err != nil {
		// Se falhar, mostra o formulário novamente com mensagem de erro
		data := map[string]any{
			"Next":  next,
			"Error": "E-mail ou senha incorretos",
			"Email": email,
		}
		RenderTemplate(w, r, "login.html", data)
		return
	}

	// --- CORREÇÃO CRÍTICA: Path "/" ---
	http.SetCookie(w, &http.Cookie{
		Name:     "sessao_loja",
		Value:    user.ID.Hex(),
		Path:     "/", // <--- ISSO CONSERTA O LOOP DE LOGIN
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	// Se tiver next, vai pra lá. Senão, home.
	if next != "" {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Para sair, "matamos" o cookie definindo MaxAge -1
	http.SetCookie(w, &http.Cookie{Name: "sessao_loja", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "sessao_admin", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ... (Mantenha RegisterPageHandler e RegisterPostHandler como estão)
func (h *AuthHandler) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, r, "register.html", nil)
}

func (h *AuthHandler) RegisterPostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	err := h.Service.RegisterCustomer(name, email, password)
	if err != nil {
		http.Redirect(w, r, "/register?error=true", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/login?success=created", http.StatusSeeOther)
}

func (h *AuthHandler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessao_loja")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, orders, err := h.Service.GetDashboardData(cookie.Value)
	if err != nil {
		// Se o usuário não existe mais, desloga
		http.Redirect(w, r, "/logout", http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"User":   user,
		"Orders": orders,
	}
	RenderTemplate(w, r, "dashboard.html", data)
}
