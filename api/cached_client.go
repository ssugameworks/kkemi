package api

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/ssugameworks/kkemi/cache"
	"github.com/ssugameworks/kkemi/constants"
	"github.com/ssugameworks/kkemi/utils"
)

// CachedSolvedACClient 캐시 기능을 포함한 SolvedAC API 클라이언트입니다
type CachedSolvedACClient struct {
	client *SolvedACClient
	cache  interface { // 기존 APICache와 EfficientAPICache 모두 지원
		GetUserInfo(string) (interface{}, bool)
		SetUserInfo(string, interface{})
		GetUserTop100(string) (interface{}, bool)
		SetUserTop100(string, interface{})
		GetUserAdditionalInfo(string) (interface{}, bool)
		SetUserAdditionalInfo(string, interface{})
		GetUserOrganizations(string) (interface{}, bool)
		SetUserOrganizations(string, interface{})
		GetStats() cache.CacheStats
		Clear()
	}
	cleanupCancel context.CancelFunc

	// 성능 메트릭
	cacheHits   int64
	cacheMisses int64
	totalCalls  int64
}

// NewCachedSolvedACClient 새로운 CachedSolvedACClient 인스턴스를 생성합니다
func NewCachedSolvedACClient() *CachedSolvedACClient {
	utils.Info("Creating cached SolvedAC API client with efficient cache")

	// 효율적인 캐시를 사용하되, 기존 인터페이스와 호환되도록 래핑
	efficientCache := cache.NewEfficientAPICache()

	client := &CachedSolvedACClient{
		client: NewSolvedACClient(),
		cache:  efficientCache,
	}

	// 효율적인 캐시 정리 워커 시작
	client.cleanupCancel = efficientCache.StartEfficientCleanupWorker(constants.CacheCleanupInterval)
	return client
}

// Close 캐시 정리 워커를 중지시킵니다.
func (cachedClient *CachedSolvedACClient) Close() {
	if cachedClient.cleanupCancel != nil {
		cachedClient.cleanupCancel()
		utils.Info("Cache cleanup worker stopped.")
	}
}

// GetUserInfo 캐시를 통해 사용자 정보를 조회합니다
func (cachedClient *CachedSolvedACClient) GetUserInfo(ctx context.Context, handle string) (*UserInfo, error) {
	atomic.AddInt64(&cachedClient.totalCalls, 1)

	// 캐시에서 먼저 조회
	if cachedData, found := cachedClient.cache.GetUserInfo(handle); found {
		atomic.AddInt64(&cachedClient.cacheHits, 1)
		utils.Debug("Cache hit for user info: %s", handle)
		return cachedData.(*UserInfo), nil
	}

	// 캐시 미스 - API 호출
	atomic.AddInt64(&cachedClient.cacheMisses, 1)
	utils.Debug("Cache miss for user info: %s, calling API", handle)

	userInfo, err := cachedClient.client.GetUserInfo(ctx, handle)
	if err != nil {
		return nil, err
	}

	// 성공한 응답을 캐시에 저장
	cachedClient.cache.SetUserInfo(handle, userInfo)

	return userInfo, nil
}

// GetUserTop100 캐시를 통해 사용자 TOP 100을 조회합니다
func (cachedClient *CachedSolvedACClient) GetUserTop100(ctx context.Context, handle string) (*Top100Response, error) {
	atomic.AddInt64(&cachedClient.totalCalls, 1)

	// 캐시에서 먼저 조회
	if cachedData, found := cachedClient.cache.GetUserTop100(handle); found {
		atomic.AddInt64(&cachedClient.cacheHits, 1)
		utils.Debug("Cache hit for user top100: %s", handle)
		return cachedData.(*Top100Response), nil
	}

	// 캐시 미스 - API 호출
	atomic.AddInt64(&cachedClient.cacheMisses, 1)
	utils.Debug("Cache miss for user top100: %s, calling API", handle)

	top100, err := cachedClient.client.GetUserTop100(ctx, handle)
	if err != nil {
		return nil, err
	}

	// 성공한 응답을 캐시에 저장
	cachedClient.cache.SetUserTop100(handle, top100)

	return top100, nil
}

// GetUserAdditionalInfo 캐시를 통해 사용자 추가 정보를 조회합니다
func (cachedClient *CachedSolvedACClient) GetUserAdditionalInfo(ctx context.Context, handle string) (*UserAdditionalInfo, error) {
	atomic.AddInt64(&cachedClient.totalCalls, 1)

	// 캐시에서 먼저 조회
	if cachedData, found := cachedClient.cache.GetUserAdditionalInfo(handle); found {
		atomic.AddInt64(&cachedClient.cacheHits, 1)
		utils.Debug("Cache hit for user additional info: %s", handle)
		return cachedData.(*UserAdditionalInfo), nil
	}

	// 캐시 미스 - API 호출
	atomic.AddInt64(&cachedClient.cacheMisses, 1)
	utils.Debug("Cache miss for user additional info: %s, calling API", handle)

	additionalInfo, err := cachedClient.client.GetUserAdditionalInfo(ctx, handle)
	if err != nil {
		return nil, err
	}

	// 성공한 응답을 캐시에 저장
	cachedClient.cache.SetUserAdditionalInfo(handle, additionalInfo)

	return additionalInfo, nil
}

// GetUserOrganizations 지정된 사용자의 소속 조직 목록을 가져옵니다 (캐시 포함)
func (cachedClient *CachedSolvedACClient) GetUserOrganizations(ctx context.Context, handle string) ([]Organization, error) {
	atomic.AddInt64(&cachedClient.totalCalls, 1)

	// 캐시에서 먼저 조회
	if cachedData, found := cachedClient.cache.GetUserOrganizations(handle); found {
		atomic.AddInt64(&cachedClient.cacheHits, 1)
		utils.Debug("Cache hit for user organizations: %s", handle)
		return cachedData.([]Organization), nil
	}

	// 캐시 미스 - API 호출
	atomic.AddInt64(&cachedClient.cacheMisses, 1)
	utils.Debug("Cache miss for user organizations: %s, calling API", handle)

	organizations, err := cachedClient.client.GetUserOrganizations(ctx, handle)
	if err != nil {
		return nil, err
	}

	// 성공한 응답을 캐시에 저장
	cachedClient.cache.SetUserOrganizations(handle, organizations)

	return organizations, nil
}

// GetCacheStats 캐시 통계를 반환합니다
func (cachedClient *CachedSolvedACClient) GetCacheStats() CacheMetrics {
	cacheStats := cachedClient.cache.GetStats()

	totalCalls := atomic.LoadInt64(&cachedClient.totalCalls)
	hits := atomic.LoadInt64(&cachedClient.cacheHits)
	misses := atomic.LoadInt64(&cachedClient.cacheMisses)

	var hitRate float64
	if totalCalls > 0 {
		hitRate = float64(hits) / float64(totalCalls) * 100
	}

	return CacheMetrics{
		TotalCalls:           totalCalls,
		CacheHits:            hits,
		CacheMisses:          misses,
		HitRate:              hitRate,
		UserInfoCached:       cacheStats.UserInfoCount,
		UserTop100Cached:     cacheStats.UserTop100Count,
		UserAdditionalCached: cacheStats.UserAdditionalCount,
	}
}

// CacheMetrics 캐시 성능 메트릭을 나타냅니다
type CacheMetrics struct {
	TotalCalls           int64
	CacheHits            int64
	CacheMisses          int64
	HitRate              float64
	UserInfoCached       int
	UserTop100Cached     int
	UserAdditionalCached int
}

// String CacheMetrics의 문자열 표현을 반환합니다
func (metrics CacheMetrics) String() string {
	return fmt.Sprintf("API Cache Stats: Calls=%d, Hits=%d, Misses=%d, Hit Rate=%.2f%%, Cached Items: UserInfo=%d, Top100=%d, Additional=%d",
		metrics.TotalCalls, metrics.CacheHits, metrics.CacheMisses, metrics.HitRate,
		metrics.UserInfoCached, metrics.UserTop100Cached, metrics.UserAdditionalCached)
}

// ClearCache 모든 캐시를 삭제합니다
func (cachedClient *CachedSolvedACClient) ClearCache() {
	cachedClient.cache.Clear()
	atomic.StoreInt64(&cachedClient.cacheHits, 0)
	atomic.StoreInt64(&cachedClient.cacheMisses, 0)
	atomic.StoreInt64(&cachedClient.totalCalls, 0)
	utils.Info("API cache cleared")
}

// WarmupCache 주요 참가자들에 대한 캐시를 미리 로드합니다
func (cachedClient *CachedSolvedACClient) WarmupCache(handles []string) error {
	utils.Info("Starting cache warmup for %d users", len(handles))

	for _, handle := range handles {
		// 이미 캐시에 있다면 스킵
		if _, found := cachedClient.cache.GetUserInfo(handle); found {
			continue
		}

		// 백그라운드에서 데이터 로드
		go func(h string) {
			ctx := context.Background()
			if _, err := cachedClient.GetUserInfo(ctx, h); err != nil {
				utils.Warn("Cache warmup failed for user info %s: %v", h, err)
			}
			if _, err := cachedClient.GetUserTop100(ctx, h); err != nil {
				utils.Warn("Cache warmup failed for top100 %s: %v", h, err)
			}
		}(handle)
	}

	utils.Info("Cache warmup initiated")
	return nil
}
