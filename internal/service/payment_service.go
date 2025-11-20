package service

import (
	"errors"
	"strings"
)

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// ProcessPayment simula uma transação com operadora de cartão
func (s *PaymentService) ProcessPayment(cardNumber, cardName, cvv string, amount int64) error {
	// 1. Validação básica
	cleanNum := strings.ReplaceAll(cardNumber, " ", "")
	
	if len(cleanNum) < 16 {
		return errors.New("número de cartão inválido")
	}

	// 2. Simulação de Recusa
	// Regra: Se o cartão terminar em "0000", simulamos recusa do banco
	if strings.HasSuffix(cleanNum, "0000") {
		return errors.New("transação recusada pela operadora (Saldo Insuficiente)")
	}

	// Sucesso
	return nil
}