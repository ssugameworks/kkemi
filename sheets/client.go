package sheets

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ssugameworks/Discord-Bot/constants"
	"github.com/ssugameworks/Discord-Bot/utils"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"github.com/ssugameworks/Discord-Bot/models"
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

// UpdateScoreboardSheet 스코어보드 정보를 스프레드시트에 업데이트합니다
func (c *SheetsClient) UpdateScoreboardSheet(spreadsheetID string, scores []models.ScoreData) error {
	if len(scores) == 0 {
		utils.Warn("No scores to update in spreadsheet")
		return nil
	}

	// 먼저 시트를 클리어
	err := c.clearSheet(spreadsheetID)
	if err != nil {
		utils.Warn("Failed to clear sheet: %v", err)
	}

	// 데이터 준비
	var values [][]interface{}

	// 타이틀과 업데이트 시간
	now := utils.GetCurrentTimeKST()
	titleRow := []interface{}{
		"🏆 잔디심기 챌린지 스코어보드",
		"",
		"",
		"",
		"",
		fmt.Sprintf("업데이트: %s", now.Format("2006-01-02 15:04:05 KST")),
	}
	values = append(values, titleRow)
	values = append(values, []interface{}{}) // 빈 행

	// 전체 헤더 행
	headers := []interface{}{
		"순위", "이름", "백준ID", "점수", "리그", "티어", "레이팅", "신규해결문제", "백준프로필",
	}
	values = append(values, headers)

	// 리그별로 점수 분류 및 정렬
	leagueScores := groupScoresByLeague(scores)
	leagueOrder := []int{0, 1, 2} // LeagueRookie, LeaguePro, LeagueMaster

	for _, league := range leagueOrder {
		if len(leagueScores[league]) == 0 {
			continue
		}

		// 리그 헤더 추가
		leagueName := getLeagueName(league)
		leagueHeader := []interface{}{
			fmt.Sprintf("🎯 %s 리그", leagueName), "", "", "", "", "", "", "", "",
		}
		values = append(values, leagueHeader)

		// 점수 데이터 추가
		var lastRawScore float64 = -1.0
		var rank int
		for i, score := range leagueScores[league] {
			if score.RawScore != lastRawScore {
				rank = i + 1
			}

			// 티어 이름 변환
			tierName := getTierName(score.CurrentTier)

			// 백준 프로필 링크
			profileLink := fmt.Sprintf("https://www.acmicpc.net/user/%s", score.BaekjoonID)

			row := []interface{}{
				rank,
				score.Name,
				score.BaekjoonID,
				int(score.Score),
				leagueName,
				tierName,
				score.CurrentRating,
				score.ProblemCount,
				profileLink,
			}
			values = append(values, row)
			lastRawScore = score.RawScore
		}

		// 빈 행 추가 (리그 간 구분)
		values = append(values, []interface{}{})
	}

	// 푸터 추가
	values = append(values, []interface{}{})
	footerRow := []interface{}{
		"📊 데이터는 30분마다 자동 업데이트됩니다",
		"",
		"",
		"",
		"",
		fmt.Sprintf("총 참가자: %d명", len(scores)),
	}
	values = append(values, footerRow)

	// 스프레드시트 업데이트
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err = c.service.Spreadsheets.Values.Update(
		spreadsheetID,
		"A1", // 시작 셀
		valueRange,
	).ValueInputOption("RAW").Do()

	if err != nil {
		return fmt.Errorf("failed to update spreadsheet: %w", err)
	}

	utils.Info("Successfully updated scoreboard spreadsheet with %d participants", len(scores))
	return nil
}

// clearSheet 시트의 모든 데이터를 클리어합니다
func (c *SheetsClient) clearSheet(spreadsheetID string) error {
	_, err := c.service.Spreadsheets.Values.Clear(
		spreadsheetID,
		"A:Z", // 전체 범위 클리어
		&sheets.ClearValuesRequest{},
	).Do()
	return err
}

// getTierName 티어 번호를 티어 이름으로 변환합니다
func getTierName(tier int) string {
	tierNames := map[int]string{
		0:  "Unrated",
		1:  "Bronze V", 2: "Bronze IV", 3: "Bronze III", 4: "Bronze II", 5: "Bronze I",
		6:  "Silver V", 7: "Silver IV", 8: "Silver III", 9: "Silver II", 10: "Silver I",
		11: "Gold V", 12: "Gold IV", 13: "Gold III", 14: "Gold II", 15: "Gold I",
		16: "Platinum V", 17: "Platinum IV", 18: "Platinum III", 19: "Platinum II", 20: "Platinum I",
		21: "Diamond V", 22: "Diamond IV", 23: "Diamond III", 24: "Diamond II", 25: "Diamond I",
		26: "Ruby V", 27: "Ruby IV", 28: "Ruby III", 29: "Ruby II", 30: "Ruby I",
	}
	if name, exists := tierNames[tier]; exists {
		return name
	}
	return fmt.Sprintf("Tier %d", tier)
}

// groupScoresByLeague 참가자들을 리그별로 분류하고 점수 순으로 정렬합니다
func groupScoresByLeague(scores []models.ScoreData) map[int][]models.ScoreData {
	leagueScores := make(map[int][]models.ScoreData)

	for _, score := range scores {
		leagueScores[score.League] = append(leagueScores[score.League], score)
	}

	// 각 리그별로 점수 순으로 정렬
	for league := range leagueScores {
		scores := leagueScores[league]
		for i := 0; i < len(scores)-1; i++ {
			for j := i + 1; j < len(scores); j++ {
				// 1. RawScore 기준 내림차순
				if scores[i].RawScore < scores[j].RawScore {
					scores[i], scores[j] = scores[j], scores[i]
				} else if scores[i].RawScore == scores[j].RawScore {
					// 2. 동점일 경우 BaekjoonID 오름차순
					if scores[i].BaekjoonID > scores[j].BaekjoonID {
						scores[i], scores[j] = scores[j], scores[i]
					}
				}
			}
		}
	}

	return leagueScores
}

// getLeagueName 리그 번호를 리그 이름으로 변환합니다
func getLeagueName(league int) string {
	switch league {
	case 0:
		return "루키"
	case 1:
		return "프로"
	case 2:
		return "마스터"
	default:
		return "알 수 없음"
	}
}
