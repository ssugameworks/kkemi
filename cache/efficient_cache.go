package cache

import (
	"container/heap"
	"context"
	"discord-bot/constants"
	"sync"
	"time"
)

// ExpirationEntry 만료 시간 기반 우선순위 큐의 항목
type ExpirationEntry struct {
	Key       string
	CacheType string // "userInfo", "userTop100", "userAdditional", "userOrganizations"
	ExpiresAt time.Time
	Index     int // 힙에서의 인덱스
}

// ExpirationQueue 만료 시간 기반 우선순위 큐 (최소 힙)
type ExpirationQueue []*ExpirationEntry

func (pq ExpirationQueue) Len() int { return len(pq) }

func (pq ExpirationQueue) Less(i, j int) bool {
	return pq[i].ExpiresAt.Before(pq[j].ExpiresAt)
}

func (pq ExpirationQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *ExpirationQueue) Push(x interface{}) {
	n := len(*pq)
	entry := x.(*ExpirationEntry)
	entry.Index = n
	*pq = append(*pq, entry)
}

func (pq *ExpirationQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil
	entry.Index = -1
	*pq = old[0 : n-1]
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
	keyToEntry     map[string]*ExpirationEntry // 빠른 조회를 위한 인덱스
	
	mu sync.RWMutex

	// 캐시 설정
	userInfoTTL          time.Duration
	userTop100TTL        time.Duration
	userAdditionalTTL    time.Duration
	userOrganizationsTTL time.Duration

	// 효율적인 정리를 위한 설정
	lastCleanup          time.Time
	cleanupBatchSize     int
	maxCleanupDuration   time.Duration
}

// NewEfficientAPICache 새로운 EfficientAPICache 인스턴스를 생성합니다
func NewEfficientAPICache() *EfficientAPICache {
	pq := &ExpirationQueue{}
	heap.Init(pq)
	
	return &EfficientAPICache{
		userInfoCache:          make(map[string]*CacheItem),
		userTop100Cache:        make(map[string]*CacheItem),
		userAdditionalCache:    make(map[string]*CacheItem),
		userOrganizationsCache: make(map[string]*CacheItem),
		
		expirationQueue: pq,
		keyToEntry:     make(map[string]*ExpirationEntry),

		// 캐시 TTL 설정
		userInfoTTL:          constants.UserInfoCacheTTL,
		userTop100TTL:        constants.UserTop100CacheTTL,
		userAdditionalTTL:    constants.UserAdditionalCacheTTL,
		userOrganizationsTTL: constants.UserAdditionalCacheTTL,
		
		// 효율적인 정리 설정
		cleanupBatchSize:   50,                    // 한 번에 50개씩 처리
		maxCleanupDuration: 10 * time.Millisecond, // 최대 10ms까지만 정리 작업
		lastCleanup:        time.Now(),
	}
}

// setWithExpiration 공통 저장 로직 (우선순위 큐에도 추가)
func (c *EfficientAPICache) setWithExpiration(cacheType, key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Now().Add(ttl)
	item := &CacheItem{
		Data:      data,
		ExpiresAt: expiresAt,
	}

	// 기존 항목이 있다면 우선순위 큐에서 제거
	if existingEntry, exists := c.keyToEntry[key]; exists {
		// 힙에서 제거하지 않고 무효화 처리 (성능상 이유)
		existingEntry.ExpiresAt = time.Time{} // 무효화 마크
	}

	// 캐시 맵에 저장
	switch cacheType {
	case "userInfo":
		c.userInfoCache[key] = item
	case "userTop100":
		c.userTop100Cache[key] = item
	case "userAdditional":
		c.userAdditionalCache[key] = item
	case "userOrganizations":
		c.userOrganizationsCache[key] = item
	}

	// 우선순위 큐에 추가
	entry := &ExpirationEntry{
		Key:       key,
		CacheType: cacheType,
		ExpiresAt: expiresAt,
	}
	heap.Push(c.expirationQueue, entry)
	c.keyToEntry[key] = entry
}

// GetUserInfo 캐시에서 사용자 정보를 조회합니다
func (c *EfficientAPICache) GetUserInfo(handle string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.userInfoCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserInfo 사용자 정보를 캐시에 저장합니다
func (c *EfficientAPICache) SetUserInfo(handle string, userInfo interface{}) {
	c.setWithExpiration("userInfo", handle, userInfo, c.userInfoTTL)
}

// GetUserTop100 캐시에서 사용자 TOP 100을 조회합니다
func (c *EfficientAPICache) GetUserTop100(handle string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.userTop100Cache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserTop100 사용자 TOP 100을 캐시에 저장합니다
func (c *EfficientAPICache) SetUserTop100(handle string, top100 interface{}) {
	c.setWithExpiration("userTop100", handle, top100, c.userTop100TTL)
}

// GetUserAdditionalInfo 캐시에서 사용자 추가 정보를 조회합니다
func (c *EfficientAPICache) GetUserAdditionalInfo(handle string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.userAdditionalCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserAdditionalInfo 사용자 추가 정보를 캐시에 저장합니다
func (c *EfficientAPICache) SetUserAdditionalInfo(handle string, additionalInfo interface{}) {
	c.setWithExpiration("userAdditional", handle, additionalInfo, c.userAdditionalTTL)
}

// GetUserOrganizations 캐시에서 사용자 조직 정보를 조회합니다
func (c *EfficientAPICache) GetUserOrganizations(handle string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.userOrganizationsCache[handle]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Data, true
}

// SetUserOrganizations 사용자 조직 정보를 캐시에 저장합니다
func (c *EfficientAPICache) SetUserOrganizations(handle string, organizations interface{}) {
	c.setWithExpiration("userOrganizations", handle, organizations, c.userOrganizationsTTL)
}

// ClearExpiredEfficient 우선순위 큐를 사용하여 효율적으로 만료된 항목을 정리합니다
func (c *EfficientAPICache) ClearExpiredEfficient() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	startTime := time.Now()
	cleaned := 0
	
	// 시간 제한과 배치 크기 제한으로 정리
	for cleaned < c.cleanupBatchSize && time.Since(startTime) < c.maxCleanupDuration {
		if c.expirationQueue.Len() == 0 {
			break
		}

		// 가장 빨리 만료되는 항목 확인
		entry := (*c.expirationQueue)[0]
		
		// 무효화된 항목이거나 아직 만료되지 않은 경우
		if entry.ExpiresAt.IsZero() || now.Before(entry.ExpiresAt) {
			if entry.ExpiresAt.IsZero() {
				// 무효화된 항목은 제거
				heap.Pop(c.expirationQueue)
				delete(c.keyToEntry, entry.Key)
				cleaned++
			} else {
				// 아직 만료되지 않았으므로 정리 중단
				break
			}
			continue
		}

		// 만료된 항목 제거
		heap.Pop(c.expirationQueue)
		delete(c.keyToEntry, entry.Key)
		
		// 해당 캐시 맵에서도 제거
		switch entry.CacheType {
		case "userInfo":
			delete(c.userInfoCache, entry.Key)
		case "userTop100":
			delete(c.userTop100Cache, entry.Key)
		case "userAdditional":
			delete(c.userAdditionalCache, entry.Key)
		case "userOrganizations":
			delete(c.userOrganizationsCache, entry.Key)
		}
		
		cleaned++
	}
	
	c.lastCleanup = now
	return cleaned
}

// GetStats 캐시 통계를 반환합니다
func (c *EfficientAPICache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		UserInfoCount:          len(c.userInfoCache),
		UserTop100Count:        len(c.userTop100Cache),
		UserAdditionalCount:    len(c.userAdditionalCache),
		UserOrganizationsCount: len(c.userOrganizationsCache),
	}
}

// Clear 모든 캐시를 삭제합니다
func (c *EfficientAPICache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userInfoCache = make(map[string]*CacheItem)
	c.userTop100Cache = make(map[string]*CacheItem)
	c.userAdditionalCache = make(map[string]*CacheItem)
	c.userOrganizationsCache = make(map[string]*CacheItem)
	
	// 우선순위 큐와 인덱스도 초기화
	c.expirationQueue = &ExpirationQueue{}
	heap.Init(c.expirationQueue)
	c.keyToEntry = make(map[string]*ExpirationEntry)
}

// StartEfficientCleanupWorker 효율적인 캐시 정리 워커를 시작합니다
func (c *EfficientAPICache) StartEfficientCleanupWorker(interval time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleaned := c.ClearExpiredEfficient()
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

// GetEfficiencyStats 효율적인 캐시의 통계를 반환합니다
func (c *EfficientAPICache) GetEfficiencyStats() EfficiencyStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return EfficiencyStats{
		QueueSize:          c.expirationQueue.Len(),
		KeyIndexSize:       len(c.keyToEntry),
		LastCleanup:        c.lastCleanup,
		CleanupBatchSize:   c.cleanupBatchSize,
		MaxCleanupDuration: c.maxCleanupDuration,
	}
}

// EfficiencyStats 효율적인 캐시의 통계 정보
type EfficiencyStats struct {
	QueueSize          int
	KeyIndexSize       int
	LastCleanup        time.Time
	CleanupBatchSize   int
	MaxCleanupDuration time.Duration
}