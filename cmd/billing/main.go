package main

import (
	"context"
	"database/sql"
	"fmt"
	"healthcheckProject/internal/gateway"
	"healthcheckProject/internal/repository/httpclient"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v3"
	"healthcheckProject/internal/config"
	"healthcheckProject/internal/controller"
	"healthcheckProject/internal/repository/billing"
	"healthcheckProject/internal/service"
)

func main() {
	file, err := os.ReadFile("/etc/billingapp/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	appConfig := config.Config{}
	err = yaml.Unmarshal(file, &appConfig)

	if err != nil {
		log.Fatalf("failed to unmarshal config.yaml: %s", err)
	}

	databaseString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		appConfig.Database.Username,
		appConfig.Database.Password,
		appConfig.Database.Host,
		appConfig.Database.Port,
		appConfig.Database.DBName,
	)
	db, err := sql.Open("postgres", databaseString)
	if err != nil {
		log.Fatalf("failed to connect to userdb: %s", err)
	}

	defer db.Close()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping userdb: %s", err)
	}

	kafkaWriter := &kafka.Writer{
		Addr:   kafka.TCP(appConfig.RedpandaBroker.Addresses...),
		Async:  true,
		Logger: gateway.KafkaLogger{},
	}

	billingController := controller.NewBillingController(
		service.NewBillingService(
			billing.NewUserAccountRepoImpl(db),
			gateway.NewHttpOrderGate(
				httpclient.NewOrderClient(
					appConfig.OrderService.URL,
					&http.Client{Timeout: time.Second},
				),
			),
			gateway.NewKafkaEventGateway(
				kafkaWriter,
				appConfig.RedpandaBroker.NewOrdersTopic,
				appConfig.RedpandaBroker.NewUsersTopic,
				appConfig.RedpandaBroker.OrderIsPaidTopic,
				appConfig.RedpandaBroker.OrderPaymentFailedTopic,
			),
		),
	)

	router := gin.Default()

	router.GET("/health", healthHandler)

	router.POST("/api/v1/billing/money/credit", billingController.CreditMoney)
	router.POST("/api/v1/billing/money/withdraw", billingController.WithdrawMoney)
	router.GET("/api/v1/billing/account/:user_id", billingController.GetUserAccountInfo)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               appConfig.RedpandaBroker.Addresses,
		GroupID:               appConfig.RedpandaBroker.ConsumerGroup,
		Topic:                 appConfig.RedpandaBroker.NewUsersTopic,
		WatchPartitionChanges: true,
	})

	newOrdersReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               appConfig.RedpandaBroker.Addresses,
		GroupID:               appConfig.RedpandaBroker.ConsumerGroup,
		Topic:                 appConfig.RedpandaBroker.NewOrdersTopic,
		WatchPartitionChanges: true,
		Logger:                gateway.KafkaLogger{},
	})

	defer func() {
		reader.Close()
	}()

	redpandaCtx, cancelConsumerContext := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}

	go func() {
		wg.Add(1)
		defer wg.Done()

		consumeErr := billingController.Consume(redpandaCtx, reader)
		if consumeErr != nil {
			log.Printf("failed to consume: %s", consumeErr)
		}

		defer func() {
			if r := recover(); r != nil {
				log.Printf("cosumer recovered from panic: %s", r)
			}
		}()

		log.Printf("Consumer exited")
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()

		consumeErr := billingController.ConsumeNewOrders(redpandaCtx, newOrdersReader)
		if consumeErr != nil {
			log.Printf("failed to consume: %s", consumeErr)
		}

		defer func() {
			if r := recover(); r != nil {
				log.Printf("cosumer recovered from panic: %s", r)
			}
		}()

		log.Printf("New orders consumer exited")
	}()

	log.Println("Listening on :8000")
	err = router.Run(":8000")

	log.Println("Shutting down")
	cancelConsumerContext()
	wg.Wait()

	if err != nil {
		log.Println(err.Error())
	}

}

func healthHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, HealthResponse{
		Status: "OK",
		Host:   os.Getenv("HOSTNAME"),
	})
}

type HealthResponse struct {
	Status string `json:"status"`
	Host   string `json:"host"`
}
