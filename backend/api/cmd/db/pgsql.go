package db

import (
	"BlessedApi/pkg/logger"
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	var err error

	dbHost, ok1 := os.LookupEnv("POSTGRES_HOST")
	dbPort, ok2 := os.LookupEnv("POSTGRES_PORT")
	dbUser, ok3 := os.LookupEnv("POSTGRES_USER")
	dbPassword, ok4 := os.LookupEnv("POSTGRES_PASSWORD")
	dbName, ok5 := os.LookupEnv("POSTGRES_DB")
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		logger.Fatal("%v", errors.New("unable to get database connection parameters from environment"))
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("%v", err)
	}
}
