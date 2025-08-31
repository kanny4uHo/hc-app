package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v3"
	"healthcheckProject/internal/api/middlewares"
	"healthcheckProject/internal/gateway"
	"healthcheckProject/internal/repository/order"
	"log"
	"net/http"
	"os"

	"healthcheckProject/internal/config"
	"healthcheckProject/internal/controller"
	"healthcheckProject/internal/metrics"
	"healthcheckProject/internal/service"
)

func main() {
	file, err := os.ReadFile("/etc/orderapp/config.yaml")
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

	router := gin.Default()

	httpRequestMetric := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "The total number of processed http requests",
	}, []string{"status", "path", "method"})

	httpRequestLatencyMetric := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_latency_seconds",
		Help: "The latency of http requests in seconds",
	}, []string{"status", "path", "method"})

	kafkaWriter := &kafka.Writer{
		Addr:   kafka.TCP(appConfig.RedpandaBroker.Addresses...),
		Async:  true,
		Logger: gateway.KafkaLogger{},
	}

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:               appConfig.RedpandaBroker.Addresses,
		GroupID:               appConfig.RedpandaBroker.ConsumerGroup,
		Topic:                 appConfig.RedpandaBroker.OrderIsPaidTopic,
		WatchPartitionChanges: true,
		Logger:                gateway.KafkaLogger{},
	})

	router.Use(metrics.HttpMiddleware{
		RequestMetrics: httpRequestMetric,
		LatencyMetrics: httpRequestLatencyMetric,
	}.Handler)

	defer kafkaWriter.Close()

	orderController := controller.NewOrderController(
		service.NewOrderService(
			order.NewPgOrderRepo(db),
			gateway.NewKafkaEventGateway(
				kafkaWriter,
				appConfig.RedpandaBroker.NewOrdersTopic,
				appConfig.RedpandaBroker.NewUsersTopic,
				appConfig.RedpandaBroker.OrderIsPaidTopic,
				appConfig.RedpandaBroker.OrderPaymentFailedTopic,
			),
		),
	)

	apiRouter := router.Group("/api/v1")

	orderRouter := apiRouter.Group("/order").Use(middlewares.RequireUser)
	internalRouter := apiRouter.Group("/internal")

	{
		orderRouter.POST("/create", orderController.CreateOrder)
		internalRouter.GET("/order/:order_id", orderController.GetOrder)
	}

	router.GET("/health", healthHandler)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	ctx, cancelFunc := context.WithCancel(context.Background())
	exitChan := make(chan bool)
	go func() {
		consumeError := orderController.ConsumeOrderIsPaid(ctx, kafkaReader)
		if consumeError != nil {
			log.Printf("failed to consume order: %s", consumeError)
		}

		exitChan <- true
	}()

	log.Println("Listening on :8000")
	err = router.Run(":8000")

	cancelFunc()
	<-exitChan
	log.Println("Shutting down")
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
