package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/MarcosAndradeV/go-ecommerce/internal/service"
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

	err := h.Service.ProcessPurchase(idStr, name, email)
	if err != nil {
		http.Error(w, "Erro na compra: "+err.Error(), 400)
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

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}