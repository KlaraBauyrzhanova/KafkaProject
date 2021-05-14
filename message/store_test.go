package message

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
)

// TestGetMessageByID tests func GetMessageByID
func TestGetMessageByID(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{
		JSONFormat: false,
		Level:      hclog.Debug,
	})

	db, mock, err := sqlmock.New()
	if err != nil {
		logger.Error("failed", "error", err)
	}
	defer db.Close()

	m := Message{}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	rows := sqlmock.NewRows([]string{"id", "f_name", "l_name"}).AddRow(m.ID, m.FName, m.LName)
	mock.ExpectQuery("SELECT (.*) FROM mess WHERE").WithArgs(1).WillReturnRows(rows)

	messageStore := NewStore(sqlxDB, logger)
	if _, err := messageStore.GetMessageByID(1); err != nil {
		t.Errorf("error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	defer sqlxDB.Close()
}
