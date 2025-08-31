package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"healthcheckProject/internal/api/middlewares"
	"healthcheckProject/internal/controller"
	"healthcheckProject/internal/gateway"
	"healthcheckProject/internal/repository"
	"healthcheckProject/internal/repository/httpclient"
	"healthcheckProject/internal/repository/notification"
	"healthcheckProject/internal/service"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
	"healthcheckProject/internal/config"
)

func main() {
	file, err := os.ReadFile("/etc/notificationapp/config.yaml")
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

	kafkaOrderIsPaidReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               appConfig.RedpandaBroker.Addresses,
		GroupID:               appConfig.RedpandaBroker.ConsumerGroup,
		Topic:                 appConfig.RedpandaBroker.OrderIsPaidTopic,
		WatchPartitionChanges: true,
		Logger:                gateway.KafkaLogger{},
	})

	kafkaOrderPaymentFailedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               appConfig.RedpandaBroker.Addresses,
		GroupID:               appConfig.RedpandaBroker.ConsumerGroup,
		Topic:                 appConfig.RedpandaBroker.OrderPaymentFailedTopic,
		WatchPartitionChanges: true,
		Logger:                gateway.KafkaLogger{},
	})

	notificationRepo := notification.NewPgNotificationRepo(db)

	userServiceRepo := repository.NewUserServiceRepo(
		httpclient.NewUserClient(
			appConfig.UserService.URL,
			&http.Client{Timeout: time.Second},
		),
	)

	httpOrderGate := gateway.NewHttpOrderGate(
		httpclient.NewOrderClient(
			appConfig.OrderService.URL,
			&http.Client{Timeout: time.Second},
		),
	)

	notificationController := controller.NewNotificationController(
		service.NewNotificationService(
			notificationRepo,
			userServiceRepo,
			httpOrderGate,
		),
	)

	router := gin.Default()
	router.Use(middlewares.RequireUser)

	router.GET("/api/v1/notification/list", notificationController.GetNotificationList)

	ctx, cancelFunc := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}

	go func() {
		wg.Add(1)
		defer wg.Done()

		err2 := notificationController.ConsumeOrderIdPaid(ctx, kafkaOrderIsPaidReader)
		if err2 != nil {
			log.Printf("finished cosuming order is paid, err %s", err2)
		}
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()

		err2 := notificationController.ConsumeOrderPaymentIsFailed(ctx, kafkaOrderPaymentFailedReader)
		if err2 != nil {
			log.Printf("finished cosuming order is paid, err %s", err2)
		}
	}()

	log.Println("Listening on :8000")
	err = router.Run(":8000")

	log.Printf("shutting down gracefully")

	cancelFunc()

	wg.Wait()
}
