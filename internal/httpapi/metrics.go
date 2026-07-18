package httpapi

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// metrics 轻量指标(Prometheus 文本格式),无第三方依赖。
type metrics struct {
	startedAt time.Time

	requestsTotal atomic.Int64
	inFlight      atomic.Int64
	status2xx     atomic.Int64
	status4xx     atomic.Int64
	status5xx     atomic.Int64
	durationSumMs atomic.Int64

	mu      sync.Mutex
	byRoute map[string]int64

	chartCacheHit  atomic.Int64
	chartCacheMiss atomic.Int64
	aiRequests     atomic.Int64
	aiErrors       atomic.Int64
}

func newMetrics() *metrics {
	return &metrics{startedAt: time.Now(), byRoute: make(map[string]int64)}
}

func (m *metrics) observe(method, path string, status int, dur time.Duration) {
	m.requestsTotal.Add(1)
	m.durationSumMs.Add(dur.Milliseconds())
	switch {
	case status >= 500:
		m.status5xx.Add(1)
	case status >= 400:
		m.status4xx.Add(1)
	default:
		m.status2xx.Add(1)
	}
	key := method + " " + routeKey(path)
	m.mu.Lock()
	m.byRoute[key]++
	m.mu.Unlock()
}

// routeKey 归一化路径,避免高基数标签(slug/id 参数折叠)。
func routeKey(path string) string {
	switch {
	case len(path) > 14 && path[:14] == "/api/v1/books/":
		return "/api/v1/books/{slug}"
	case len(path) > 15 && path[:15] == "/api/v1/famous/":
		return "/api/v1/famous/{id}"
	case len(path) > 14 && path[:14] == "/api/v1/nihai/":
		return "/api/v1/nihai/{section}"
	}
	return path
}

func (m *metrics) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintf(w, "# TYPE ziwei_uptime_seconds gauge\nziwei_uptime_seconds %d\n", int(time.Since(m.startedAt).Seconds()))
		fmt.Fprintf(w, "# TYPE ziwei_http_requests_total counter\nziwei_http_requests_total %d\n", m.requestsTotal.Load())
		fmt.Fprintf(w, "# TYPE ziwei_http_in_flight gauge\nziwei_http_in_flight %d\n", m.inFlight.Load())
		fmt.Fprintf(w, "# TYPE ziwei_http_responses_total counter\n")
		fmt.Fprintf(w, "ziwei_http_responses_total{class=\"2xx\"} %d\n", m.status2xx.Load())
		fmt.Fprintf(w, "ziwei_http_responses_total{class=\"4xx\"} %d\n", m.status4xx.Load())
		fmt.Fprintf(w, "ziwei_http_responses_total{class=\"5xx\"} %d\n", m.status5xx.Load())
		fmt.Fprintf(w, "# TYPE ziwei_http_duration_ms_sum counter\nziwei_http_duration_ms_sum %d\n", m.durationSumMs.Load())
		fmt.Fprintf(w, "# TYPE ziwei_chart_cache_total counter\n")
		fmt.Fprintf(w, "ziwei_chart_cache_total{result=\"hit\"} %d\n", m.chartCacheHit.Load())
		fmt.Fprintf(w, "ziwei_chart_cache_total{result=\"miss\"} %d\n", m.chartCacheMiss.Load())
		fmt.Fprintf(w, "# TYPE ziwei_ai_requests_total counter\nziwei_ai_requests_total %d\n", m.aiRequests.Load())
		fmt.Fprintf(w, "# TYPE ziwei_ai_errors_total counter\nziwei_ai_errors_total %d\n", m.aiErrors.Load())

		m.mu.Lock()
		keys := make([]string, 0, len(m.byRoute))
		for k := range m.byRoute {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Fprintf(w, "# TYPE ziwei_http_route_requests_total counter\n")
		for _, k := range keys {
			fmt.Fprintf(w, "ziwei_http_route_requests_total{route=%q} %d\n", k, m.byRoute[k])
		}
		m.mu.Unlock()
	}
}

// withInFlight 在途请求计数。
func (m *metrics) withInFlight(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.inFlight.Add(1)
		defer m.inFlight.Add(-1)
		next.ServeHTTP(w, r)
	})
}
