package handlers

import (
	"net/http"
	"time"

	"github.com/MarcosAndradeV/go-ecommerce/internal/service"
)

type AuthHandler struct {
	Service *service.AuthService
}

// Construtor
func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{Service: s}
}

// --- REGISTRO ---

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

// --- LOGIN (Híbrido) ---

func (h *AuthHandler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, r, "login.html", nil)
}

func (h *AuthHandler) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	// 1. Admin Hardcoded (Prioridade)
	if email == "admin" && password == "admin123" {
		http.SetCookie(w, &http.Cookie{
			Name:    "sessao_admin",
			Value:   "true",
			Path:    "/",
			Expires: time.Now().Add(24 * time.Hour),
		})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// 2. Cliente via Banco (Service)
	user, err := h.Service.AuthenticateUser(email, password)
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:    "sessao_loja",
			Value:   user.Email,
			Path:    "/",
			Expires: time.Now().Add(24 * time.Hour),
		})
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Falha
	http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
}

// --- DASHBOARD / LOGOUT ---

func (h *AuthHandler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessao_loja")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Busca dados completos
	user, orders, err := h.Service.GetDashboardData(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/logout", http.StatusSeeOther)
		return
	}

	// Struct anônima para passar dados combinados
	data := struct {
		User   interface{}
		Orders interface{}
	}{user, orders}

	RenderTemplate(w, r, "dashboard.html", data)
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "sessao_admin", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "sessao_loja", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
