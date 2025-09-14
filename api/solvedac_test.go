package api

import (
	"github.com/ssugameworks/Discord-Bot/constants"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSolvedACClient(t *testing.T) {
	client := NewSolvedACClient()

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.baseURL != "https://solved.ac/api/v3" {
		t.Errorf("Expected base URL 'https://solved.ac/api/v3', got '%s'", client.baseURL)
	}

	if client.client == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

func TestSolvedACClient_GetUserInfo_Success(t *testing.T) {
	// Mock 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/show" {
			t.Errorf("Expected path '/user/show', got '%s'", r.URL.Path)
		}

		handle := r.URL.Query().Get("handle")
		if handle != "testuser" {
			t.Errorf("Expected handle 'testuser', got '%s'", handle)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"handle": "testuser",
			"bio": "Test user",
			"rating": 1500,
			"tier": 15,
			"class": 5,
			"classDecoration": "gold",
			"profileImageUrl": "https://example.com/avatar.png",
			"solvedCount": 100,
			"verified": true,
			"rank": 1000
		}`))
	}))
	defer server.Close()

	client := &SolvedACClient{
		client:  &http.Client{Timeout: constants.TestAPITimeout},
		baseURL: server.URL,
	}

	userInfo, err := client.GetUserInfo("testuser")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if userInfo.Handle != "testuser" {
		t.Errorf("Expected handle 'testuser', got '%s'", userInfo.Handle)
	}

	if userInfo.Bio != "Test user" {
		t.Errorf("Expected bio 'Test user', got '%s'", userInfo.Bio)
	}

	if userInfo.Rating != 1500 {
		t.Errorf("Expected rating 1500, got %d", userInfo.Rating)
	}

	if userInfo.Tier != 15 {
		t.Errorf("Expected tier 15, got %d", userInfo.Tier)
	}
}

func TestSolvedACClient_GetUserInfo_NotFound(t *testing.T) {
	// Mock 서버 - 404 응답
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "User not found"}`))
	}))
	defer server.Close()

	client := &SolvedACClient{
		client:  &http.Client{Timeout: constants.TestAPITimeout},
		baseURL: server.URL,
	}

	userInfo, err := client.GetUserInfo("nonexistent")

	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	if userInfo != nil {
		t.Error("Expected nil userInfo on error")
	}
}

func TestSolvedACClient_GetUserTop100_Success(t *testing.T) {
	// Mock 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/top_100" {
			t.Errorf("Expected path '/user/top_100', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"count": 2,
			"items": [
				{
					"problemId": 1000,
					"level": 1,
					"titleKo": "A+B",
					"acceptedUserCount": 100000,
					"averageTries": 1.5
				},
				{
					"problemId": 1001,
					"level": 2,
					"titleKo": "A-B",
					"acceptedUserCount": 80000,
					"averageTries": 2.0
				}
			]
		}`))
	}))
	defer server.Close()

	client := &SolvedACClient{
		client:  &http.Client{Timeout: constants.TestAPITimeout},
		baseURL: server.URL,
	}

	top100, err := client.GetUserTop100("testuser")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if top100.Count != 2 {
		t.Errorf("Expected count 2, got %d", top100.Count)
	}

	if len(top100.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(top100.Items))
	}

	if top100.Items[0].ProblemID != 1000 {
		t.Errorf("Expected problem ID 1000, got %d", top100.Items[0].ProblemID)
	}

	if top100.Items[0].TitleKo != "A+B" {
		t.Errorf("Expected title 'A+B', got '%s'", top100.Items[0].TitleKo)
	}
}

func TestSolvedACClient_GetUserAdditionalInfo_Success(t *testing.T) {
	// Mock 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"countryCode": "KR"
		}`))
	}))
	defer server.Close()

	client := &SolvedACClient{
		client:  &http.Client{Timeout: constants.TestAPITimeout},
		baseURL: server.URL,
	}

	additionalInfo, err := client.GetUserAdditionalInfo("testuser")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if additionalInfo.CountryCode == nil || *additionalInfo.CountryCode != "KR" {
		t.Error("Expected country code 'KR'")
	}
}

func TestSolvedACClient_Timeout(t *testing.T) {
	// 느린 Mock 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(constants.TestRetryDelay) // 테스트용 대기
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &SolvedACClient{
		client:  &http.Client{Timeout: 100 * time.Millisecond}, // 100ms 타임아웃
		baseURL: server.URL,
	}

	_, err := client.GetUserInfo("testuser")

	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestSolvedACClient_InvalidJSON(t *testing.T) {
	// 잘못된 JSON을 반환하는 Mock 서버
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := &SolvedACClient{
		client:  &http.Client{Timeout: constants.TestAPITimeout},
		baseURL: server.URL,
	}

	_, err := client.GetUserInfo("testuser")

	if err == nil {
		t.Error("Expected JSON parsing error")
	}
}

// 통합 테스트
func TestSolvedACClient_Integration(t *testing.T) {
	client := NewSolvedACClient()

	if client == nil {
		t.Fatal("Failed to create client")
	}

	// 실제 API 테스트는 외부 의존성이 있으므로 생략
	// 대신 클라이언트 초기화만 테스트

	if client.baseURL == "" {
		t.Error("Base URL should not be empty")
	}

	if client.client == nil {
		t.Error("HTTP client should not be nil")
	}
}
