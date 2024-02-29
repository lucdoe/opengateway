package databases

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// https://redis.io/docs/connect/clients/go/
var RDB *redis.Client
var DB *gorm.DB

func InitializeRedis() error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	_, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		return err
	}

	return nil
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

// In databases.go
func InitializePostgres() (*gorm.DB, error) {
	dsn := "host=postgres user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/Berlin"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err = MigrateDB(db); err != nil {
		return nil, err
	}
	return db, nil
}
