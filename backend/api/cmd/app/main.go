package main

import (
	"BlessedApi/internal/app"
	"BlessedApi/internal/service"
	"BlessedApi/internal/logger"
	"database/sql"
)

func main() {
	// Инициализируем подключение к базе данных
	connStr := "postgres://postgres:postgres@localhost:5432/blessed?sslmode=disable"
	if err := service.InitDB(connStr); err != nil {
		logger.Fatal("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal("Failed to close database: %v", err)
		}
	}()

	app.Start()
}
