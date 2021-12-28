package auth

import (
	"context"
	pgxp "github.com/jackc/pgx/v4/pgxpool"
	"os"
)

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func OpenDB() *pgxp.Pool {
	dbpool, err := pgxp.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	CheckError(err)

	return dbpool
}

func CloseDB(dbpool *pgxp.Pool) {
	dbpool.Close()
}
