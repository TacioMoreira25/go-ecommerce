package service

import (
	// "errors"
	// "strings"

)

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// ProcessPaymentCard simula cartão
func (s *PaymentService) ProcessPaymentCard(cardNumber, cardName, cvv string, amount int64) error {
	// cleanNum := strings.ReplaceAll(cardNumber, " ", "")
	// if len(cleanNum) < 16 {
	// 	return errors.New("número de cartão inválido")
	// }
	// if strings.HasSuffix(cleanNum, "0000") {
	// 	return errors.New("transação recusada pela operadora (Saldo Insuficiente)")
	// }
	return nil
}

// GeneratePix gera o código e a imagem QR Code em Base64
func (s *PaymentService) GeneratePix(amount int64) (string, string, error) {
	// 1. Código "Copia e Cola" Simulado (Formato padrão EMV)
	// Num cenário real, isso viria da API do Banco (PSP)
	pixPayload := "00020126330014BR.GOV.BCB.PIX011112345678900520400005303986540510.005802BR5913Olecram Shop6008Sao Paulo62070503***6304B6CD"

	imagePath := "/static/pix-qrcode.png"

	return pixPayload, imagePath, nil
}
