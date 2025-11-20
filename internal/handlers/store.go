package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MarcosAndradeV/go-ecommerce/internal/service"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StoreHandler struct {
	Service *service.StoreService
}

// Construtor
func NewStoreHandler(s *service.StoreService) *StoreHandler {
	return &StoreHandler{Service: s}
}

// --- ÁREA PÚBLICA ---

func (h *StoreHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	products, err := h.Service.GetShowcase()
	if err != nil {
		http.Error(w, "Erro ao carregar produtos", 500)
		return
	}
	RenderTemplate(w, r, "index.html", products)
}

func (h *StoreHandler) EditProductFormHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "product_id")

	if !CheckAuth(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	product, err := h.Service.GetProductDetails(idStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	RenderTemplate(w, r, "edit.html", product)
}

func (h *StoreHandler) EditProductHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	desc := r.FormValue("description")
	img := r.FormValue("image_url")
	idStr := r.FormValue("id")
	stock, _ := strconv.Atoi(r.FormValue("stock"))

	// Parse do preço (10.50 -> 1050)
	priceStr := strings.ReplaceAll(r.FormValue("price"), ",", ".")
	priceFloat, _ := strconv.ParseFloat(priceStr, 64)
	priceInt := int64(priceFloat * 100)

	// Chama Service para criar (Você precisará adicionar CreateProduct no StoreService se não tiver)
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Erro ao criar: "+err.Error(), 500)
		return
	}
	err = h.Service.EditProduct(id,
		name, desc, img, priceInt, stock,
	)
	if err != nil {
		http.Error(w, "Erro ao criar: "+err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (h *StoreHandler) ProductDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	product, err := h.Service.GetProductDetails(idStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	RenderTemplate(w, r, "product.html", product)
}

func (h *StoreHandler) CheckoutPageHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("product_id")
	product, err := h.Service.GetProductDetails(idStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	RenderTemplate(w, r, "checkout.html", product)
}

func (h *StoreHandler) PurchaseHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("product_id")
	name := r.FormValue("name")
	email := r.FormValue("email")

	// Novos campos do formulário
	cardNum := r.FormValue("card_number")
	cardCVV := r.FormValue("card_cvv")

	// Passamos tudo para o serviço
	err := h.Service.ProcessPurchase(idStr, name, email, cardNum, cardCVV)

	if err != nil {
		// Se der erro (ex: cartão recusado), voltamos para o checkout com erro
		// Idealmente passariamos a mensagem de erro para o template
		http.Error(w, "Falha na compra: "+err.Error(), 400)
		return
	}

	RenderTemplate(w, r, "success.html", nil)
}

// --- ÁREA ADMIN (Incluída aqui pois usa StoreService) ---

func (h *StoreHandler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Reutiliza a lógica de listar produtos
	products, _ := h.Service.GetShowcase()
	RenderTemplate(w, r, "admin.html", products)
}

func (h *StoreHandler) AdminCreateProductHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	desc := r.FormValue("description")
	img := r.FormValue("image_url")
	stock, _ := strconv.Atoi(r.FormValue("stock"))

	// Parse do preço (10.50 -> 1050)
	priceStr := strings.ReplaceAll(r.FormValue("price"), ",", ".")
	priceFloat, _ := strconv.ParseFloat(priceStr, 64)
	priceInt := int64(priceFloat * 100)

	// Chama Service para criar (Você precisará adicionar CreateProduct no StoreService se não tiver)
	err := h.Service.CreateProduct(name, desc, img, priceInt, stock)
	if err != nil {
		http.Error(w, "Erro ao criar: "+err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
