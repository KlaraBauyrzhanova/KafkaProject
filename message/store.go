package message

import (
	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	db     *sqlx.DB
	logger hclog.Logger
}

func NewStore(db *sqlx.DB, logger hclog.Logger) *Store {
	return &Store{
		db:     db,
		logger: logger,
	}
}

// GetMessageByID gets message by ID
func (s *Store) GetMessageByID(id int) (mess Message, err error) {
	logger := s.logger.With("operation", "GetMessageByID")
	logger.Debug("get message from db")

	a := Message{ID: id}
	row, err := s.db.NamedQuery(`SELECT * FROM mess WHERE id=:id`, a)
	if err != nil {
		logger.Error("failed to select message by ID from db", "error", err)
		return
	}
	if !row.Next() {
		logger.Debug("did not find the message")
		return
	}
	err = row.StructScan(&mess)
	if err != nil {
		logger.Error("failed to scan the message", "error", err)
		return
	}
	return
}
