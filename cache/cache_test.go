package cache

import (
	"container/heap"
	"testing"
	"time"
)

func TestNewEfficientAPICache(t *testing.T) {
	cache := NewEfficientAPICache()

	if cache == nil {
		t.Fatal("NewEfficientAPICache가 nil을 반환했습니다")
	}

	if cache.userInfoCache == nil {
		t.Error("userInfoCache가 초기화되지 않았습니다")
	}

	if cache.userTop100Cache == nil {
		t.Error("userTop100Cache가 초기화되지 않았습니다")
	}

	if cache.userAdditionalCache == nil {
		t.Error("userAdditionalCache가 초기화되지 않았습니다")
	}

	if cache.userOrganizationsCache == nil {
		t.Error("userOrganizationsCache가 초기화되지 않았습니다")
	}

	if cache.expirationQueue == nil {
		t.Error("expirationQueue가 초기화되지 않았습니다")
	}

	if cache.keyToEntry == nil {
		t.Error("keyToEntry가 초기화되지 않았습니다")
	}
}

func TestCacheItemIsExpired(t *testing.T) {
	// 만료되지 않은 아이템
	notExpiredItem := &CacheItem{
		Data:      "테스트 데이터",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if notExpiredItem.IsExpired() {
		t.Error("아직 만료되지 않은 아이템이 만료된 것으로 판단됩니다")
	}

	// 만료된 아이템
	expiredItem := &CacheItem{
		Data:      "만료된 데이터",
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	if !expiredItem.IsExpired() {
		t.Error("만료된 아이템이 만료되지 않은 것으로 판단됩니다")
	}
}

func TestUserInfoCache(t *testing.T) {
	cache := NewEfficientAPICache()
	handle := "testuser"
	testData := "사용자 정보 데이터"

	// 캐시 미스
	data, exists := cache.GetUserInfo(handle)
	if exists {
		t.Error("존재하지 않는 데이터가 존재하는 것으로 조회됩니다")
	}
	if data != nil {
		t.Error("존재하지 않는 데이터가 nil이 아닙니다")
	}

	// 캐시 저장
	cache.SetUserInfo(handle, testData)

	// 캐시 히트
	data, exists = cache.GetUserInfo(handle)
	if !exists {
		t.Error("저장된 데이터를 찾을 수 없습니다")
	}
	if data != testData {
		t.Errorf("데이터가 일치하지 않습니다. 예상: %v, 실제: %v", testData, data)
	}
}

func TestUserTop100Cache(t *testing.T) {
	cache := NewEfficientAPICache()
	handle := "testuser"
	testData := "TOP 100 데이터"

	// 캐시 미스
	data, exists := cache.GetUserTop100(handle)
	if exists {
		t.Error("존재하지 않는 데이터가 존재하는 것으로 조회됩니다")
	}

	// 캐시 저장
	cache.SetUserTop100(handle, testData)

	// 캐시 히트
	data, exists = cache.GetUserTop100(handle)
	if !exists {
		t.Error("저장된 데이터를 찾을 수 없습니다")
	}
	if data != testData {
		t.Errorf("데이터가 일치하지 않습니다. 예상: %v, 실제: %v", testData, data)
	}
}

func TestUserAdditionalInfoCache(t *testing.T) {
	cache := NewEfficientAPICache()
	handle := "testuser"
	testData := "추가 정보 데이터"

	// 캐시 미스
	data, exists := cache.GetUserAdditionalInfo(handle)
	if exists {
		t.Error("존재하지 않는 데이터가 존재하는 것으로 조회됩니다")
	}

	// 캐시 저장
	cache.SetUserAdditionalInfo(handle, testData)

	// 캐시 히트
	data, exists = cache.GetUserAdditionalInfo(handle)
	if !exists {
		t.Error("저장된 데이터를 찾을 수 없습니다")
	}
	if data != testData {
		t.Errorf("데이터가 일치하지 않습니다. 예상: %v, 실제: %v", testData, data)
	}
}

func TestUserOrganizationsCache(t *testing.T) {
	cache := NewEfficientAPICache()
	handle := "testuser"
	testData := "조직 정보 데이터"

	// 캐시 미스
	data, exists := cache.GetUserOrganizations(handle)
	if exists {
		t.Error("존재하지 않는 데이터가 존재하는 것으로 조회됩니다")
	}

	// 캐시 저장
	cache.SetUserOrganizations(handle, testData)

	// 캐시 히트
	data, exists = cache.GetUserOrganizations(handle)
	if !exists {
		t.Error("저장된 데이터를 찾을 수 없습니다")
	}
	if data != testData {
		t.Errorf("데이터가 일치하지 않습니다. 예상: %v, 실제: %v", testData, data)
	}
}

func TestCacheStats(t *testing.T) {
	cache := NewEfficientAPICache()

	// 초기 상태 확인
	stats := cache.GetStats()
	if stats.UserInfoCount != 0 {
		t.Errorf("UserInfoCount가 0이어야 합니다. 실제값: %d", stats.UserInfoCount)
	}
	if stats.UserTop100Count != 0 {
		t.Errorf("UserTop100Count가 0이어야 합니다. 실제값: %d", stats.UserTop100Count)
	}
	if stats.UserAdditionalCount != 0 {
		t.Errorf("UserAdditionalCount가 0이어야 합니다. 실제값: %d", stats.UserAdditionalCount)
	}
	if stats.UserOrganizationsCount != 0 {
		t.Errorf("UserOrganizationsCount가 0이어야 합니다. 실제값: %d", stats.UserOrganizationsCount)
	}

	// 데이터 추가 후 확인
	cache.SetUserInfo("user1", "데이터1")
	cache.SetUserInfo("user2", "데이터2")
	cache.SetUserTop100("user1", "TOP100 데이터")
	cache.SetUserAdditionalInfo("user1", "추가 정보")
	cache.SetUserOrganizations("user1", "조직 정보")

	stats = cache.GetStats()
	if stats.UserInfoCount != 2 {
		t.Errorf("UserInfoCount가 2여야 합니다. 실제값: %d", stats.UserInfoCount)
	}
	if stats.UserTop100Count != 1 {
		t.Errorf("UserTop100Count가 1이어야 합니다. 실제값: %d", stats.UserTop100Count)
	}
	if stats.UserAdditionalCount != 1 {
		t.Errorf("UserAdditionalCount가 1이어야 합니다. 실제값: %d", stats.UserAdditionalCount)
	}
	if stats.UserOrganizationsCount != 1 {
		t.Errorf("UserOrganizationsCount가 1이어야 합니다. 실제값: %d", stats.UserOrganizationsCount)
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewEfficientAPICache()

	// 데이터 추가
	cache.SetUserInfo("user1", "데이터1")
	cache.SetUserTop100("user1", "TOP100 데이터")
	cache.SetUserAdditionalInfo("user1", "추가 정보")
	cache.SetUserOrganizations("user1", "조직 정보")

	// 데이터가 있는지 확인
	stats := cache.GetStats()
	if stats.UserInfoCount == 0 {
		t.Error("데이터가 추가되지 않았습니다")
	}

	// 캐시 클리어
	cache.Clear()

	// 클리어 후 확인
	stats = cache.GetStats()
	if stats.UserInfoCount != 0 {
		t.Errorf("Clear 후 UserInfoCount가 0이어야 합니다. 실제값: %d", stats.UserInfoCount)
	}
	if stats.UserTop100Count != 0 {
		t.Errorf("Clear 후 UserTop100Count가 0이어야 합니다. 실제값: %d", stats.UserTop100Count)
	}
	if stats.UserAdditionalCount != 0 {
		t.Errorf("Clear 후 UserAdditionalCount가 0이어야 합니다. 실제값: %d", stats.UserAdditionalCount)
	}
	if stats.UserOrganizationsCount != 0 {
		t.Errorf("Clear 후 UserOrganizationsCount가 0이어야 합니다. 실제값: %d", stats.UserOrganizationsCount)
	}
}

func TestExpirationQueue(t *testing.T) {
	// 우선순위 큐 직접 테스트
	var pq ExpirationQueue
	heap.Init(&pq) // heap 초기화

	now := time.Now()
	entry1 := &ExpirationEntry{
		Key:       "key1",
		CacheType: "userInfo",
		ExpiresAt: now.Add(time.Hour),
	}
	entry2 := &ExpirationEntry{
		Key:       "key2",
		CacheType: "userInfo",
		ExpiresAt: now.Add(30 * time.Minute),
	}
	entry3 := &ExpirationEntry{
		Key:       "key3",
		CacheType: "userInfo",
		ExpiresAt: now.Add(2 * time.Hour),
	}

	// heap.Push 사용
	heap.Push(&pq, entry1)
	heap.Push(&pq, entry3)
	heap.Push(&pq, entry2)

	if pq.Len() != 3 {
		t.Errorf("큐 길이가 3이어야 합니다. 실제값: %d", pq.Len())
	}

	// 가장 빠른 만료 시간 확인 (entry2가 30분)
	firstEntry := pq[0]
	if firstEntry.Key != "key2" {
		t.Errorf("첫 번째 요소가 key2여야 합니다. 실제값: %s", firstEntry.Key)
	}
}

func TestCacheExpiredItems(t *testing.T) {
	cache := NewEfficientAPICache()

	// TTL을 매우 짧게 설정하여 테스트 (반사 사용)
	cache.userInfoTTL = 10 * time.Millisecond

	handle := "testuser"
	testData := "테스트 데이터"

	// 데이터 저장
	cache.SetUserInfo(handle, testData)

	// 즉시 조회 - 성공해야 함
	data, exists := cache.GetUserInfo(handle)
	if !exists || data != testData {
		t.Error("방금 저장한 데이터를 조회할 수 없습니다")
	}

	// TTL 대기
	time.Sleep(20 * time.Millisecond)

	// 만료 후 조회 - 실패해야 함
	data, exists = cache.GetUserInfo(handle)
	if exists {
		t.Error("만료된 데이터가 여전히 존재합니다")
	}
}

func TestStartEfficientCleanupWorker(t *testing.T) {
	cache := NewEfficientAPICache()

	// 매우 짧은 간격으로 정리 워커 시작
	cancel := cache.StartEfficientCleanupWorker(50 * time.Millisecond)
	defer cancel()

	// 워커가 실행되는지 확인하기 위해 짧은 시간 대기
	time.Sleep(100 * time.Millisecond)

	// cancel 함수가 정상적으로 작동하는지 확인
	cancel()

	// 추가 대기 후 정상 종료 확인
	time.Sleep(100 * time.Millisecond)
}

func TestClearExpiredEfficient(t *testing.T) {
	cache := NewEfficientAPICache()

	// 짧은 TTL로 설정
	cache.userInfoTTL = 10 * time.Millisecond

	// 여러 데이터 추가
	cache.SetUserInfo("user1", "데이터1")
	cache.SetUserInfo("user2", "데이터2")
	cache.SetUserInfo("user3", "데이터3")

	// 데이터가 있는지 확인
	stats := cache.GetStats()
	if stats.UserInfoCount != 3 {
		t.Errorf("UserInfoCount가 3이어야 합니다. 실제값: %d", stats.UserInfoCount)
	}

	// TTL 만료 대기
	time.Sleep(20 * time.Millisecond)

	// 만료된 항목 정리
	cleaned := cache.ClearExpiredEfficient()

	// 정리된 항목 수 확인
	if cleaned <= 0 {
		t.Errorf("만료된 항목이 정리되어야 합니다. 정리된 수: %d", cleaned)
	}

	// 정리 후 통계 확인
	statsAfter := cache.GetStats()
	if statsAfter.UserInfoCount >= stats.UserInfoCount {
		t.Error("정리 후 캐시 항목 수가 줄어들어야 합니다")
	}
}

func TestCacheOverwrite(t *testing.T) {
	cache := NewEfficientAPICache()
	handle := "testuser"

	// 첫 번째 데이터 저장
	firstData := "첫 번째 데이터"
	cache.SetUserInfo(handle, firstData)

	data, exists := cache.GetUserInfo(handle)
	if !exists || data != firstData {
		t.Error("첫 번째 데이터가 저장되지 않았습니다")
	}

	// 같은 키로 두 번째 데이터 저장 (덮어쓰기)
	secondData := "두 번째 데이터"
	cache.SetUserInfo(handle, secondData)

	data, exists = cache.GetUserInfo(handle)
	if !exists || data != secondData {
		t.Error("데이터가 올바르게 덮어써지지 않았습니다")
	}

	// 통계는 여전히 1이어야 함
	stats := cache.GetStats()
	if stats.UserInfoCount != 1 {
		t.Errorf("덮어쓰기 후에도 항목 수는 1이어야 합니다. 실제값: %d", stats.UserInfoCount)
	}
}
