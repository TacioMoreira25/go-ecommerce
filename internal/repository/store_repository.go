package repository

import (
	"context"
	"errors"
	"time"

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

// EditProduct
func (r *StoreRepository) EditProduct(ID primitive.ObjectID, product models.Product) error {
	coll := r.db.Collection("products")
	_, err := coll.UpdateByID(context.Background(), ID, product)
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

func (r *StoreRepository) AddItemToCart(userID primitive.ObjectID, item models.OrderItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userColl := r.db.Collection("users")

	filter := bson.M{"_id": userID}
	update := bson.M{"$push": bson.M{"cart": item}}

	_, err := userColl.UpdateOne(ctx, filter, update)
	return err
}

func (r *StoreRepository) RemoveItemFromCart(userID primitive.ObjectID, productID primitive.ObjectID, size string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userColl := r.db.Collection("users")

	// Remove o item do array 'cart' onde o product_id bate E o tamanho bate (se tiver tamanho)
	// Se size for vazio, removemos onde product_id bate e size é vazio ou não existe?
	// Simplificação: Removemos exatamente o que veio.

	filter := bson.M{"_id": userID}

	// $pull remove todos os itens que dão match no critério
	pullFilter := bson.M{"product_id": productID}
	if size != "" {
		pullFilter["size"] = size
	} else {
		// Se não tem tamanho, talvez devêssemos remover itens sem tamanho?
		// Ou remover qualquer um desse produto?
		// Vamos assumir que se size é vazio, removemos itens onde size é vazio ou null
		// Mas o $pull simples já funciona se passarmos o objeto exato.
		// Vamos usar o filtro composto.
	}

	// Se size for vazio, o pullFilter fica só com product_id, o que removeria TODOS os tamanhos desse produto.
	// Isso é perigoso se o usuário tiver P e M e clicar remover no que não tem tamanho (se existir).
	// Mas se o produto tem tamanho, o link sempre manda o tamanho.
	// Se o produto NÃO tem tamanho, size vem vazio.
	// Então se size == "", removemos itens onde size == "" ou não existe.

	if size == "" {
		// Remove itens onde product_id bate E (size não existe OU size é vazio)
		// MongoDB query complexa para array element match.
		// Vamos simplificar: Se size veio vazio, removemos pelo ID.
		// Se isso apagar todos, é o comportamento esperado para produtos sem variação.
		// O problema é se tivermos um produto com variação e o request vier sem size.
	} else {
		// Se tem size, removemos só aquele size.
	}

	// Melhor abordagem: Sempre tentar dar match no size se ele foi fornecido.
	// Se não foi fornecido, removemos pelo ID (todos).

	var update bson.M
	if size != "" {
		update = bson.M{"$pull": bson.M{"cart": bson.M{"product_id": productID, "size": size}}}
	} else {
		// Se não tem size, removemos onde product_id bate.
		// CUIDADO: Isso remove P e M se o request vier sem size.
		// Mas o request do cart.html sempre manda o size se ele existir no item.
		update = bson.M{"$pull": bson.M{"cart": bson.M{"product_id": productID}}}
	}

	_, err := userColl.UpdateOne(ctx, filter, update)
	return err
}

func (r *StoreRepository) GetUserWithCart(userID primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userColl := r.db.Collection("users")

	var user models.User
	err := userColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	return &user, err
}
