package metrics

import (
	"GoWebFrame/micro/observability"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"time"
)

type ServerMetricsBuilder struct {
	Namespace string
	Subsystem string
}

func (b *ServerMetricsBuilder) Build() grpc.UnaryServerInterceptor {
	addr := observability.GetOutBoundIP()

	reqGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      "active_request_cnt",
		Help:      "当前正在处理的请求数",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   addr,
		},
	}, []string{"service"})

	prometheus.MustRegister(reqGauge)
	errCnt := prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      "active_request_cnt",
		Help:      "当前正在处理的请求数",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   addr,
		},
	}, []string{"service"})

	response := prometheus.NewSummaryVec(prometheus.SummaryOpts{Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      "active_request_cnt",
		Help:      "当前正在处理的请求数",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   addr,
		},
	}, []string{"service"})

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		startTime := time.Now()
		reqGauge.WithLabelValues(info.FullMethod).Add(1)
		defer func() {
			reqGauge.WithLabelValues(info.FullMethod).Add(-1)
			if err != nil {
				errCnt.WithLabelValues(info.FullMethod).Add(1)
			}

			response.WithLabelValues(info.FullMethod).Observe(float64(time.Now().Sub(startTime).Milliseconds()))
		}()

		resp, err = handler(ctx, req)
		return
	}
}
