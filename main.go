package main

import (
	"context"
	"flag"
	"fmt"
	"kaf/message"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	port     = "5432"
	userName = "postgres"
	password = "mypassword"
	host     = "localhost"
	dbname   = "postgres"
)

var (
	brokers  = ""
	version  = ""
	group    = ""
	topics   = ""
	assignor = ""
	oldest   = true
	verbose  = false
)

func init() {
	flag.StringVar(&brokers, "brokers", "172.30.205.30:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&group, "group", "message", "Kafka consumer group definition")
	flag.StringVar(&version, "version", "2.1.1", "Kafka cluster version")
	flag.StringVar(&topics, "topics", "test", "Kafka topics to be consumed, as a comma separated list")
	flag.StringVar(&assignor, "assignor", "range", "Consumer group partition assignment strategy (range, roundrobin, sticky)")
	flag.BoolVar(&oldest, "oldest", true, "Kafka consumer consume initial offset from oldest")
	flag.BoolVar(&verbose, "verbose", false, "Sarama logging")
	flag.Parse()

	if len(brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if len(topics) == 0 {
		panic("no topics given to be consumed, please set the -topics flag")
	}

	if len(group) == 0 {
		panic("no Kafka consumer group defined, please set the -group flag")
	}
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "my-kafka",
		Level: hclog.LevelFromString("DEBUG"),
	})

	dbStr := "postgres://" + userName + ":" + password + "@" + host + ":" + port + "/" + dbname + "?" + "sslmode=disable"
	db, err := sqlx.Connect("postgres", dbStr)
	if err != nil {
		logger.Error("failed to connect to db", "error", err)
		fmt.Println("failed to connect to db")
		return
	}

	m, err := migrate.New(
		"file://migrates",
		dbStr)

	if err != nil {
		logger.Error("failed to migreate", "error", err)
		fmt.Println("failed to make migrate", err)
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("failed to up migrate", "error", err)
		fmt.Println("failed to m.Up() in migrate")
		return
	}

	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	conf.Consumer.Return.Errors = true
	conf.Consumer.Offsets.Initial = sarama.OffsetOldest

	s, err := sarama.NewSyncProducer([]string{"172.30.205.30:9092"}, conf)
	if err != nil {
		logger.Error("failed to creates a new SyncProducer", "error", err)
		panic(err)
	}
	cons := message.Consumer{DB: db, Logger: logger}
	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, conf)
	if err != nil {
		logger.Error("failed to create a new consumer group", "error", err)
		log.Panicf("Error creating consumer group client: %v", err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err = client.Consume(ctx, strings.Split(topics, ","), &cons); err != nil {
				logger.Error("error from consumer", "error", err)
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}

		}
	}()
	logger.Debug("Sarama consumer up and running!...")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	e := echo.New()
	storeMessage := message.NewStore(db, logger)
	message.NewService(db, logger, e, s, storeMessage)

	go func() {
		err := e.Start(":7000")
		if err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start HTTP server", "error", err)
			panic(err)
		}
	}()
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}

	cancel()
	wg.Wait()
	e.Shutdown(context.Background())

	if err = client.Close(); err != nil {
		logger.Error("failed to close client", "error", err)
		log.Panicf("Error closing client: %v", err)
	}

}
