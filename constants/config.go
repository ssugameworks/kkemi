package constants

import "time"

// API 및 캐시 설정 상수
const (
	// 캐시 TTL 설정
	UserInfoCacheTTL       = 5 * time.Minute  // 사용자 정보 캐시 만료 시간
	UserTop100CacheTTL     = 10 * time.Minute // TOP 100 캐시 만료 시간
	UserAdditionalCacheTTL = 30 * time.Minute // 추가 정보 캐시 만료 시간
	CacheCleanupInterval   = 5 * time.Minute  // 캐시 정리 간격

	// Discord API 재시도 설정
	MaxDiscordRetries = 3                // 최대 재시도 횟수
	BaseRetryDelay    = 1 * time.Second  // 기본 재시도 지연 시간

	// 성능 및 메모리 관리
	DefaultSliceCapacity = 100           // 기본 슬라이스 용량
	MaxFunctionLines     = 80            // 함수 최대 권장 라인 수
)

// 검증 규칙 상수
const (
	MinBaekjoonIDLength = 3              // 백준 ID 최소 길이
	MaxBaekjoonIDLength = 20             // 백준 ID 최대 길이
	MinNameLength       = 1              // 이름 최소 길이
	MaxNameLength       = 50             // 이름 최대 길이
	MaxParticipants     = 1000           // 최대 참가자 수
)