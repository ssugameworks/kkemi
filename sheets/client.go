package sheets

import (
	"context"
	"fmt"
	"github.com/ssugameworks/Discord-Bot/constants"
	"github.com/ssugameworks/Discord-Bot/utils"
	"os"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient Google Sheets API 클라이언트
type SheetsClient struct {
	service *sheets.Service
	ctx     context.Context
}

// NewSheetsClient 새로운 Google Sheets 클라이언트를 생성합니다
func NewSheetsClient() (*SheetsClient, error) {
	ctx := context.Background()

	// Firebase 인증 정보 사용 (Google Cloud 프로젝트와 동일)
	credentialsJSON := setupGoogleCredentials()
	if credentialsJSON == "" {
		return nil, fmt.Errorf("Google credentials not available")
	}

	service, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credentialsJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets service: %w", err)
	}

	utils.Info("Google Sheets client initialized successfully")
	return &SheetsClient{
		service: service,
		ctx:     ctx,
	}, nil
}

// IsNameInParticipantList 주어진 이름이 참가자 명단에 있는지 확인합니다
func (c *SheetsClient) IsNameInParticipantList(name string) (bool, error) {
	// 스프레드시트 데이터 읽기
	resp, err := c.service.Spreadsheets.Values.Get(
		constants.ParticipantSpreadsheetID,
		constants.ParticipantSheetRange,
	).Do()
	if err != nil {
		return false, fmt.Errorf("failed to read spreadsheet: %w", err)
	}

	if len(resp.Values) == 0 {
		utils.Warn("Spreadsheet is empty")
		return false, nil
	}

	// 헤더 행에서 "이름 (ex.홍길동)" 컬럼 찾기
	headers := resp.Values[0]
	nameColumnIndex := -1
	for i, header := range headers {
		if headerStr, ok := header.(string); ok {
			if strings.Contains(headerStr, constants.ParticipantNameColumn) {
				nameColumnIndex = i
				break
			}
		}
	}

	if nameColumnIndex == -1 {
		return false, fmt.Errorf("name column '%s' not found in spreadsheet", constants.ParticipantNameColumn)
	}

	// 데이터 행에서 이름 검색
	normalizedTargetName := normalizeKoreanName(name)
	for i := 1; i < len(resp.Values); i++ { // 헤더 행 제외
		row := resp.Values[i]
		if nameColumnIndex < len(row) {
			if cellValue, ok := row[nameColumnIndex].(string); ok {
				normalizedCellName := normalizeKoreanName(cellValue)
				if normalizedCellName == normalizedTargetName {
					utils.Info("Name '%s' found in participant list at row %d", name, i+1)
					return true, nil
				}
			}
		}
	}

	utils.Info("Name '%s' not found in participant list", name)
	return false, nil
}

// normalizeKoreanName 한글 이름을 정규화합니다 (공백 제거, 대소문자 통일 등)
func normalizeKoreanName(name string) string {
	// 앞뒤 공백 제거
	normalized := strings.TrimSpace(name)
	// 중간 공백 제거
	normalized = strings.ReplaceAll(normalized, " ", "")
	// 소문자로 변환 (영어가 포함된 경우)
	normalized = strings.ToLower(normalized)
	return normalized
}

// setupGoogleCredentials Google 인증 정보를 설정합니다
func setupGoogleCredentials() string {
	// Firebase 인증 JSON 사용
	firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if firebaseCredentials == "" {
		utils.Warn("FIREBASE_CREDENTIALS_JSON environment variable is not set")
		return ""
	}

	return firebaseCredentials
}