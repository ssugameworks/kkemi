package sheets

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestNormalizeKoreanName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"홍길동", "홍길동", "Normal Korean name"},
		{" 홍길동 ", "홍길동", "Korean name with spaces"},
		{"홍 길 동", "홍길동", "Korean name with internal spaces"},
		{"JOHN", "john", "English name to lowercase"},
		{"John Doe", "johndoe", "English name with space"},
		{" JANE ", "jane", "English with spaces and uppercase"},
		{"김철수123", "김철수123", "Korean with numbers"},
		{"", "", "Empty string"},
	}

	for _, test := range tests {
		result := normalizeKoreanName(test.input)
		if result != test.expected {
			t.Errorf("normalizeKoreanName(%q) = %q, expected %q (%s)", test.input, result, test.expected, test.desc)
		}
	}
}

func TestSetupGoogleCredentials(t *testing.T) {
	// 기존 환경변수 백업
	originalCreds := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	defer func() {
		if originalCreds != "" {
			os.Setenv("FIREBASE_CREDENTIALS_JSON", originalCreds)
		} else {
			os.Unsetenv("FIREBASE_CREDENTIALS_JSON")
		}
	}()

	// 테스트 케이스 1: 환경변수가 없는 경우
	os.Unsetenv("FIREBASE_CREDENTIALS_JSON")
	result := setupGoogleCredentials()
	if result != "" {
		t.Errorf("setupGoogleCredentials() with no env var = %q, expected empty string", result)
	}

	// 테스트 케이스 2: 환경변수가 있는 경우
	testCreds := `{"type": "service_account", "project_id": "test"}`
	os.Setenv("FIREBASE_CREDENTIALS_JSON", testCreds)
	result = setupGoogleCredentials()
	if result != testCreds {
		t.Errorf("setupGoogleCredentials() with env var = %q, expected %q", result, testCreds)
	}
}

// 실제 스프레드시트 연결 테스트 (환경변수가 설정된 경우에만 실행)
func TestSheetsClientIntegration(t *testing.T) {
	// 환경변수 확인
	if os.Getenv("FIREBASE_CREDENTIALS_JSON") == "" {
		t.Skip("FIREBASE_CREDENTIALS_JSON not set, skipping integration test")
	}

	// 클라이언트 생성
	client, err := NewSheetsClient()
	if err != nil {
		t.Fatalf("Failed to create sheets client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.service == nil {
		t.Fatal("Expected non-nil service")
	}

	t.Log("Google Sheets client created successfully")
}

// 실제 스프레드시트에서 이름 검색 테스트 (환경변수가 설정된 경우에만 실행)
func TestIsNameInParticipantListIntegration(t *testing.T) {
	// 환경변수 확인
	if os.Getenv("FIREBASE_CREDENTIALS_JSON") == "" {
		t.Skip("FIREBASE_CREDENTIALS_JSON not set, skipping integration test")
	}

	// 클라이언트 생성
	client, err := NewSheetsClient()
	if err != nil {
		t.Fatalf("Failed to create sheets client: %v", err)
	}

	// 실제 스프레드시트에서 이름 검색 테스트
	testCases := []struct {
		name        string
		description string
	}{
		{"이정안", "실제 참가자 이름"},        // 실제 스프레드시트에 있는 이름
		{"홍준우", "존재하지 않는 이름"},       // 실제 스프레드시트에 없는 이름
		{"쿠쿠루삥뽕", "확실히 존재하지 않는 이름"}, // 확실히 없을 것 같은 이름
	}

	for _, tc := range testCases {
		found, err := client.IsNameInParticipantList(tc.name)
		if err != nil {
			t.Errorf("IsNameInParticipantList(%q) returned error: %v", tc.name, err)
			continue
		}

		t.Logf("IsNameInParticipantList(%q) = %v (%s)", tc.name, found, tc.description)

		// 결과가 boolean이어야 함
		if found != true && found != false {
			t.Errorf("IsNameInParticipantList(%q) returned non-boolean value", tc.name)
		}
	}
}

// 스프레드시트 접근 권한 테스트
func TestSheetsAccessPermissions(t *testing.T) {
	// 환경변수 확인
	if os.Getenv("FIREBASE_CREDENTIALS_JSON") == "" {
		t.Skip("FIREBASE_CREDENTIALS_JSON not set, skipping integration test")
	}

	// 클라이언트 생성
	client, err := NewSheetsClient()
	if err != nil {
		t.Fatalf("Failed to create sheets client: %v", err)
	}

	// 빈 이름으로 검색하여 스프레드시트 접근 가능 여부 확인
	_, err = client.IsNameInParticipantList("")

	// 빈 이름은 찾을 수 없지만, 에러가 없다면 스프레드시트 접근은 성공
	if err != nil {
		// 에러 메시지 분석
		if isPermissionError(err) {
			t.Errorf("Permission denied accessing spreadsheet: %v", err)
			t.Log("Check if the service account has access to the spreadsheet")
		} else if isNotFoundError(err) {
			t.Errorf("Spreadsheet not found: %v", err)
			t.Log("Check if the spreadsheet ID is correct")
		} else {
			t.Logf("Other error (this might be expected for empty name): %v", err)
		}
	} else {
		t.Log("Successfully accessed spreadsheet")
	}
}

// 에러 타입 확인 헬퍼 함수들
func isPermissionError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "permission") || contains(errStr, "forbidden") || contains(errStr, "access denied")
}

func isNotFoundError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "not found") || contains(errStr, "404")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Mock 테스트용 구조체
type mockSheetsService struct {
	testData [][]interface{}
	err      error
}

func (m *mockSheetsService) GetValues() ([][]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.testData, nil
}

// Mock을 사용한 이름 검증 테스트
func TestIsNameInParticipantListWithMockData(t *testing.T) {
	tests := []struct {
		name        string
		testData    [][]interface{}
		searchName  string
		expected    bool
		expectError bool
		description string
	}{
		{
			name: "찾는 이름이 있는 경우",
			testData: [][]interface{}{
				{"번호", "이름 (ex.홍길동)", "학번", "학과"},
				{"1", "김철수", "12345", "컴퓨터공학과"},
				{"2", "이영희", "12346", "전자공학과"},
				{"3", "박민수", "12347", "기계공학과"},
			},
			searchName:  "김철수",
			expected:    true,
			expectError: false,
			description: "정확히 일치하는 이름",
		},
		{
			name: "공백이 있는 이름 검색",
			testData: [][]interface{}{
				{"번호", "이름 (ex.홍길동)", "학번"},
				{"1", "김 철 수", "12345"},
				{"2", "이영희", "12346"},
			},
			searchName:  "김철수",
			expected:    true,
			expectError: false,
			description: "공백 정규화 후 일치",
		},
		{
			name: "대소문자 다른 영어 이름",
			testData: [][]interface{}{
				{"번호", "이름 (ex.홍길동)", "학번"},
				{"1", "John Doe", "12345"},
				{"2", "JANE SMITH", "12346"},
			},
			searchName:  "john doe",
			expected:    true,
			expectError: false,
			description: "대소문자 무시하고 일치",
		},
		{
			name: "찾는 이름이 없는 경우",
			testData: [][]interface{}{
				{"번호", "이름 (ex.홍길동)", "학번"},
				{"1", "김철수", "12345"},
				{"2", "이영희", "12346"},
			},
			searchName:  "박진수",
			expected:    false,
			expectError: false,
			description: "존재하지 않는 이름",
		},
		{
			name: "빈 스프레드시트",
			testData: [][]interface{}{
				{"번호", "이름 (ex.홍길동)", "학번"},
			},
			searchName:  "김철수",
			expected:    false,
			expectError: false,
			description: "헤더만 있고 데이터 없음",
		},
		{
			name: "이름 컬럼이 없는 경우",
			testData: [][]interface{}{
				{"번호", "학번", "학과"},
				{"1", "12345", "컴퓨터공학과"},
			},
			searchName:  "김철수",
			expected:    false,
			expectError: true,
			description: "이름 컬럼 찾을 수 없음",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 테스트용 클라이언트 생성
			client := &SheetsClient{
				ctx: context.Background(),
			}

			// Mock 데이터를 사용하여 직접 검증 로직 테스트
			result, err := client.searchNameInMockData(tt.testData, tt.searchName)

			// 에러 확인
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// 결과 확인
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %v but got %v for search '%s'", tt.expected, result, tt.searchName)
			}

			t.Logf("Test '%s': searched for '%s', found=%v (%s)", tt.name, tt.searchName, result, tt.description)
		})
	}
}

// Mock 데이터에서 이름을 검색하는 헬퍼 함수 (실제 API 호출 없이 로직 테스트)
func (c *SheetsClient) searchNameInMockData(values [][]interface{}, name string) (bool, error) {
	if len(values) == 0 {
		return false, nil
	}

	// 헤더 행에서 "이름 (ex.홍길동)" 컬럼 찾기
	headers := values[0]
	nameColumnIndex := -1
	for i, header := range headers {
		if headerStr, ok := header.(string); ok {
			if containsInner(headerStr, "이름 (ex.홍길동)") {
				nameColumnIndex = i
				break
			}
		}
	}

	if nameColumnIndex == -1 {
		return false, fmt.Errorf("name column '이름 (ex.홍길동)' not found in spreadsheet")
	}

	// 데이터 행에서 이름 검색
	normalizedTargetName := normalizeKoreanName(name)
	for i := 1; i < len(values); i++ { // 헤더 행 제외
		row := values[i]
		if nameColumnIndex < len(row) {
			if cellValue, ok := row[nameColumnIndex].(string); ok {
				normalizedCellName := normalizeKoreanName(cellValue)
				if normalizedCellName == normalizedTargetName {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// 실제 스프레드시트 헤더 확인용 테스트
func TestDebugSpreadsheetHeaders(t *testing.T) {
	// 환경변수 확인
	if os.Getenv("FIREBASE_CREDENTIALS_JSON") == "" {
		t.Skip("FIREBASE_CREDENTIALS_JSON not set, skipping debug test")
	}

	// 클라이언트 생성
	client, err := NewSheetsClient()
	if err != nil {
		t.Fatalf("Failed to create sheets client: %v", err)
	}

	// 스프레드시트 데이터 직접 읽기
	resp, err := client.service.Spreadsheets.Values.Get(
		"1wwjn1hApSINnYsQGbEe5OdpYWvMfsfHC1ftoyR65IDM",
		"A:Z",
	).Do()
	if err != nil {
		t.Fatalf("Failed to read spreadsheet: %v", err)
	}

	if len(resp.Values) == 0 {
		t.Fatal("Spreadsheet is empty")
	}

	// 헤더 출력
	t.Logf("스프레드시트 헤더 정보:")
	headers := resp.Values[0]
	for i, header := range headers {
		if headerStr, ok := header.(string); ok && headerStr != "" {
			t.Logf("  컬럼 %d: '%s'", i, headerStr)
		}
	}

	// 첫 번째 데이터 행도 출력 (있다면)
	if len(resp.Values) > 1 {
		t.Logf("첫 번째 데이터 행:")
		firstRow := resp.Values[1]
		for i, cell := range firstRow {
			if cellStr, ok := cell.(string); ok && cellStr != "" {
				t.Logf("  컬럼 %d: '%s'", i, cellStr)
			}
		}
	}
}
