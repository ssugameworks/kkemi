package performance

import (
	"sync"
	"time"

	"github.com/ssugameworks/kkemi/constants"
)

// AdaptiveConcurrencyManager 시스템 부하와 API 응답 시간에 따라 동시성을 동적으로 조정합니다
type AdaptiveConcurrencyManager struct {
	mutex               sync.RWMutex
	currentLimit        int
	minLimit            int
	maxLimit            int
	responseTimeWindow  []time.Duration
	windowSize          int
	adjustmentThreshold time.Duration
	decreaseThreshold   time.Duration
	lastAdjustment      time.Time
	adjustmentCooldown  time.Duration
	successiveIncreases int
	successiveDecreases int
}

// NewAdaptiveConcurrencyManager 새로운 적응형 동시성 관리자를 생성합니다
func NewAdaptiveConcurrencyManager() *AdaptiveConcurrencyManager {
	return &AdaptiveConcurrencyManager{
		currentLimit:        constants.MaxConcurrentRequests,       // 기본값 5로 시작
		minLimit:            constants.AdaptiveConcurrencyMinLimit, // 최소 2개
		maxLimit:            constants.AdaptiveConcurrencyMaxLimit, // 최대 20개
		responseTimeWindow:  make([]time.Duration, 0, constants.ResponseTimeWindowSize),
		windowSize:          constants.ResponseTimeWindowSize,
		adjustmentThreshold: constants.ConcurrencyAdjustmentThreshold, // 500ms 이상이면 감소 고려
		decreaseThreshold:   constants.ConcurrencyDecreaseThreshold,   // 1초 이상이면 즉시 감소
		adjustmentCooldown:  constants.ConcurrencyAdjustmentCooldown,  // 5초마다 조정
		lastAdjustment:      time.Now(),
	}
}

// GetCurrentLimit 현재 동시성 제한을 반환합니다
func (manager *AdaptiveConcurrencyManager) GetCurrentLimit() int {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.currentLimit
}

// RecordResponseTime API 응답 시간을 기록하고 필요시 동시성을 조정합니다
func (manager *AdaptiveConcurrencyManager) RecordResponseTime(responseTime time.Duration) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	// 응답 시간 윈도우에 추가
	manager.responseTimeWindow = append(manager.responseTimeWindow, responseTime)
	if len(manager.responseTimeWindow) > manager.windowSize {
		manager.responseTimeWindow = manager.responseTimeWindow[1:]
	}

	// 충분한 데이터가 있고 쿨다운이 지났으면 조정 시도
	if len(manager.responseTimeWindow) >= constants.MinResponseTimeWindowSize && time.Since(manager.lastAdjustment) > manager.adjustmentCooldown {
		manager.adjustConcurrency()
	}
}

// adjustConcurrency 응답 시간 통계를 기반으로 동시성을 조정합니다
// 주의: 이 메서드는 Lock()이 걸린 상태에서만 호출되어야 합니다
func (manager *AdaptiveConcurrencyManager) adjustConcurrency() {
	avgResponseTime := manager.calculateAverageResponseTime()
	p95ResponseTime := manager.calculateP95ResponseTime()

	oldLimit := manager.currentLimit

	// 응답 시간이 너무 느리면 동시성 감소
	if p95ResponseTime > manager.decreaseThreshold || avgResponseTime > manager.adjustmentThreshold {
		if manager.currentLimit > manager.minLimit {
			manager.currentLimit = max(manager.minLimit, manager.currentLimit-1)
			manager.successiveDecreases++
			manager.successiveIncreases = 0
		}
	} else if avgResponseTime < manager.adjustmentThreshold/2 {
		// 응답 시간이 충분히 빠르고 연속으로 성능이 좋으면 동시성 증가
		if manager.currentLimit < manager.maxLimit && manager.successiveDecreases == 0 {
			// 보수적으로 증가 (연속 증가 횟수에 따라 제한)
			if manager.successiveIncreases < constants.MaxSuccessiveIncreases {
				manager.currentLimit = min(manager.maxLimit, manager.currentLimit+1)
				manager.successiveIncreases++
			}
		}
		manager.successiveDecreases = 0
	}

	if oldLimit != manager.currentLimit {
		manager.lastAdjustment = time.Now()
		// 로깅은 utils 패키지 순환 참조 방지를 위해 제거
	}
}

// calculateAverageResponseTime 평균 응답 시간을 계산합니다
// 주의: 이 메서드는 RLock() 또는 Lock()이 걸린 상태에서만 호출되어야 합니다
func (manager *AdaptiveConcurrencyManager) calculateAverageResponseTime() time.Duration {
	if len(manager.responseTimeWindow) == 0 {
		return 0
	}

	var total time.Duration
	for _, responseTime := range manager.responseTimeWindow {
		total += responseTime
	}
	return total / time.Duration(len(manager.responseTimeWindow))
}

// calculateP95ResponseTime 95 퍼센타일 응답 시간을 계산합니다
// 주의: 이 메서드는 RLock() 또는 Lock()이 걸린 상태에서만 호출되어야 합니다
func (manager *AdaptiveConcurrencyManager) calculateP95ResponseTime() time.Duration {
	if len(manager.responseTimeWindow) == 0 {
		return 0
	}

	// 간단한 95 퍼센타일 계산 (정렬 없이)
	var maxTime time.Duration
	for _, responseTime := range manager.responseTimeWindow {
		if responseTime > maxTime {
			maxTime = responseTime
		}
	}

	// 상위 5%에 해당하는 시간들의 최솟값을 근사치로 사용
	cutoff := time.Duration(float64(maxTime) * constants.P95PercentileRatio) // 대략적인 95 퍼센타일

	return cutoff
}

// GetStats 현재 통계를 반환합니다
func (manager *AdaptiveConcurrencyManager) GetStats() ConcurrencyStats {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	return ConcurrencyStats{
		CurrentLimit:    manager.currentLimit,
		MinLimit:        manager.minLimit,
		MaxLimit:        manager.maxLimit,
		AverageResponse: manager.calculateAverageResponseTime(),
		P95Response:     manager.calculateP95ResponseTime(),
		WindowSize:      len(manager.responseTimeWindow),
		LastAdjustment:  manager.lastAdjustment,
		SuccessiveInc:   manager.successiveIncreases,
		SuccessiveDec:   manager.successiveDecreases,
	}
}

// ConcurrencyStats 동시성 관리자의 통계 정보
type ConcurrencyStats struct {
	CurrentLimit    int
	MinLimit        int
	MaxLimit        int
	AverageResponse time.Duration
	P95Response     time.Duration
	WindowSize      int
	LastAdjustment  time.Time
	SuccessiveInc   int
	SuccessiveDec   int
}
