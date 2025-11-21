package repository

import (
	"context"
	"errors"

	"github.com/MarcosAndradeV/go-ecommerce/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Salva um novo usuário no banco
func (ur *UserRepository) CreateUser(user models.User) error {
	coll := ur.db.Collection("users")
	_, err := coll.InsertOne(context.Background(), user)
	return err
}

// Busca usuário por Email (Usado no Login)
func (ur *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	coll := ur.db.Collection("users")
	var user models.User

	err := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}
	return &user, nil
}

// Busca usuário por ID
func (ur *UserRepository) GetUserByID(id primitive.ObjectID) (*models.User, error) {
	coll := ur.db.Collection("users")
	var user models.User

	err := coll.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}
	return &user, nil
}

// Busca todos os pedidos de um email específico (Para o Dashboard)
func (ur *UserRepository) GetOrdersByEmail(email string) ([]models.Order, error) {
	coll := ur.db.Collection("orders")

	// Filtra onde customer_email é igual ao email do usuário
	// Ordena por data decrescente (mais recentes primeiro)
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := coll.Find(context.Background(), bson.M{"customer_email": email}, opts)
	if err != nil {
		return nil, err
	}

	var orders []models.Order
	err = cursor.All(context.Background(), &orders)
	return orders, err
}
