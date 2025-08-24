package metrics

import "time"

type RequestMetric struct {
	Duration   time.Duration
	StatusCode int
}

type NetworkMetric struct {
	BytesSent int64
	BytesRecv int64
}

type CheckMetric struct {
	Id     string
	status bool
}
