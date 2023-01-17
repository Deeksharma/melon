package driver

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type DB struct {
	SQL *pgxpool.Pool
}

var dbConn = &DB{}

const maxOpenDbConnections = 10
const maxIdleDbConnection = 5         // how many connections can remain in poll that can remain idle
const maxDbLifetime = 5 * time.Minute // maximum life for a database connection pool

// ConnectSQL creates database pool for postgress
func ConnectSQL(ctx context.Context, dsn string) (*DB, error) { //dsn database connection string
	d, err := NewDatabase(ctx, dsn)
	if err != nil {
		panic(err)
	}
	dbConn.SQL = d
	err = testDb(ctx, d)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// testDb tests the db connection by pinging
func testDb(ctx context.Context, d *pgxpool.Pool) error {
	err := d.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

// NewDatabase returns a pool object with required config for the application
func NewDatabase(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	// setting the config
	config.MaxConns = maxOpenDbConnections
	config.MaxConnIdleTime = maxIdleDbConnection
	config.MaxConnLifetime = maxDbLifetime

	conn, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(ctx); err != nil {
		return nil, err
	}
	return conn, nil

}
