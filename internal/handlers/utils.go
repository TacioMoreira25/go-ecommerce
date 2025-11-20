package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type PageData struct {
	IsAdmin    bool
	IsLoggedIn bool // <--- NOVO CAMPO
	Data       any
	Error      string
}

// Verifica cookie de ADMIN
func CheckAuth(r *http.Request) bool {
	cookie, err := r.Cookie("sessao_admin")
	return err == nil && cookie.Value == "true"
}

// Verifica cookie de CLIENTE COMUM
func CheckUserLogin(r *http.Request) bool {
	_, err := r.Cookie("sessao_loja")
	return err == nil
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, data any) {
	isAdmin := CheckAuth(r)
	isUser := CheckUserLogin(r)

	// Se for Admin ou Usuário, consideramos como Logado
	isLoggedIn := isAdmin || isUser

	pageData := PageData{
		IsAdmin:    isAdmin,
		IsLoggedIn: isLoggedIn, // Passamos essa info pro HTML agora
		Data:       data,
	}

	layout := filepath.Join("templates", "layouts", "base.html")
	view := filepath.Join("templates", tmplName)

	tmpl, err := template.ParseFiles(layout, view)
	if err != nil {
		http.Error(w, "Erro interno (Template): "+err.Error(), 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", pageData)
	if err != nil {
		http.Error(w, "Erro renderização: "+err.Error(), 500)
	}
}
