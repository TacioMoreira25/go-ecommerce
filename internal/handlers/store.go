package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/MarcosAndradeV/go-ecommerce/internal/models"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Home: Lista produtos
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	coll := h.db.DB.Collection("products")
	cursor, _ := coll.Find(context.TODO(), bson.M{})

	var products []models.Product
	cursor.All(context.TODO(), &products)

	renderTemplate(w, r, "index.html", products)
}

// Detalhe do Produto
func (h *Handler) ProductDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	objID, _ := primitive.ObjectIDFromHex(idStr)

	coll := h.db.DB.Collection("products")
	var product models.Product
	coll.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&product)

	renderTemplate(w, r, "product.html", product)
}

// Página de Checkout (GET)
func (h *Handler) CheckoutPageHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("product_id")
	objID, _ := primitive.ObjectIDFromHex(idStr)

	coll := h.db.DB.Collection("products")
	var product models.Product
	coll.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&product)

	renderTemplate(w, r, "checkout.html", product)
}

// Processar Compra (POST)
func (h *Handler) PurchaseHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Dados do Form
	productID, _ := primitive.ObjectIDFromHex(r.FormValue("product_id"))
	name := r.FormValue("name")
	email := r.FormValue("email")

	// 2. Buscar Produto Real no Banco (Segurança de Preço)
	collProds := h.db.DB.Collection("products")
	var product models.Product
	err := collProds.FindOne(context.TODO(), bson.M{"_id": productID}).Decode(&product)

	if err != nil || product.Stock <= 0 {
		http.Error(w, "Produto esgotado ou inválido", http.StatusBadRequest)
		return
	}

	// 3. Criar Pedido (Snapshot)
	order := models.Order{
		ID:            primitive.NewObjectID(),
		CustomerName:  name,
		CustomerEmail: email,
		Status:        "PAGO",
		Total:         product.Price,
		CreatedAt:     time.Now(),
		Items: []models.OrderItem{
			{
				ProductID:   product.ID,
				ProductName: product.Name,
				Price:       product.Price, // Copia preço atual
				Quantity:    1,
			},
		},
	}

	// 4. Transação (Simulada): Salvar Pedido e Baixar Estoque
	collOrders := h.db.DB.Collection("orders")
	collOrders.InsertOne(context.TODO(), order)

	collProds.UpdateOne(context.TODO(),
		bson.M{"_id": productID},
		bson.M{"$inc": bson.M{"stock": -1}},
	)

	renderTemplate(w, r, "success.html", nil)
}
