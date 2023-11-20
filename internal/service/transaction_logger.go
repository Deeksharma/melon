package service

import (
	"fmt"
	"melon/internal/transaction"
)

var logger transaction.TransactionLogger

func InitializeTransactionLog() error {
	var err error
	logger, err = transaction.NewFileTransactionLogger("transaction.log")
	// logger, err = NewPostgresTransactionLogger("localhost") // TODO test it by runnin postgeryy
	if err != nil {
		return fmt.Errorf("failed to create event logger: %w", err)
	}

	events, errors := logger.ReadEvents()
	e, ok := transaction.Event{}, true
	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case transaction.EventDelete:
				err = Delete(e.Key)
			case transaction.EventPut:
				err = Put(e.Key, e.Value)
			}
		}
	}
	logger.Run()
	return err
}

func WritePut(key, value string) {
	logger.WritePut(key, value)
}

func WriteDelete(key string) {
	logger.WriteDelete(key)
}
