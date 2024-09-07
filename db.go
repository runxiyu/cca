package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func setupDatabase() error {
	var err error
	if config.Db.Type != "postgres" {
		return errors.New("At the moment, the only supported database type is postgres")
	}
	db, err = pgxpool.New(context.Background(), config.Db.Conn)
	return err
}
