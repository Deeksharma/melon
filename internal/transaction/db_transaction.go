package transaction

import (
	"context"
	"fmt"
	_ "github.com/lib/pq" // Anonymously import the driver package
	"log"
	"melon/pkg/driver"
	"time"
)

type PostgresTransactionLogger struct {
	events chan<- Event // Write-only channel for sending events
	errors <-chan error // Read-only channel for receiving errors
	db     *driver.DB   // The database access interface
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value, CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key, CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *PostgresTransactionLogger) createTable(tableName string) error {
	_, err := l.db.SQL.Exec(context.TODO(), "CREATE TABLE "+tableName+"(id SERIAL PRIMARY KEY, name TEXT NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (l *PostgresTransactionLogger) verifyTableExists(tableName string) (bool, error) {
	var tableExists bool
	err := l.db.SQL.QueryRow(context.TODO(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", tableName).Scan(&tableExists)
	if err != nil {
		log.Fatal(err)
	}
	return tableExists, nil
}

type PostgresDBParams struct {
	DbName   string
	Host     string
	Port     string
	User     string
	Password string
}

func NewPostgresTransactionLogger(config PostgresDBParams) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s",
		config.Host, config.Port, config.DbName, config.User, config.Password)

	db, err := driver.ConnectSQL(context.TODO(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	logger := &PostgresTransactionLogger{db: db}
	var tableName = "transactions" // TODO: get this from config
	exists, err := logger.verifyTableExists(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to verify table exists: %w", err)
	}
	if !exists {
		if err = logger.createTable(tableName); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}
	return logger, nil
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16) // Make an events channel
	l.events = events
	errors := make(chan error, 1) // Make an errors channel
	l.errors = errors
	go func() {
		query := `INSERT INTO transactions 
			(event_type, key, value, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5)` // The INSERT query
		for e := range events { // Retrieve the next Event
			_, err := l.db.SQL.Exec( // Execute the INSERT query
				context.TODO(),
				query,
				e.EventType, e.Key, e.Value, e.CreatedAt, e.UpdatedAt)
			if err != nil {
				fmt.Println(err, "92")
				errors <- err
			}
		}
	}()
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)    // An unbuffered events channel
	outError := make(chan error, 1) // A buffered errors channel
	go func() {
		defer close(outEvent) // Close the channels when the
		defer close(outError) // goroutine ends
		query := `SELECT id, event_type, key, value, created_at, updated_at FROM transactions
          ORDER BY id`
		rows, err := l.db.SQL.Query(context.TODO(), query) // Run query; get result set
		if err != nil {
			fmt.Println(err, "109")
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}
		defer rows.Close() // This is important!
		e := Event{}       // Create an empty Event
		for rows.Next() {  // Iterate over the rows
			err = rows.Scan(
				&e.Sequence, &e.EventType,
				&e.Key, &e.Value, &e.CreatedAt, &e.UpdatedAt) // Read the values from the row into the Event.
			if err != nil {
				outError <- fmt.Errorf("error reading row: %w", err)
				return
			}
			outEvent <- e // Send e to the channel
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()
	return outEvent, outError
}
