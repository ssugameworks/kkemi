package cache

import (
	"context"
	"discord-bot/constants"
	"sync"
	"time"
)

// CacheItem 캐시에 저장되는 개별 아이템을 나타냅니다
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired 캐시 아이템이 만료되었는지 확인합니다
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// APICache API 응답을 캐싱하는 인메모리 캐시입니다
type APICache struct {
	userInfoCache          map[string]*CacheItem
	userTop100Cache        map[string]*CacheItem
	userAdditionalCache    map[string]*CacheItem
	userOrganizationsCache map[string]*CacheItem
	mu                     sync.RWMutex

	// 캐시 설정
	userInfoTTL          time.Duration
	userTop100TTL        time.Duration
	userAdditionalTTL    time.Duration
	userOrganizationsTTL time.Duration
}

// NewAPICache 새로운 APICache 인스턴스를 생성합니다
func NewAPICache() *APICache {
	return &APICache{
		userInfoCache:          make(map[string]*CacheItem),
		userTop100Cache:        make(map[string]*CacheItem),
		userAdditionalCache:    make(map[string]*CacheItem),
		userOrganizationsCache: make(map[string]*CacheItem),

		// 캐시 TTL 설정
		userInfoTTL:          constants.UserInfoCacheTTL,       // 사용자 정보 캐시 TTL
		userTop100TTL:        constants.UserTop100CacheTTL,     // TOP 100 캐시 TTL
		userAdditionalTTL:    constants.UserAdditionalCacheTTL, // 추가 정보 캐시 TTL
		userOrganizationsTTL: constants.UserAdditionalCacheTTL, // 조직 정보 캐시 TTL (추가 정보와 동일)
	}
}

// GetUserInfo 캐시에서 사용자 정보를 조회합니다
func (c *APICache) GetUserInfo(handle string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.userInfoCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserInfo 사용자 정보를 캐시에 저장합니다
func (c *APICache) SetUserInfo(handle string, userInfo interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userInfoCache[handle] = &CacheItem{
		Data:      userInfo,
		ExpiresAt: time.Now().Add(c.userInfoTTL),
	}
}

// GetUserTop100 캐시에서 사용자 TOP 100을 조회합니다
func (c *APICache) GetUserTop100(handle string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.userTop100Cache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserTop100 사용자 TOP 100을 캐시에 저장합니다
func (c *APICache) SetUserTop100(handle string, top100 interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userTop100Cache[handle] = &CacheItem{
		Data:      top100,
		ExpiresAt: time.Now().Add(c.userTop100TTL),
	}
}

// GetUserAdditionalInfo 캐시에서 사용자 추가 정보를 조회합니다
func (c *APICache) GetUserAdditionalInfo(handle string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.userAdditionalCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserAdditionalInfo 사용자 추가 정보를 캐시에 저장합니다
func (c *APICache) SetUserAdditionalInfo(handle string, additionalInfo interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userAdditionalCache[handle] = &CacheItem{
		Data:      additionalInfo,
		ExpiresAt: time.Now().Add(c.userAdditionalTTL),
	}
}

// GetUserOrganizations 캐시에서 사용자 조직 정보를 조회합니다
func (c *APICache) GetUserOrganizations(handle string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.userOrganizationsCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserOrganizations 사용자 조직 정보를 캐시에 저장합니다
func (c *APICache) SetUserOrganizations(handle string, organizations interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userOrganizationsCache[handle] = &CacheItem{
		Data:      organizations,
		ExpiresAt: time.Now().Add(c.userOrganizationsTTL),
	}
}

// ClearExpired 만료된 캐시 항목들을 정리합니다
func (c *APICache) ClearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// 만료된 사용자 정보 캐시 정리
	for key, item := range c.userInfoCache {
		if now.After(item.ExpiresAt) {
			delete(c.userInfoCache, key)
		}
	}

	// 만료된 TOP 100 캐시 정리
	for key, item := range c.userTop100Cache {
		if now.After(item.ExpiresAt) {
			delete(c.userTop100Cache, key)
		}
	}

	// 만료된 추가 정보 캐시 정리
	for key, item := range c.userAdditionalCache {
		if now.After(item.ExpiresAt) {
			delete(c.userAdditionalCache, key)
		}
	}

	// 만료된 조직 정보 캐시 정리
	for key, item := range c.userOrganizationsCache {
		if now.After(item.ExpiresAt) {
			delete(c.userOrganizationsCache, key)
		}
	}
}

// GetStats 캐시 통계를 반환합니다
func (c *APICache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		UserInfoCount:          len(c.userInfoCache),
		UserTop100Count:        len(c.userTop100Cache),
		UserAdditionalCount:    len(c.userAdditionalCache),
		UserOrganizationsCount: len(c.userOrganizationsCache),
	}
}

// CacheStats 캐시 통계 정보를 나타냅니다
type CacheStats struct {
	UserInfoCount          int
	UserTop100Count        int
	UserAdditionalCount    int
	UserOrganizationsCount int
}

// Clear 모든 캐시를 삭제합니다
func (c *APICache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userInfoCache = make(map[string]*CacheItem)
	c.userTop100Cache = make(map[string]*CacheItem)
	c.userAdditionalCache = make(map[string]*CacheItem)
	c.userOrganizationsCache = make(map[string]*CacheItem)
}

// StartCleanupWorker 주기적으로 만료된 캐시를 정리하는 워커를 시작합니다
func (c *APICache) StartCleanupWorker(interval time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.ClearExpired()
			case <-ctx.Done():
				return
			}
		}
	}()

	return cancel
}
