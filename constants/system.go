package constants

import "time"

// 시스템 관련 상수
const (
	// 애플리케이션 버전
	BotVersion = "0.2.0" // Discord Bot 버전
	APIVersion = "1.0.0" // 내부 API 버전

	// 네트워크 관련
	DefaultHTTPPort = "8080" // 기본 HTTP 포트 (Railway 헬스체크용)

	// 메모리 관련
	BytesToMB = 1024 * 1024 // 바이트를 MB로 변환하는 계수

	// 헬스체크 관련
	FirestoreHealthCheckTimeout = 5 * time.Second             // Firestore 헬스체크 타임아웃
	HealthCheckCollectionName   = "health_check"              // 헬스체크용 컬렉션명
	HealthStatusHealthy         = "healthy"                   // 정상 상태
	HealthStatusUnhealthy       = "unhealthy"                 // 비정상 상태
	FirestoreNoItemsError       = "no more items in iterator" // Firestore 빈 컬렉션 오류

	// 테스트 관련
	TestAPITimeout = 10 * time.Second // 테스트용 API 타임아웃
	TestRetryDelay = 2 * time.Second  // 테스트용 재시도 지연
)
