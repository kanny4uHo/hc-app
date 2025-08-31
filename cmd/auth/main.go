package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
	"healthcheckProject/internal/config"
	"healthcheckProject/internal/controller"
	"healthcheckProject/internal/metrics"
	"healthcheckProject/internal/repository"
	"healthcheckProject/internal/repository/httpclient"
	"healthcheckProject/internal/service"
)

var started = time.Now()

func main() {
	file, err := os.ReadFile("/etc/authapp/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	appConfig := config.Config{}
	err = yaml.Unmarshal(file, &appConfig)

	if err != nil {
		log.Fatalf("failed to unmarshal config.yaml: %s", err)
	}

	jwtTokenSecret := os.Getenv("jwt-token")

	router := gin.Default()

	httpRequestMetric := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "The total number of processed http requests",
	}, []string{"status", "path", "method"})

	httpRequestLatencyMetric := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_latency_seconds",
		Help: "The latency of http requests in seconds",
	}, []string{"status", "path", "method"})

	router.Use(metrics.HttpMiddleware{
		RequestMetrics: httpRequestMetric,
		LatencyMetrics: httpRequestLatencyMetric,
	}.Handler)

	credRepo := repository.NewUserServiceRepo(
		httpclient.NewUserClient(
			appConfig.UserService.URL,
			&http.Client{Timeout: time.Second},
		),
	)

	authService := service.NewAuthService(credRepo, []byte(jwtTokenSecret))
	authController := controller.NewAuthController(authService)

	apiRouter := router.Group("/api/v1")
	{
		apiRouter.POST("/login", authController.Login)
	}

	internalApiRouter := router.Group("/internal/api/v1")
	{
		internalApiRouter.POST("/register", authController.Register)
		internalApiRouter.GET("/check", authController.AuthCheck)
	}

	router.GET("/health", healthHandler)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
