package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// HealthStatus 헬스체크 응답 구조체
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
	Version   string    `json:"version"`
	GoVersion string    `json:"go_version"`
	Memory    string    `json:"memory_usage"`
}

var startTime = time.Now()

// StartHealthServer 헬스체크 HTTP 서버를 시작합니다
func StartHealthServer(port string) {
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", healthHandler) // Railway의 기본 헬스체크
	
	go func() {
		fmt.Printf("Health check server starting on port %s\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Printf("Health server error: %v\n", err)
		}
	}()
}

// healthHandler 헬스체크 핸들러
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
		Version:   "v1.0.0",
		GoVersion: runtime.Version(),
		Memory:    fmt.Sprintf("%.2f MB", float64(memStats.Alloc)/1024/1024),
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}