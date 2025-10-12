package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"

	"healthcheckProject/internal/config"
	"healthcheckProject/internal/controller"
	deliveryRepo "healthcheckProject/internal/repository/delivery"
	"healthcheckProject/internal/service"
)

func main() {
	file, err := os.ReadFile("/etc/deliveryapp/config.yaml")
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

	deliveryController := controller.NewDeliveryController(
		service.NewDeliveryService(
			deliveryRepo.NewPgDeliveryRepository(db),
		),
	)

	router := gin.Default()

	router.GET("/health", healthHandler)

	router.POST("/api/v1/internal/delivery/apply", deliveryController.ApplyCourierForOrder)
	router.GET("/api/v1/internal/delivery/list", deliveryController.GetAllDeliveries)
	router.GET("/api/v1/internal/delivery/:delivery_id", deliveryController.GetDeliveryInfo)

	log.Println("Listening on :8000")
	err = router.Run(":8000")

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
