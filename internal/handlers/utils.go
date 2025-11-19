package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/MarcosAndradeV/go-ecommerce/internal/database"
	"go.mongodb.org/mongo-driver/mongo"
)

// O envelope que vai para o HTML
type PageData struct {
	IsAdmin bool
	Data    interface{}
	Error   string // Opcional, para mensagens de erro
}

// Verifica se o cookie de admin existe
func CheckAuth(r *http.Request) bool {
	_, err := r.Cookie("sessao_admin")
	return err == nil
}

// Renderiza juntando base.html + arquivo da página
func renderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, data interface{}) {
	isAdmin := CheckAuth(r)

	pageData := PageData{
		IsAdmin: isAdmin,
		Data:    data,
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

type Handler struct {
	db *database.MongoStore
}

func (h *Handler) GetCollection(name string) *mongo.Collection {
	return h.db.DB.Collection(name)
}
