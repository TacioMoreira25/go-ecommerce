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

func (s *StoreService) CreateProduct(name, desc, img string, price int64, stock int, sizes []string) error {
	product := models.Product{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Description: desc,
		ImageURL:    img,
		Price:       price,
		Stock:       stock,
		Sizes:       sizes,
		CreatedAt:   time.Now(),
	}
	// Assumindo que seu Repo tem CreateProduct (se não, adicione no store_repository)
	return s.Repo.CreateProduct(product)
}

func (s *StoreService) EditProduct(ID primitive.ObjectID, name, desc, img string, price int64, stock int, sizes []string) error {

	existingProduct, err := s.Repo.GetProductByID(ID)
	if err != nil {
		return err
	}

	product := models.Product{
		ID:          ID,
		Name:        name,
		Description: desc,
		ImageURL:    img,
		Price:       price,
		Stock:       stock,
		Sizes:       sizes,
		CreatedAt:   existingProduct.CreatedAt, // Mantém a data original
		UpdatedAt:   time.Now(),                // Atualiza a data de modificação
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

func (s *StoreService) ProcessCartPurchase(userIDStr, customerName, customerEmail, customerAddress, paymentMethod, cardNum, cardCVV string, selectedItems []string) (*models.Order, string, string, error) {
	userID, _ := primitive.ObjectIDFromHex(userIDStr)

	// 1. Buscar Carrinho
	user, err := s.Repo.GetUserWithCart(userID)
	if err != nil {
		return nil, "", "", err
	}

	// 2. Filtrar itens e Calcular Total
	var itemsToBuy []models.OrderItem
	var total int64 = 0

	for _, item := range user.Cart {
		shouldBuy := false
		for _, selID := range selectedItems {
			if item.ProductID.Hex() == selID {
				shouldBuy = true
				break
			}
		}
		if shouldBuy {
			product, err := s.Repo.GetProductByID(item.ProductID)
			if err != nil || product.Stock < item.Quantity {
				return nil, "", "", errors.New("produto " + item.ProductName + " sem estoque")
			}
			itemsToBuy = append(itemsToBuy, item)
			total += item.Price * int64(item.Quantity)
		}
	}

	if len(itemsToBuy) == 0 {
		return nil, "", "", errors.New("nenhum item selecionado")
	}

	// 3. PROCESSAR PAGAMENTO
	status := "PAGO"
	var pixCode, qrCodeImg string

	if paymentMethod == "pix" {
		status = "AGUARDANDO_PAGAMENTO"
		// Gera o PIX
		code, img, err := s.Payment.GeneratePix(total)
		if err != nil {
			return nil, "", "", err
		}
		pixCode = code
		qrCodeImg = img
	} else {
		// Processa Cartão (usa o método renomeado ou antigo)
		err = s.Payment.ProcessPaymentCard(cardNum, customerName, cardCVV, total)
		if err != nil {
			return nil, "", "", err
		}
	}

	// 4. Baixar Estoque e Remover do Carrinho
	for _, item := range itemsToBuy {
		for i := 0; i < item.Quantity; i++ {
			s.Repo.DecrementStock(item.ProductID)
		}
		s.Repo.RemoveItemFromCart(userID, item.ProductID, item.Size)
	}

	// 5. Gerar Pedido
	order := models.Order{
		ID:              primitive.NewObjectID(),
		CustomerName:    customerName,
		CustomerEmail:   customerEmail,
		CustomerAddress: customerAddress,
		Status:          status,
		Total:           total,
		CreatedAt:       time.Now(),
		Items:           itemsToBuy,
	}

	if err := s.Repo.CreateOrder(order); err != nil {
		return nil, "", "", err
	}

	// Retorna dados do PIX (se houver)
	return &order, pixCode, qrCodeImg, nil
}

func (s *StoreService) ConfirmPayment(orderIDStr string) error {
	objID, err := primitive.ObjectIDFromHex(orderIDStr)
	if err != nil {
		return err
	}
	return s.Repo.UpdateOrderStatus(objID, "PAGO")
}

func (s *StoreService) AddProductToCart(userIDStr, productIDStr string, quantity int, size string) error {

	// 1. Converter IDs
	userID, _ := primitive.ObjectIDFromHex(userIDStr)

	// Convert product ID string to ObjectID
	productID, _ := primitive.ObjectIDFromHex(productIDStr)

	// 2. Buscar dados atuais do Produto (Preço, Nome, Imagem)
	product, err := s.Repo.GetProductByID(productID)
	if err != nil {
		return err
	}

	if quantity <= 0 {
		quantity = 1
	}

	// 3. Montar o Item do Carrinho (Reutilizando OrderItem)
	item := models.OrderItem{
		ProductID:   product.ID,
		ProductName: product.Name,
		Price:       product.Price,
		Quantity:    quantity,
		Size:        size,
		ImageURL:    product.ImageURL,
	}

	// 4. Salvar no User
	return s.Repo.AddItemToCart(userID, item)
}

func (s *StoreService) RemoveProductFromCart(userIDStr, productIDStr, size string) error {
	userID, _ := primitive.ObjectIDFromHex(userIDStr)
	productID, _ := primitive.ObjectIDFromHex(productIDStr)

	return s.Repo.RemoveItemFromCart(userID, productID, size)
}

func (s *StoreService) UpdateCartItemQuantity(userIDStr, productIDStr string, quantity int, size string) error {
	if quantity <= 0 {
		return errors.New("quantidade deve ser maior que zero")
	}

	userID, _ := primitive.ObjectIDFromHex(userIDStr)
	productID, _ := primitive.ObjectIDFromHex(productIDStr)

	return s.Repo.UpdateCartItemQuantity(userID, productID, quantity, size)
}

func (s *StoreService) GetUserCart(userIDStr string) (*models.User, float64, error) {
	userID, _ := primitive.ObjectIDFromHex(userIDStr)

	user, err := s.Repo.GetUserWithCart(userID)
	if err != nil {
		return nil, 0, err
	}

	// Calcular Total do Carrinho
	var total int64 = 0
	for _, item := range user.Cart {
		total += item.Price * int64(item.Quantity)
	}

	// Retorna total formatado (float64 para o template)
	return user, float64(total) / 100.0, nil
}

func (s *StoreService) DeleteProduct(idStr string) error {
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return err
	}
	return s.Repo.DeleteProduct(objID)
}
