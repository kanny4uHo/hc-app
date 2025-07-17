package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func GetMetrics(ctx *gin.Context) {
	promhttp.Handler()
}
