package service

import (
	"errors"
	"time"

	"github.com/MarcosAndradeV/go-ecommerce/internal/models"
	"github.com/MarcosAndradeV/go-ecommerce/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StoreService struct {
	Repo *repository.StoreRepository
}

func NewStoreService(repo *repository.StoreRepository) *StoreService {
	return &StoreService{Repo: repo}
}

func (s *StoreService) CreateProduct(name, desc, img string, price int64, stock int) error {
	product := models.Product{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Description: desc,
		ImageURL:    img,
		Price:       price,
		Stock:       stock,
		CreatedAt:   time.Now(),
	}
	// Assumindo que seu Repo tem CreateProduct (se não, adicione no store_repository)
	return s.Repo.CreateProduct(product)
}

func (s *StoreService) GetShowcase() ([]models.Product, error) {
	return s.Repo.GetAllProducts()
}

func (s *StoreService) GetProductDetails(idStr string) (*models.Product, error) {
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, err
	}
	return s.Repo.GetProductByID(objID)
}

func (s *StoreService) ProcessPurchase(productIDStr, customerName, customerEmail string) error {
	objID, _ := primitive.ObjectIDFromHex(productIDStr)

	product, err := s.Repo.GetProductByID(objID)
	if err != nil || product.Stock <= 0 {
		return errors.New("produto indisponível")
	}

	err = s.Repo.DecrementStock(objID)
	if err != nil {
		return err
	}

	order := models.Order{
		ID:            primitive.NewObjectID(),
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Status:        "PAGO",
		Total:         product.Price,
		CreatedAt:     time.Now(),
		Items: []models.OrderItem{
			{
				ProductID:   product.ID,
				ProductName: product.Name,
				Price:       product.Price,
				Quantity:    1,
			},
		},
	}

	return s.Repo.CreateOrder(order)
}
