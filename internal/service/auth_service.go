package service

import (
	"errors"
	"time"

	"github.com/MarcosAndradeV/go-ecommerce/internal/models"
	"github.com/MarcosAndradeV/go-ecommerce/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo *repository.UserRepository
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{Repo: repo}
}

// Registra um cliente novo
func (as *AuthService) RegisterCustomer(name, email, password string) error {
	// 1. Verifica se já existe
	existing, _ := as.Repo.GetUserByEmail(email)
	if existing != nil {
		return errors.New("este e-mail já está cadastrado")
	}

	// 2. Hash da senha (Segurança)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 3. Cria o objeto
	user := models.User{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPass),
		IsAdmin:      false, // Clientes nunca nascem admin
		CreatedAt:    time.Now(),
	}

	// 4. Salva
	return as.Repo.CreateUser(user)
}

// Autentica o usuário (Login)
func (as *AuthService) AuthenticateUser(email, password string) (*models.User, error) {
	// 1. Busca usuário
	user, err := as.Repo.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("usuário ou senha inválidos")
	}

	// 2. Compara Hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("usuário ou senha inválidos")
	}

	return user, nil
}

// Dados para o Dashboard
func (as *AuthService) GetDashboardData(email string) (*models.User, []models.Order, error) {
	user, err := as.Repo.GetUserByEmail(email)
	if err != nil {
		return nil, nil, err
	}

	orders, err := as.Repo.GetOrdersByEmail(email)
	if err != nil {
		// Se der erro ao buscar pedidos, retorna lista vazia, mas não trava o user
		orders = []models.Order{}
	}

	return user, orders, nil
}
