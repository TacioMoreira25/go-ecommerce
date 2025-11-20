package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/MarcosAndradeV/go-ecommerce/internal/database"
	"github.com/MarcosAndradeV/go-ecommerce/internal/handlers"
	"github.com/MarcosAndradeV/go-ecommerce/internal/repository"
	"github.com/MarcosAndradeV/go-ecommerce/internal/routes" // <--- Importe o novo pacote
	"github.com/MarcosAndradeV/go-ecommerce/internal/service"
)

func main() {
	// 1. Configurações
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()

	// 2. Banco de Dados
	store := database.NewMongoStore("ecommerce_go")
	if err := store.Connect(ctx, os.Getenv("MONGO_URI")); err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}
	defer store.Disconnect(ctx)

	dbInstance := store.DB

	// 3. Camada de Repositório (Injeta Banco)
	userRepo := repository.NewUserRepository(dbInstance)
	storeRepo := repository.NewStoreRepository(dbInstance)

	// 4. Camada de Serviço (Injeta Repositórios)
	authService := service.NewAuthService(userRepo)
	storeService := service.NewStoreService(storeRepo)

	// 5. Camada de Handlers (Injeta Serviços)
	authHandler := handlers.NewAuthHandler(authService)
	storeHandler := handlers.NewStoreHandler(storeService)

	// 6. ROTAS (Injeta Handlers e recebe o Router pronto)
	r := routes.NewRouter(authHandler, storeHandler)

	// 7. Servidor
	log.Println("Servidor rodando em http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
