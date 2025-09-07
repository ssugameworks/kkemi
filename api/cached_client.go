package api

import (
	"context"
	"discord-bot/cache"
	"discord-bot/constants"
	"discord-bot/utils"
	"fmt"
	"sync/atomic"
)

// CachedSolvedACClient 캐시 기능을 포함한 SolvedAC API 클라이언트입니다
type CachedSolvedACClient struct {
	client        *SolvedACClient
	cache         *cache.APICache
	cleanupCancel context.CancelFunc

	// 성능 메트릭
	cacheHits   int64
	cacheMisses int64
	totalCalls  int64
}

// NewCachedSolvedACClient 새로운 CachedSolvedACClient 인스턴스를 생성합니다
func NewCachedSolvedACClient() *CachedSolvedACClient {
	utils.Info("Creating cached SolvedAC API client")
	client := &CachedSolvedACClient{
		client: NewSolvedACClient(),
		cache:  cache.NewAPICache(),
	}

	// 캐시 정리 워커 시작
	client.cleanupCancel = client.cache.StartCleanupWorker(constants.CacheCleanupInterval)
	return client
}

// Close는 캐시 정리 워커를 중지시킵니다.
func (c *CachedSolvedACClient) Close() {
	if c.cleanupCancel != nil {
		c.cleanupCancel()
		utils.Info("Cache cleanup worker stopped.")
	}
}

// GetUserInfo 캐시를 통해 사용자 정보를 조회합니다
func (c *CachedSolvedACClient) GetUserInfo(handle string) (*UserInfo, error) {
	atomic.AddInt64(&c.totalCalls, 1)

	// 캐시에서 먼저 조회
	if cachedData, found := c.cache.GetUserInfo(handle); found {
		atomic.AddInt64(&c.cacheHits, 1)
		utils.Debug("Cache hit for user info: %s", handle)
		return cachedData.(*UserInfo), nil
	}

	// 캐시 미스 - API 호출
	atomic.AddInt64(&c.cacheMisses, 1)
	utils.Debug("Cache miss for user info: %s, calling API", handle)

	userInfo, err := c.client.GetUserInfo(handle)
	if err != nil {
		return nil, err
	}

	// 성공한 응답을 캐시에 저장
	c.cache.SetUserInfo(handle, userInfo)

	return userInfo, nil
}

// GetUserTop100 캐시를 통해 사용자 TOP 100을 조회합니다
func (c *CachedSolvedACClient) GetUserTop100(handle string) (*Top100Response, error) {
	atomic.AddInt64(&c.totalCalls, 1)

	// 캐시에서 먼저 조회
	if cachedData, found := c.cache.GetUserTop100(handle); found {
		atomic.AddInt64(&c.cacheHits, 1)
		utils.Debug("Cache hit for user top100: %s", handle)
		return cachedData.(*Top100Response), nil
	}

	// 캐시 미스 - API 호출
	atomic.AddInt64(&c.cacheMisses, 1)
	utils.Debug("Cache miss for user top100: %s, calling API", handle)

	top100, err := c.client.GetUserTop100(handle)
	if err != nil {
		return nil, err
	}

	// 성공한 응답을 캐시에 저장
	c.cache.SetUserTop100(handle, top100)

	return top100, nil
}

// GetUserAdditionalInfo 캐시를 통해 사용자 추가 정보를 조회합니다
func (c *CachedSolvedACClient) GetUserAdditionalInfo(handle string) (*UserAdditionalInfo, error) {
	atomic.AddInt64(&c.totalCalls, 1)

	// 캐시에서 먼저 조회
	if cachedData, found := c.cache.GetUserAdditionalInfo(handle); found {
		atomic.AddInt64(&c.cacheHits, 1)
		utils.Debug("Cache hit for user additional info: %s", handle)
		return cachedData.(*UserAdditionalInfo), nil
	}

	// 캐시 미스 - API 호출
	atomic.AddInt64(&c.cacheMisses, 1)
	utils.Debug("Cache miss for user additional info: %s, calling API", handle)

	additionalInfo, err := c.client.GetUserAdditionalInfo(handle)
	if err != nil {
		return nil, err
	}

	// 성공한 응답을 캐시에 저장
	c.cache.SetUserAdditionalInfo(handle, additionalInfo)

	return additionalInfo, nil
}

// GetCacheStats 캐시 통계를 반환합니다
func (c *CachedSolvedACClient) GetCacheStats() CacheMetrics {
	cacheStats := c.cache.GetStats()

	totalCalls := atomic.LoadInt64(&c.totalCalls)
	hits := atomic.LoadInt64(&c.cacheHits)
	misses := atomic.LoadInt64(&c.cacheMisses)

	var hitRate float64
	if totalCalls > 0 {
		hitRate = float64(hits) / float64(totalCalls) * 100
	}

	return CacheMetrics{
		TotalCalls:          totalCalls,
		CacheHits:           hits,
		CacheMisses:         misses,
		HitRate:             hitRate,
		UserInfoCached:      cacheStats.UserInfoCount,
		UserTop100Cached:    cacheStats.UserTop100Count,
		UserAdditionalCached: cacheStats.UserAdditionalCount,
	}
}

// CacheMetrics 캐시 성능 메트릭을 나타냅니다
type CacheMetrics struct {
	TotalCalls          int64
	CacheHits           int64
	CacheMisses         int64
	HitRate             float64
	UserInfoCached      int
	UserTop100Cached    int
	UserAdditionalCached int
}

// String CacheMetrics의 문자열 표현을 반환합니다
func (m CacheMetrics) String() string {
	return fmt.Sprintf("API Cache Stats: Calls=%d, Hits=%d, Misses=%d, Hit Rate=%.2f%%, Cached Items: UserInfo=%d, Top100=%d, Additional=%d",
		m.TotalCalls, m.CacheHits, m.CacheMisses, m.HitRate,
		m.UserInfoCached, m.UserTop100Cached, m.UserAdditionalCached)
}

// ClearCache 모든 캐시를 삭제합니다
func (c *CachedSolvedACClient) ClearCache() {
	c.cache.Clear()
	atomic.StoreInt64(&c.cacheHits, 0)
	atomic.StoreInt64(&c.cacheMisses, 0)
	atomic.StoreInt64(&c.totalCalls, 0)
	utils.Info("API cache cleared")
}

// WarmupCache 주요 참가자들에 대한 캐시를 미리 로드합니다
func (c *CachedSolvedACClient) WarmupCache(handles []string) error {
	utils.Info("Starting cache warmup for %d users", len(handles))

	for _, handle := range handles {
		// 이미 캐시에 있다면 스킵
		if _, found := c.cache.GetUserInfo(handle); found {
			continue
		}

		// 백그라운드에서 데이터 로드
		go func(h string) {
			if _, err := c.GetUserInfo(h); err != nil {
				utils.Warn("Cache warmup failed for user info %s: %v", h, err)
			}
			if _, err := c.GetUserTop100(h); err != nil {
				utils.Warn("Cache warmup failed for top100 %s: %v", h, err)
			}
		}(handle)
	}

	utils.Info("Cache warmup initiated")
	return nil
}
