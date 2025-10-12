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
	inventoryRepo "healthcheckProject/internal/repository/inventory"
	"healthcheckProject/internal/service"
)

func main() {
	file, err := os.ReadFile("/etc/inventoryapp/config.yaml")
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

	inventoryController := controller.NewInventoryController(
		service.NewInventoryService(
			inventoryRepo.NewPgInventoryRepo(db),
		),
	)

	router := gin.Default()

	router.GET("/health", healthHandler)

	router.POST("/api/v1/internal/inventory/item/add", inventoryController.AddItems)
	router.POST("/api/v1/internal/inventory/item/reserve", inventoryController.ReserveItems)
	router.POST("/api/v1/internal/inventory/reservation/:reservation_id/cancel", inventoryController.CancelReservation)
	router.GET("/api/v1/internal/inventory/reservation/:reservation_id", inventoryController.GetReservationInfo)

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
