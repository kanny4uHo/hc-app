package metrics

type prometheusMetrics struct {
}

var _ Metrics = (*prometheusMetrics)(nil)
