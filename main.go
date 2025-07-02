package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
	"healthcheckProject/internal/config"
	"healthcheckProject/internal/controller"
	"healthcheckProject/internal/repository"
	"healthcheckProject/internal/service"
)

var started = time.Now()

func main() {
	file, err := os.ReadFile("/etc/userapp/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	appConfig := config.Config{}
	err = yaml.Unmarshal(file, &appConfig)

	if err != nil {
		log.Fatalf("failed to unmarshal config.yaml: %s", err)
	}

	pwdBytes, err := os.ReadFile("/etc/pgsecret/postgres-password")
	if err != nil {
		log.Fatal("failed to read postgres-password from /etc/pgsecret/postgres-password")
	}

	databaseString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		appConfig.Database.Username,
		pwdBytes,
		appConfig.Database.Host,
		appConfig.Database.Port,
		appConfig.Database.DBName,
	)

	router := gin.Default()

	db, err := sql.Open("postgres", databaseString)
	if err != nil {
		log.Fatalf("failed to connect to userdb: %s", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping userdb: %s", err)
	}

	userController := controller.CreateUserController(
		service.NewUserService(repository.NewPgRepo(db)),
	)

	userGroup := router.Group("/user")

	userGroup.POST("", userController.CreateUser)
	userGroup.GET("/:user_id", userController.GetUser)
	userGroup.DELETE("/:user_id", userController.DeleteUser)
	userGroup.PUT("/:user_id", userController.UpdateUser)

	router.GET("/health", healthHandler)

	fmt.Println("Listening on :8000")
	err = router.Run(":8000")

	fmt.Println("Shutting down")
	if err != nil {
		fmt.Println(err.Error())
	}

}

func healthHandler(ctx *gin.Context) {
	if time.Since(started).Seconds() < 5 {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "Service is unavailable",
		})

		return
	}

	ctx.JSON(http.StatusOK, HealthResponse{
		Status: "OK",
		Host:   os.Getenv("HOSTNAME"),
	})
}

type HealthResponse struct {
	Status string `json:"status"`
	Host   string `json:"host"`
}
