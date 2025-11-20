package repository

import (
	"context"
	"errors"

	"github.com/MarcosAndradeV/go-ecommerce/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Struct que segura a conexão com o banco
type StoreRepository struct {
	db *mongo.Database
}

// Construtor para injetar a dependência do banco
func NewStoreRepository(db *mongo.Database) *StoreRepository {
	return &StoreRepository{db: db}
}

// ---------------------------------------------------------
// MÉTODOS DE PRODUTOS
// ---------------------------------------------------------

// GetAllProducts: Lista todos os produtos para a Home e Admin
func (r *StoreRepository) GetAllProducts() ([]models.Product, error) {
	coll := r.db.Collection("products")

	// Busca sem filtro (todos)
	cursor, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var products []models.Product
	err = cursor.All(context.Background(), &products)
	return products, err
}

// GetProductByID: Busca detalhes de um produto específico
func (r *StoreRepository) GetProductByID(id primitive.ObjectID) (*models.Product, error) {
	coll := r.db.Collection("products")

	var product models.Product
	err := coll.FindOne(context.Background(), bson.M{"_id": id}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// CreateProduct: Usado pelo Admin para cadastrar novos itens
func (r *StoreRepository) CreateProduct(product models.Product) error {
	coll := r.db.Collection("products")
	_, err := coll.InsertOne(context.Background(), product)
	return err
}

// DecrementStock: Baixa o estoque de forma atômica e segura
func (r *StoreRepository) DecrementStock(id primitive.ObjectID) error {
	coll := r.db.Collection("products")

	// O filtro é o segredo: Só atualiza SE o ID bater E se o estoque for maior que 0.
	// Isso impede que o estoque fique negativo se dois clientes comprarem ao mesmo tempo.
	filter := bson.M{"_id": id, "stock": bson.M{"$gt": 0}}
	update := bson.M{"$inc": bson.M{"stock": -1}}

	result, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	// Se nenhum documento foi modificado, significa que o estoque já era 0
	// ou o produto não existe.
	if result.ModifiedCount == 0 {
		return errors.New("estoque insuficiente para realizar a compra")
	}

	return nil
}

// ---------------------------------------------------------
// MÉTODOS DE PEDIDOS
// ---------------------------------------------------------

// CreateOrder: Salva o pedido finalizado no banco
func (r *StoreRepository) CreateOrder(order models.Order) error {
	coll := r.db.Collection("orders")
	_, err := coll.InsertOne(context.Background(), order)
	return err
}
