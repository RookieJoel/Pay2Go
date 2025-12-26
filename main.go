package main

import (
	"Pay2Go/adapter"
	"Pay2Go/entities"
	"Pay2Go/repositories"
	"Pay2Go/usecases"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {

	//Load env
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}

	app := fiber.New()

	port := os.Getenv("port")
	host := os.Getenv("host")
	user := os.Getenv("user")
	pass := os.Getenv("pass")
	dbname := os.Getenv("dbname")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Enable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatal("failed to connect database")
	}

	db.AutoMigrate(&entities.Transaction{})
	repo := repositories.NewGormTransactionRepository(db)
	usecase := usecases.NewTransactionService(repo)
	handler := adapter.NewHTTPHandler(usecase)

	app.Post("/transactions", handler.CreateTransaction)
	app.Get("/transactions/:id", handler.GetTransactionByID)
	app.Put("/transactions/:id", handler.UpdateTransaction)

	app.Listen(":8080")
	fmt.Println("Server is running on port", port)
}
