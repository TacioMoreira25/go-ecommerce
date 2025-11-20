package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// PageData: O envelope padrão enviado para o HTML
type PageData struct {
	IsAdmin bool
	Data    interface{}
	Error   string
}

// CheckAuth: Verifica se o cookie de ADMIN existe
func CheckAuth(r *http.Request) bool {
	cookie, err := r.Cookie("sessao_admin")
	return err == nil && cookie.Value == "true"
}

// RenderTemplate: Função auxiliar para renderizar HTML com o layout base
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, data interface{}) {
	isAdmin := CheckAuth(r)

	pageData := PageData{
		IsAdmin: isAdmin,
		Data:    data,
	}

	// Ajuste os caminhos se necessário
	layout := filepath.Join("templates", "layouts", "base.html")
	view := filepath.Join("templates", tmplName)

	// ParseFiles junta o layout com a view
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
