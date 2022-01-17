package db

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/gabihodoroaga/cloudrun-casbin/config"
)

var db *pgxpool.Pool

// SetupDB create the database connection pool
func SetupDB() error {
	conn, err := pgxpool.Connect(context.Background(), config.GetConfig().ConnString)
	if err != nil {
		return errors.Wrap(err, "unable to connect to database")
	}
	db = conn
	zap.L().Sugar().Infof("database: connection initialized, pool size %d, add pool_max_conns=x in connection dns to change", db.Config().MaxConns)

	return nil
}

// GetDB returns the database connection pool
func GetDB() *pgxpool.Pool {
	return db
}
