package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"cloud.google.com/go/firestore"
)

// HealthStatus 헬스체크 응답 구조체
type HealthStatus struct {
	Status       string            `json:"status"`
	Timestamp    time.Time         `json:"timestamp"`
	Uptime       string            `json:"uptime"`
	Version      string            `json:"version"`
	GoVersion    string            `json:"go_version"`
	Memory       string            `json:"memory_usage"`
	Dependencies map[string]string `json:"dependencies"`
}

// HealthChecker 헬스체크를 수행하는 인터페이스
type HealthChecker interface {
	CheckHealth() (string, error)
}

// FirestoreHealthChecker Firestore 연결 상태를 확인하는 구조체
type FirestoreHealthChecker struct {
	client *firestore.Client
}

// NewFirestoreHealthChecker 새로운 FirestoreHealthChecker를 생성합니다
func NewFirestoreHealthChecker(client *firestore.Client) *FirestoreHealthChecker {
	return &FirestoreHealthChecker{client: client}
}

// CheckHealth Firestore 연결 상태를 확인합니다
func (f *FirestoreHealthChecker) CheckHealth() (string, error) {
	if f.client == nil {
		return "disconnected", fmt.Errorf("firestore client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 간단한 쿼리로 연결 상태 확인
	_, err := f.client.Collection("health_check").Limit(1).Documents(ctx).Next()
	if err != nil {
		// 컬렉션이 없어도 연결은 정상이므로 특정 에러는 무시
		if err.Error() == "no more items in iterator" {
			return "connected", nil
		}
		return "error", err
	}

	return "connected", nil
}

var (
	startTime      = time.Now()
	healthCheckers = make(map[string]HealthChecker)
)

// RegisterHealthChecker 헬스체크 시스템에 의존성을 등록합니다
func RegisterHealthChecker(name string, checker HealthChecker) {
	healthCheckers[name] = checker
}

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
	
	dependencies := make(map[string]string)
	overallStatus := "healthy"
	
	// 등록된 모든 의존성 헬스체크 수행
	for name, checker := range healthCheckers {
		status, err := checker.CheckHealth()
		if err != nil {
			dependencies[name] = fmt.Sprintf("%s: %v", status, err)
			if status == "error" || status == "disconnected" {
				overallStatus = "unhealthy"
			}
		} else {
			dependencies[name] = status
		}
	}
	
	status := HealthStatus{
		Status:       overallStatus,
		Timestamp:    time.Now(),
		Uptime:       time.Since(startTime).String(),
		Version:      "v1.0.0",
		GoVersion:    runtime.Version(),
		Memory:       fmt.Sprintf("%.2f MB", float64(memStats.Alloc)/1024/1024),
		Dependencies: dependencies,
	}
	
	// 전체 상태에 따라 HTTP 상태 코드 설정
	if overallStatus == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(status)
}