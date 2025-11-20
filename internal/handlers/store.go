package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MarcosAndradeV/go-ecommerce/internal/models"
	"github.com/MarcosAndradeV/go-ecommerce/internal/service"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StoreHandler struct {
	Service *service.StoreService
}

func NewStoreHandler(s *service.StoreService) *StoreHandler {
	return &StoreHandler{Service: s}
}

// --- PÁGINA INICIAL ---

func (h *StoreHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	products, err := h.Service.GetShowcase()
	if err != nil {
		http.Error(w, "Erro ao carregar produtos", 500)
		return
	}
	// CORREÇÃO: Enviando como Mapa para o .Data.Products funcionar
	data := map[string]any{
		"Products": products,
	}
	RenderTemplate(w, r, "index.html", data)
}

// --- DETALHES DO PRODUTO ---

func (h *StoreHandler) ProductDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	product, err := h.Service.GetProductDetails(idStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"Product": product,
	}
	RenderTemplate(w, r, "product.html", data)
}


func (h *StoreHandler) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")
	quantityStr := r.URL.Query().Get("quantity")
	size := r.URL.Query().Get("size")

	quantity, _ := strconv.Atoi(quantityStr)
	if quantity <= 0 {
		quantity = 1
	}

	cookie, _ := r.Cookie("sessao_loja")

	err := h.Service.AddProductToCart(cookie.Value, productID, quantity, size)
	if err != nil {
		http.Redirect(w, r, "/?msg=error_cart", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (h *StoreHandler) RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")
	size := r.URL.Query().Get("size")
	cookie, _ := r.Cookie("sessao_loja")

	err := h.Service.RemoveProductFromCart(cookie.Value, productID, size)
	if err != nil {
		http.Redirect(w, r, "/cart?msg=error_remove", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (h *StoreHandler) ViewCartHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("sessao_loja")
    // 1. Se não tem cookie, manda logar
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    user, total, err := h.Service.GetUserCart(cookie.Value)
    if err != nil {

        http.SetCookie(w, &http.Cookie{
            Name: "sessao_loja",
            MaxAge: -1,
            Path: "/",
        })
        http.Redirect(w, r, "/login?msg=session_expired", http.StatusSeeOther)
        return
    }

    data := map[string]any{
        "Cart":       user.Cart,
        "Total":      total,
        "User":       user,
        "IsLoggedIn": true,
    }
    RenderTemplate(w, r, "cart.html", data)
}
// --- CHECKOUT E COMPRA ---

func (h *StoreHandler) CheckoutPageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("sessao_loja")

	user, _, err := h.Service.GetUserCart(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Se for POST (vindo do carrinho), filtra os itens selecionados
	var selectedItems []string
	if r.Method == http.MethodPost {
		r.ParseForm()
		selectedItems = r.Form["selected_items"]
	}

	// Filtra o carrinho para mostrar apenas os selecionados (se houver seleção)
	// Se não houver seleção (acesso direto ou nada marcado), mostra tudo ou avisa
	var total int64 = 0

	// Se veio do POST mas nada foi selecionado, redireciona pro carrinho
	if r.Method == http.MethodPost && len(selectedItems) == 0 {
		http.Redirect(w, r, "/cart?msg=select_items", http.StatusSeeOther)
		return
	}

	// Se é GET, assume tudo (ou nada, dependendo da regra. Vamos assumir tudo por compatibilidade)
	// Mas o usuário pediu "selecionar quais comprar". Então GET direto pode ser "comprar tudo" ou vazio.
	// Vamos assumir: GET = Tudo. POST = Selecionados.

	finalCart := user.Cart
	if len(selectedItems) > 0 {
		finalCart = nil
		for _, item := range user.Cart {
			for _, selID := range selectedItems {
				if item.ProductID.Hex() == selID {
					finalCart = append(finalCart, item)
					break
				}
			}
		}
	}

	for _, item := range finalCart {
		total += item.Price * int64(item.Quantity)
	}

	data := map[string]any{
		"Cart":  finalCart,
		"Total": float64(total) / 100.0,
		"User":  user,
	}
	RenderTemplate(w, r, "checkout.html", data)
}

func (h *StoreHandler) PurchaseHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	
	// Captura novos campos
	paymentMethod := r.FormValue("payment_method") // "pix" ou "credit_card"
	cardNumber := r.FormValue("card_number")
	cardCVV := r.FormValue("card_cvv")
	selectedItems := r.Form["selected_items"]

	name := r.FormValue("name")
	email := r.FormValue("email")
	address := r.FormValue("address")

	cookie, _ := r.Cookie("sessao_loja")

	// Chama o serviço atualizado
	pixCode, qrCodeImg, err := h.Service.ProcessCartPurchase(cookie.Value, name, email, address, paymentMethod, cardNumber, cardCVV, selectedItems)
	
	if err != nil {
		http.Error(w, "Erro na compra: "+err.Error(), 500)
		return
	}

	// Prepara dados para o template de sucesso
	data := map[string]any{
		"PixCode":     pixCode,
		"QRCodeImage": qrCodeImg,
		"IsPix":       paymentMethod == "pix",
	}

	RenderTemplate(w, r, "success.html", data)
}

// --- ÁREA ADMIN ---

func (h *StoreHandler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	products, _ := h.Service.GetShowcase()

	data := map[string]any{
		"Products": products,
	}
	RenderTemplate(w, r, "admin.html", data)
}

func (h *StoreHandler) AdminCreateProductHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r) { http.Redirect(w, r, "/login", http.StatusSeeOther); return }

	name := r.FormValue("name")
	desc := r.FormValue("description")
	img := r.FormValue("image_url")
	stock, _ := strconv.Atoi(r.FormValue("stock"))
	priceStr := strings.ReplaceAll(r.FormValue("price"), ",", ".")
	priceFloat, _ := strconv.ParseFloat(priceStr, 64)
	priceInt := int64(priceFloat * 100)

	sizesStr := r.FormValue("sizes")
	var sizes []string
	if sizesStr != "" {
		parts := strings.Split(sizesStr, ",")
		for _, p := range parts {
			sizes = append(sizes, strings.TrimSpace(p))
		}
	}

	h.Service.CreateProduct(name, desc, img, priceInt, stock, sizes)
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (h *StoreHandler) EditProductFormHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "product_id")
	if !CheckAuth(r) { http.Redirect(w, r, "/login", http.StatusSeeOther); return }

	product, err := h.Service.GetProductDetails(idStr)
	if err != nil { http.Redirect(w, r, "/", http.StatusSeeOther); return }

	data := map[string]any{ "Product": product }
	RenderTemplate(w, r, "edit.html", data)
}

func (h *StoreHandler) EditProductHandler(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r) { http.Redirect(w, r, "/login", http.StatusSeeOther); return }

	idStr := r.FormValue("id")
	id, _ := primitive.ObjectIDFromHex(idStr)

	name := r.FormValue("name")
	desc := r.FormValue("description")
	img := r.FormValue("image_url")
	stock, _ := strconv.Atoi(r.FormValue("stock"))
	priceStr := strings.ReplaceAll(r.FormValue("price"), ",", ".")
	priceFloat, _ := strconv.ParseFloat(priceStr, 64)
	priceInt := int64(priceFloat * 100)

	sizesStr := r.FormValue("sizes")
	var sizes []string
	if sizesStr != "" {
		parts := strings.Split(sizesStr, ",")
		for _, p := range parts {
			sizes = append(sizes, strings.TrimSpace(p))
		}
	}

	h.Service.EditProduct(id, name, desc, img, priceInt, stock, sizes)

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (h *StoreHandler) PaymentPageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("sessao_loja")
	user, _, err := h.Service.GetUserCart(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	selectedItems := r.Form["selected_items"]
	name := r.FormValue("name")
	email := r.FormValue("email")
	address := r.FormValue("address")
	city := r.FormValue("city")
	zip := r.FormValue("zip")

	fullAddress := address + ", " + city + " - " + zip

	if len(selectedItems) == 0 {
		http.Redirect(w, r, "/cart?msg=select_items", http.StatusSeeOther)
		return
	}

	// Recalcular total dos itens selecionados
	var total int64 = 0
	var itemsToBuy []models.OrderItem

	for _, item := range user.Cart {
		for _, selID := range selectedItems {
			if item.ProductID.Hex() == selID {
				itemsToBuy = append(itemsToBuy, item)
				total += item.Price * int64(item.Quantity)
				break
			}
		}
	}

	data := map[string]any{
		"Items":         itemsToBuy,
		"Total":         float64(total) / 100.0,
		"SelectedItems": selectedItems,
		"Shipping": map[string]string{
			"Name":    name,
			"Email":   email,
			"Address": fullAddress,
		},
	}

	RenderTemplate(w, r, "payment.html", data)
}
