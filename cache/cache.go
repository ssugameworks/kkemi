package cache

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/ssugameworks/Discord-Bot/constants"
)

// CacheItem 캐시에 저장되는 개별 아이템을 나타냅니다
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired 캐시 아이템이 만료되었는지 확인합니다
func (item *CacheItem) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

// CacheStats 캐시 통계 정보를 나타냅니다
type CacheStats struct {
	UserInfoCount          int
	UserTop100Count        int
	UserAdditionalCount    int
	UserOrganizationsCount int
}

// ExpirationEntry 만료 시간 기반 우선순위 큐의 항목
type ExpirationEntry struct {
	Key       string
	CacheType string // "userInfo", "userTop100", "userAdditional", "userOrganizations"
	ExpiresAt time.Time
	Index     int // 힙에서의 인덱스
}

// ExpirationQueue 만료 시간 기반 우선순위 큐 (최소 힙)
type ExpirationQueue []*ExpirationEntry

func (priorityQueue ExpirationQueue) Len() int { return len(priorityQueue) }

func (priorityQueue ExpirationQueue) Less(i, j int) bool {
	return priorityQueue[i].ExpiresAt.Before(priorityQueue[j].ExpiresAt)
}

func (priorityQueue ExpirationQueue) Swap(i, j int) {
	priorityQueue[i], priorityQueue[j] = priorityQueue[j], priorityQueue[i]
	priorityQueue[i].Index = i
	priorityQueue[j].Index = j
}

func (priorityQueue *ExpirationQueue) Push(x interface{}) {
	n := len(*priorityQueue)
	entry := x.(*ExpirationEntry)
	entry.Index = n
	*priorityQueue = append(*priorityQueue, entry)
}

func (priorityQueue *ExpirationQueue) Pop() interface{} {
	old := *priorityQueue
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil
	entry.Index = -1
	*priorityQueue = old[0 : n-1]
	return entry
}

// EfficientAPICache 우선순위 큐를 사용한 효율적인 API 캐시
type EfficientAPICache struct {
	userInfoCache          map[string]*CacheItem
	userTop100Cache        map[string]*CacheItem
	userAdditionalCache    map[string]*CacheItem
	userOrganizationsCache map[string]*CacheItem

	// 만료 시간 추적을 위한 우선순위 큐와 인덱스
	expirationQueue *ExpirationQueue
	keyToEntry      map[string]*ExpirationEntry // 빠른 조회를 위한 인덱스

	mu sync.RWMutex

	// 캐시 설정
	userInfoTTL          time.Duration
	userTop100TTL        time.Duration
	userAdditionalTTL    time.Duration
	userOrganizationsTTL time.Duration

	// 효율적인 정리를 위한 설정
	lastCleanup        time.Time
	cleanupBatchSize   int
	maxCleanupDuration time.Duration
}

// NewEfficientAPICache 새로운 EfficientAPICache 인스턴스를 생성합니다
func NewEfficientAPICache() *EfficientAPICache {
	priorityQueue := &ExpirationQueue{}
	heap.Init(priorityQueue)

	return &EfficientAPICache{
		userInfoCache:          make(map[string]*CacheItem),
		userTop100Cache:        make(map[string]*CacheItem),
		userAdditionalCache:    make(map[string]*CacheItem),
		userOrganizationsCache: make(map[string]*CacheItem),

		expirationQueue: priorityQueue,
		keyToEntry:      make(map[string]*ExpirationEntry),

		// 캐시 TTL 설정
		userInfoTTL:          constants.UserInfoCacheTTL,
		userTop100TTL:        constants.UserTop100CacheTTL,
		userAdditionalTTL:    constants.UserAdditionalCacheTTL,
		userOrganizationsTTL: constants.UserAdditionalCacheTTL,

		// 효율적인 정리 설정
		cleanupBatchSize:   constants.CacheCleanupBatchSize,   // 한 번에 처리할 항목 수
		maxCleanupDuration: constants.MaxCacheCleanupDuration, // 최대 정리 시간
		lastCleanup:        time.Now(),
	}
}

// setWithExpiration 공통 저장 로직 (우선순위 큐에도 추가)
func (cache *EfficientAPICache) setWithExpiration(cacheType, key string, data interface{}, ttl time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	expiresAt := time.Now().Add(ttl)
	item := &CacheItem{
		Data:      data,
		ExpiresAt: expiresAt,
	}

	// 기존 항목이 있다면 우선순위 큐에서 제거
	if existingEntry, exists := cache.keyToEntry[key]; exists {
		// 힙에서 제거하지 않고 무효화 처리 (성능상 이유)
		existingEntry.ExpiresAt = time.Time{} // 무효화 마크
	}

	// 캐시 맵에 저장
	switch cacheType {
	case "userInfo":
		cache.userInfoCache[key] = item
	case "userTop100":
		cache.userTop100Cache[key] = item
	case "userAdditional":
		cache.userAdditionalCache[key] = item
	case "userOrganizations":
		cache.userOrganizationsCache[key] = item
	}

	// 우선순위 큐에 추가
	entry := &ExpirationEntry{
		Key:       key,
		CacheType: cacheType,
		ExpiresAt: expiresAt,
	}
	heap.Push(cache.expirationQueue, entry)
	cache.keyToEntry[key] = entry
}

// GetUserInfo 캐시에서 사용자 정보를 조회합니다
func (cache *EfficientAPICache) GetUserInfo(handle string) (interface{}, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	item, exists := cache.userInfoCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserInfo 사용자 정보를 캐시에 저장합니다
func (cache *EfficientAPICache) SetUserInfo(handle string, userInfo interface{}) {
	cache.setWithExpiration("userInfo", handle, userInfo, cache.userInfoTTL)
}

// GetUserTop100 캐시에서 사용자 TOP 100을 조회합니다
func (cache *EfficientAPICache) GetUserTop100(handle string) (interface{}, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	item, exists := cache.userTop100Cache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserTop100 사용자 TOP 100을 캐시에 저장합니다
func (cache *EfficientAPICache) SetUserTop100(handle string, top100 interface{}) {
	cache.setWithExpiration("userTop100", handle, top100, cache.userTop100TTL)
}

// GetUserAdditionalInfo 캐시에서 사용자 추가 정보를 조회합니다
func (cache *EfficientAPICache) GetUserAdditionalInfo(handle string) (interface{}, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	item, exists := cache.userAdditionalCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserAdditionalInfo 사용자 추가 정보를 캐시에 저장합니다
func (cache *EfficientAPICache) SetUserAdditionalInfo(handle string, additionalInfo interface{}) {
	cache.setWithExpiration("userAdditional", handle, additionalInfo, cache.userAdditionalTTL)
}

// GetUserOrganizations 캐시에서 사용자 조직 정보를 조회합니다
func (cache *EfficientAPICache) GetUserOrganizations(handle string) (interface{}, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	item, exists := cache.userOrganizationsCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserOrganizations 사용자 조직 정보를 캐시에 저장합니다
func (cache *EfficientAPICache) SetUserOrganizations(handle string, organizations interface{}) {
	cache.setWithExpiration("userOrganizations", handle, organizations, cache.userOrganizationsTTL)
}

// ClearExpiredEfficient 우선순위 큐를 사용하여 효율적으로 만료된 항목을 정리합니다
func (cache *EfficientAPICache) ClearExpiredEfficient() int {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	now := time.Now()
	startTime := time.Now()
	cleaned := 0

	// 시간 제한과 배치 크기 제한으로 정리
	for cleaned < cache.cleanupBatchSize && time.Since(startTime) < cache.maxCleanupDuration {
		if cache.expirationQueue.Len() == 0 {
			break
		}

		// 가장 빨리 만료되는 항목 확인
		entry := (*cache.expirationQueue)[0]

		// 무효화된 항목이거나 아직 만료되지 않은 경우
		if entry.ExpiresAt.IsZero() || now.Before(entry.ExpiresAt) {
			if entry.ExpiresAt.IsZero() {
				// 무효화된 항목은 제거
				heap.Pop(cache.expirationQueue)
				delete(cache.keyToEntry, entry.Key)
				cleaned++
			} else {
				// 아직 만료되지 않았으므로 정리 중단
				break
			}
			continue
		}

		// 만료된 항목 제거
		heap.Pop(cache.expirationQueue)
		delete(cache.keyToEntry, entry.Key)

		// 해당 캐시 맵에서도 제거
		switch entry.CacheType {
		case "userInfo":
			delete(cache.userInfoCache, entry.Key)
		case "userTop100":
			delete(cache.userTop100Cache, entry.Key)
		case "userAdditional":
			delete(cache.userAdditionalCache, entry.Key)
		case "userOrganizations":
			delete(cache.userOrganizationsCache, entry.Key)
		}

		cleaned++
	}

	cache.lastCleanup = now
	return cleaned
}

// GetStats 캐시 통계를 반환합니다
func (cache *EfficientAPICache) GetStats() CacheStats {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	return CacheStats{
		UserInfoCount:          len(cache.userInfoCache),
		UserTop100Count:        len(cache.userTop100Cache),
		UserAdditionalCount:    len(cache.userAdditionalCache),
		UserOrganizationsCount: len(cache.userOrganizationsCache),
	}
}

// Clear 모든 캐시를 삭제합니다
func (cache *EfficientAPICache) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.userInfoCache = make(map[string]*CacheItem)
	cache.userTop100Cache = make(map[string]*CacheItem)
	cache.userAdditionalCache = make(map[string]*CacheItem)
	cache.userOrganizationsCache = make(map[string]*CacheItem)

	// 우선순위 큐와 인덱스도 초기화
	cache.expirationQueue = &ExpirationQueue{}
	heap.Init(cache.expirationQueue)
	cache.keyToEntry = make(map[string]*ExpirationEntry)
}

// StartEfficientCleanupWorker 효율적인 캐시 정리 워커를 시작합니다
func (cache *EfficientAPICache) StartEfficientCleanupWorker(interval time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleaned := cache.ClearExpiredEfficient()
				if cleaned > 0 {
					// 로깅은 순환 참조 방지를 위해 제거
					// utils.Debug("Cleaned %d expired cache entries", cleaned)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return cancel
}
