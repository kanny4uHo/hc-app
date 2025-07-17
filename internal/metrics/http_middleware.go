package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type HttpMiddleware struct {
	RequestMetrics *prometheus.CounterVec
	LatencyMetrics *prometheus.HistogramVec
}

func (m HttpMiddleware) Handler(ctx *gin.Context) {
	start := time.Now()
	path := ctx.FullPath()
	method := ctx.Request.Method

	ctx.Next()

	executionLatency := time.Now().Sub(start)
	resultStatus := ctx.Writer.Status()

	m.RequestMetrics.With(prometheus.Labels{
		"method": method,
		"path":   path,
		"status": strconv.Itoa(resultStatus),
	}).Inc()

	m.LatencyMetrics.With(prometheus.Labels{
		"method": method,
		"path":   path,
		"status": strconv.Itoa(resultStatus),
	}).Observe(executionLatency.Seconds())

}
