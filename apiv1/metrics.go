package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	downloadCounter      = prometheus.NewCounter(prometheus.CounterOpts{Name: "mchub_mapdownloads", Help: "count of map downloads"})
	imagedownloadCounter = prometheus.NewCounter(prometheus.CounterOpts{Name: "mchub_imagedownloads", Help: "count of image downloads"})
	uploadCounter        = prometheus.NewCounter(prometheus.CounterOpts{Name: "mchub_uploads", Help: "count of all file uploads"})
	apiCallCounter       = prometheus.NewCounter(prometheus.CounterOpts{Name: "mchub_apicalls", Help: "count of api calls"})
	apilatencyms         = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "mchub_latency_ms",
		Help:    "request latency in miliseconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 20),
	})
	apibackendlatencyms = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "mchub_backend_latency_ms",
		Help:    "database backend latency in miliseconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 20),
	})
)

func registerMetrics(mux *http.ServeMux) {
	mux.Handle("/metrics", prometheus.Handler())
}

func init() {
	prometheus.MustRegister(downloadCounter)
	prometheus.MustRegister(imagedownloadCounter)
	prometheus.MustRegister(uploadCounter)
	prometheus.MustRegister(apiCallCounter)
	prometheus.MustRegister(apilatencyms)
	prometheus.MustRegister(apibackendlatencyms)
}

func apiCounter(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		apiCallCounter.Add(1)
		fn(rw, r)
		defer func() {
			l := time.Since(start)
			ms := float64(l.Nanoseconds() * 1000)
			apilatencyms.Observe(ms)
		}()
	}
}
