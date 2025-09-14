package constants

import "time"

// 적응형 동시성 제어 관련 상수
const (
	// 동시성 제한
	AdaptiveConcurrencyMinLimit = 2  // 최소 동시 요청 수
	AdaptiveConcurrencyMaxLimit = 20 // 최대 동시 요청 수

	// 응답 시간 관련
	ResponseTimeWindowSize         = 50                     // 응답 시간 윈도우 크기
	MinResponseTimeWindowSize      = 10                     // 조정을 위한 최소 윈도우 크기
	ConcurrencyAdjustmentThreshold = 500 * time.Millisecond // 동시성 감소 고려 임계값
	ConcurrencyDecreaseThreshold   = 1 * time.Second        // 즉시 감소 임계값
	ConcurrencyAdjustmentCooldown  = 5 * time.Second        // 조정 간 쿨다운 시간
	MaxSuccessiveIncreases         = 3                      // 최대 연속 증가 횟수
	P95PercentileRatio             = 0.8                    // 95 퍼센타일 근사 비율 (8/10)

	// 메모리 풀 관련
	MaxPoolSliceCapacity   = 200  // 풀에 반환할 최대 슬라이스 용량
	MaxPoolChannelCapacity = 100  // 풀에 반환할 최대 채널 용량
	MaxPoolSemaphoreSize   = 20   // 풀에 반환할 최대 세마포어 크기
	MaxStringBuilderSize   = 1024 // 풀에 반환할 최대 문자열 빌더 크기

	// 캐시 효율성 관련
	CacheCleanupBatchSize   = 50                    // 한 번에 정리할 캐시 항목 수
	MaxCacheCleanupDuration = 10 * time.Millisecond // 최대 캐시 정리 시간
)
