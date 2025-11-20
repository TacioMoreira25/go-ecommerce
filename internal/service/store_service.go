package service

import (
	"errors"
	"time"

	"github.com/MarcosAndradeV/go-ecommerce/internal/models"
	"github.com/MarcosAndradeV/go-ecommerce/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type StoreService struct {
	Repo    *repository.StoreRepository
	Payment *PaymentService
}

func NewStoreService(repo *repository.StoreRepository, payment *PaymentService) *StoreService {
	return &StoreService{
		Repo:    repo,
		Payment: payment,
	}
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

func (s *StoreService) EditProduct(ID primitive.ObjectID, name, desc, img string, price int64, stock int) error {
	product := models.Product{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Description: desc,
		ImageURL:    img,
		Price:       price,
		Stock:       stock,
		CreatedAt:   time.Now(),
	}
	return s.Repo.EditProduct(ID, product)
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

func (s *StoreService) ProcessPurchase(productIDStr, customerName, customerEmail, cardNum, cardCVV string) error {
	objID, _ := primitive.ObjectIDFromHex(productIDStr)

	// 1. Buscar Produto e Validar Estoque
	product, err := s.Repo.GetProductByID(objID)
	if err != nil || product.Stock <= 0 {
		return errors.New("produto indisponível ou não encontrado")
	}

	// 2. PROCESSAR PAGAMENTO (Antes de mexer no estoque)
	// Se falhar aqui, retornamos o erro e nada acontece no banco
	err = s.Payment.ProcessPayment(cardNum, customerName, cardCVV, product.Price)
	if err != nil {
		return err // Ex: "transação recusada"
	}

	// 3. Baixar Estoque (Só chega aqui se pagou)
	err = s.Repo.DecrementStock(objID)
	if err != nil {
		// Num cenário real, aqui faríamos o estorno do pagamento (Reverse)
		return errors.New("erro ao atualizar estoque, compra cancelada")
	}

	// 4. Gerar Pedido
	order := models.Order{
		ID:            primitive.NewObjectID(),
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Status:        "PAGO", // Confirmado
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
