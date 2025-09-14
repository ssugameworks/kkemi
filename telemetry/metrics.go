package telemetry

import (
	"context"
	"discord-bot/utils"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MetricsClient Google Cloud Monitoring 클라이언트를 래핑합니다
type MetricsClient struct {
	client    *monitoring.MetricClient
	projectID string
	enabled   bool
}

// NewMetricsClient 새로운 MetricsClient 인스턴스를 생성합니다
func NewMetricsClient(projectID string) *MetricsClient {
	if projectID == "" {
		utils.Warn("Project ID not provided, telemetry disabled")
		return &MetricsClient{enabled: false}
	}

	// Firebase 인증 정보를 임시 파일로 생성하여 Google Cloud 인증에 사용
	if err := setupGoogleCloudCredentials(); err != nil {
		utils.Warn("Failed to setup Google Cloud credentials: %v", err)
		utils.Warn("Telemetry disabled - ensure Firebase credentials are available")
		return &MetricsClient{enabled: false}
	}

	client, err := monitoring.NewMetricClient(context.Background())
	if err != nil {
		utils.Warn("Failed to create monitoring client: %v", err)
		utils.Warn("Telemetry disabled")
		return &MetricsClient{enabled: false}
	}

	utils.Info("Google Cloud Monitoring telemetry enabled for project: %s", projectID)
	return &MetricsClient{
		client:    client,
		projectID: projectID,
		enabled:   true,
	}
}

// SendCacheMetrics 캐시 메트릭을 Google Cloud Monitoring으로 전송합니다
func (m *MetricsClient) SendCacheMetrics(totalCalls, cacheHits, cacheMisses int64, hitRate float64) {
	if !m.enabled {
		return
	}

	ctx := context.Background()
	now := &timestamppb.Timestamp{
		Seconds: time.Now().Unix(),
	}

	// 캐시 히트율 메트릭
	if err := m.sendCustomMetric(ctx, "discord_bot/cache/hit_rate", hitRate, now); err != nil {
		utils.Warn("Failed to send cache hit rate metric: %v", err)
	}

	// 총 API 호출 수 메트릭
	if err := m.sendCustomMetric(ctx, "discord_bot/cache/total_calls", float64(totalCalls), now); err != nil {
		utils.Warn("Failed to send total calls metric: %v", err)
	}

	// 캐시 히트 수 메트릭
	if err := m.sendCustomMetric(ctx, "discord_bot/cache/hits", float64(cacheHits), now); err != nil {
		utils.Warn("Failed to send cache hits metric: %v", err)
	}

	// 캐시 미스 수 메트릭
	if err := m.sendCustomMetric(ctx, "discord_bot/cache/misses", float64(cacheMisses), now); err != nil {
		utils.Warn("Failed to send cache misses metric: %v", err)
	}

	utils.Debug("Cache metrics sent to Google Cloud Monitoring")
}

// SendCommandMetric 명령어 사용 메트릭을 전송합니다
func (m *MetricsClient) SendCommandMetric(command string, isAdmin bool) {
	if !m.enabled {
		return
	}

	ctx := context.Background()
	now := &timestamppb.Timestamp{
		Seconds: time.Now().Unix(),
	}

	metricType := "discord_bot/commands/usage"

	// 라벨을 포함한 리소스 생성
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", m.projectID),
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metric.Metric{
					Type: fmt.Sprintf("custom.googleapis.com/%s", metricType),
					Labels: map[string]string{
						"command":  command,
						"is_admin": fmt.Sprintf("%t", isAdmin),
					},
				},
				Resource: &monitoredres.MonitoredResource{
					Type: "generic_task",
					Labels: map[string]string{
						"project_id": m.projectID,
						"location":   "global",
						"namespace":  "discord-bot",
						"job":        "competition-bot",
						"task_id":    "main",
					},
				},
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: now,
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: 1,
							},
						},
					},
				},
			},
		},
	}

	if err := m.client.CreateTimeSeries(ctx, req); err != nil {
		utils.Warn("Failed to send command metric: %v", err)
		return
	}

	utils.Debug("Command metric sent: %s (admin: %t)", command, isAdmin)

}

// SendCompetitionMetric 대회 관련 메트릭을 전송합니다
func (m *MetricsClient) SendCompetitionMetric(action string, participantCount int) {
	if !m.enabled {
		return
	}

	ctx := context.Background()
	now := &timestamppb.Timestamp{
		Seconds: time.Now().Unix(),
	}

	// 대회 액션 메트릭
	if err := m.sendLabeledMetric(ctx, "discord_bot/competition/actions", 1.0, now, map[string]string{
		"action": action,
	}); err != nil {
		utils.Warn("Failed to send competition action metric: %v", err)
	}

	// 참가자 수 메트릭
	if err := m.sendCustomMetric(ctx, "discord_bot/competition/participants", float64(participantCount), now); err != nil {
		utils.Warn("Failed to send participant count metric: %v", err)
	}

	utils.Debug("Competition metric sent: %s (participants: %d)", action, participantCount)

}

// SendPerformanceMetric 성능 메트릭을 전송합니다
func (m *MetricsClient) SendPerformanceMetric(operation string, duration time.Duration, success bool) {
	if !m.enabled {
		return
	}

	ctx := context.Background()
	now := &timestamppb.Timestamp{
		Seconds: time.Now().Unix(),
	}

	// 응답 시간 메트릭
	if err := m.sendLabeledMetric(ctx, "discord_bot/performance/duration", duration.Seconds(), now, map[string]string{
		"operation": operation,
		"success":   fmt.Sprintf("%t", success),
	}); err != nil {
		utils.Warn("Failed to send performance duration metric: %v", err)
	}

	// 성공률 메트릭
	successValue := 0.0
	if success {
		successValue = 1.0
	}
	if err := m.sendLabeledMetric(ctx, "discord_bot/performance/success_rate", successValue, now, map[string]string{
		"operation": operation,
	}); err != nil {
		utils.Warn("Failed to send success rate metric: %v", err)
	}

	utils.Debug("Performance metric sent: %s (duration: %v, success: %t)", operation, duration, success)

}

// sendCustomMetric 단순한 커스텀 메트릭을 전송합니다
func (m *MetricsClient) sendCustomMetric(ctx context.Context, metricType string, value float64, timestamp *timestamppb.Timestamp) error {
	return m.sendLabeledMetric(ctx, metricType, value, timestamp, nil)
}

// sendLabeledMetric 라벨이 포함된 커스텀 메트릭을 전송합니다
func (m *MetricsClient) sendLabeledMetric(ctx context.Context, metricType string, value float64, timestamp *timestamppb.Timestamp, labels map[string]string) error {
	if labels == nil {
		labels = make(map[string]string)
	}

	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", m.projectID),
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metric.Metric{
					Type:   fmt.Sprintf("custom.googleapis.com/%s", metricType),
					Labels: labels,
				},
				Resource: &monitoredres.MonitoredResource{
					Type: "generic_task",
					Labels: map[string]string{
						"project_id": m.projectID,
						"location":   "global",
						"namespace":  "discord-bot",
						"job":        "competition-bot",
						"task_id":    "main",
					},
				},
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: timestamp,
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_DoubleValue{
								DoubleValue: value,
							},
						},
					},
				},
			},
		},
	}

	return m.client.CreateTimeSeries(ctx, req)
}

// Close 클라이언트를 정리합니다
func (m *MetricsClient) Close() error {
	if !m.enabled || m.client == nil {
		return nil
	}
	return m.client.Close()
}

// setupGoogleCloudCredentials Firebase 인증 정보를 Google Cloud 인증으로 설정합니다
func setupGoogleCloudCredentials() error {
	// 이미 GOOGLE_APPLICATION_CREDENTIALS가 설정되어 있다면 스킵
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		return nil
	}

	// Firebase 인증 JSON이 있는지 확인
	firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if firebaseCredentials == "" {
		return fmt.Errorf("neither GOOGLE_APPLICATION_CREDENTIALS nor FIREBASE_CREDENTIALS_JSON is set")
	}

	// 임시 파일 생성
	tempDir := os.TempDir()
	credFile := filepath.Join(tempDir, "discord-bot-gcloud-credentials.json")

	// JSON 내용을 임시 파일에 저장
	err := ioutil.WriteFile(credFile, []byte(firebaseCredentials), 0600)
	if err != nil {
		return fmt.Errorf("failed to write temporary credentials file: %v", err)
	}

	// 환경변수 설정
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFile)

	utils.Debug("Created temporary Google Cloud credentials file: %s", credFile)
	return nil
}
