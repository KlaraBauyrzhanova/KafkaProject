package message

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Consumer struct {
	DB     *sqlx.DB
	Logger hclog.Logger
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger := consumer.Logger.With("ConsumeClaim", "operation")
	var mes Message
	for m := range claim.Messages() {
		err := json.Unmarshal(m.Value, &mes)
		if err != nil {
			return err
		}
		logger.Debug("save user from kafka to db")
		fmt.Println((mes))

		_, err = consumer.DB.NamedExec("INSERT INTO mess(id, f_name, l_name) VALUES(:id, :fname, :lname)",
			map[string]interface{}{
				"id":    mes.ID,
				"fname": mes.FName,
				"lname": mes.LName,
			})
		if err != nil {
			logger.Error("failed to insert message to db", "error", err)

			pqErr := err.(*pq.Error)

			if pqErr.Code == "23505" {
				logger.Error(`duplicate key value violates unique constraint "mess_pkey"`, "error")
				err = nil
			} else {
				return err
			}
		}
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", m.Value, m.Timestamp, m.Topic)
		session.MarkMessage(m, "")
	}
	return nil
}
