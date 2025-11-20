package service

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/skip2/go-qrcode"
)

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// ProcessPaymentCard simula cartão
func (s *PaymentService) ProcessPaymentCard(cardNumber, cardName, cvv string, amount int64) error {
	cleanNum := strings.ReplaceAll(cardNumber, " ", "")
	if len(cleanNum) < 16 {
		return errors.New("número de cartão inválido")
	}
	if strings.HasSuffix(cleanNum, "0000") {
		return errors.New("transação recusada pela operadora (Saldo Insuficiente)")
	}
	return nil
}

// GeneratePix gera o código e a imagem QR Code em Base64
func (s *PaymentService) GeneratePix(amount int64) (string, string, error) {
	// 1. Código "Copia e Cola" Simulado (Formato padrão EMV)
	// Num cenário real, isso viria da API do Banco (PSP)
	pixPayload := "00020126330014BR.GOV.BCB.PIX011112345678900520400005303986540510.005802BR5913Olecram Shop6008Sao Paulo62070503***6304B6CD"

	// 2. Gerar a imagem do QR Code (256x256)
	png, err := qrcode.Encode(pixPayload, qrcode.Medium, 256)
	if err != nil {
		return "", "", err
	}

	// 3. Converter para Base64 para exibir no HTML
	base64Img := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)

	return pixPayload, base64Img, nil
}