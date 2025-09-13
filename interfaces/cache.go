package interfaces

import (
	"context"
	"discord-bot/cache"
	"time"
)

// CacheInterface 캐시 인터페이스를 정의합니다
type CacheInterface interface {
	// 사용자 정보 캐시
	GetUserInfo(handle string) (interface{}, bool)
	SetUserInfo(handle string, userInfo interface{})
	
	// TOP 100 캐시
	GetUserTop100(handle string) (interface{}, bool)
	SetUserTop100(handle string, top100 interface{})
	
	// 추가 정보 캐시
	GetUserAdditionalInfo(handle string) (interface{}, bool)
	SetUserAdditionalInfo(handle string, additionalInfo interface{})
	
	// 조직 정보 캐시
	GetUserOrganizations(handle string) (interface{}, bool)
	SetUserOrganizations(handle string, organizations interface{})
	
	// 통계 및 관리
	GetStats() cache.CacheStats
	Clear()
}

// CleanupWorkerInterface 정리 워커 인터페이스
type CleanupWorkerInterface interface {
	StartCleanupWorker(interval time.Duration) context.CancelFunc
}

// EfficientCleanupWorkerInterface 효율적인 정리 워커 인터페이스
type EfficientCleanupWorkerInterface interface {
	StartEfficientCleanupWorker(interval time.Duration) context.CancelFunc
}