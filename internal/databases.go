package internal

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// https://redis.io/docs/connect/clients/go/
var RDB *redis.Client

func InitializeRedis() error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		return err
	}

	return nil
}

func InitializePostgres() (*gorm.DB, error) {
	// use them in db connection
	dsn := "host=postgres user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/Berlin"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MigrateDB(db *gorm.DB) error {
	err := db.AutoMigrate(
		&Service{},
		&Endpoint{},
		&Middleware{},
		&MiddlewareConfig{},
		&Path{},
		&Tag{},
		&Protocol{},
		&HTTPHeader{},
		&HTTPMethod{},
	)
	if err != nil {
		return err
	}
	return nil
}
