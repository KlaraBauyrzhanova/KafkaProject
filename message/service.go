package message

import (
	"encoding/json"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type Service struct {
	store  *Store
	db     *sqlx.DB
	sar    sarama.SyncProducer
	logger hclog.Logger
}

func NewService(db *sqlx.DB, logger hclog.Logger, e *echo.Echo, s sarama.SyncProducer, store *Store) *echo.Echo {
	m := Service{
		store:  store,
		db:     db,
		sar:    s,
		logger: logger,
	}
	e.POST("/", m.sendMessage)
	e.GET("/:id", m.getMessageByID)
	return e
}

// sendMessage creates handler of sendMessage
func (s *Service) sendMessage(c echo.Context) error {
	logger := s.logger.With("handler", "sendMessage")

	mess := &Message{}
	if err := c.Bind(mess); err != nil {
		logger.Error("failed to bind author", "error", err)
		return c.String(500, err.Error())
	}

	b, err := json.Marshal(&mess)
	if err != nil {
		logger.Error("failed to Marshal json", "error", err)
		return err
	}

	_, _, err = s.sar.SendMessage(&sarama.ProducerMessage{Topic: "test",
		Value: sarama.ByteEncoder(b)})
	if err != nil {
		logger.Error("failed to send message to kafka", "error", err)
		panic(err)
	}
	return nil
}

// getMessageByID creates handler of get message by ID
func (s *Service) getMessageByID(c echo.Context) error {
	logger := s.logger.With("handler", "getMessageByID")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("bad id request", "error", err)
		return err
	}
	m, err := s.store.GetMessageByID(id)
	if err != nil {
		logger.Error("failed to get message from db", "error", err)
		return err
	}
	return c.JSON(200, m)
}
