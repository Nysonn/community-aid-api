package db

import (
	"database/sql"
	"log"

	"community-aid-api/internal/config"

	_ "github.com/lib/pq"
)

func InitDB(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to reach database: %v", err)
	}

	log.Println("database connection established")
	return db
}
