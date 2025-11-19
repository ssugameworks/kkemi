package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ssugameworks/kkemi/constants"
	"github.com/ssugameworks/kkemi/utils"
)

// SolvedACClient solved.ac API와 통신하는 클라이언트입니다
type SolvedACClient struct {
	client  *http.Client
	baseURL string
}

// UserInfo solved.ac 사용자 정보를 나타냅니다
type UserInfo struct {
	Handle          string `json:"handle"`
	Bio             string `json:"bio"`
	Rating          int    `json:"rating"`
	Tier            int    `json:"tier"`
	Class           int    `json:"class"`
	ClassDecoration string `json:"classDecoration"`
	ProfileImageURL string `json:"profileImageUrl"`
	SolvedCount     int    `json:"solvedCount"`
	Verified        bool   `json:"verified"`
	Rank            int    `json:"rank"`
}

// ProblemInfo solved.ac 문제 정보를 나타냅니다
type ProblemInfo struct {
	ProblemID         int     `json:"problemId"`
	Level             int     `json:"level"`
	TitleKo           string  `json:"titleKo"`
	AcceptedUserCount int     `json:"acceptedUserCount"`
	AverageTries      float64 `json:"averageTries"`
}

// Top100Response 사용자의 TOP 100 문제 응답을 나타냅니다
type Top100Response struct {
	Count int           `json:"count"`
	Items []ProblemInfo `json:"items"`
}

// UserAdditionalInfo 사용자의 추가 정보를 나타냅니다
type UserAdditionalInfo struct {
	CountryCode *string `json:"countryCode"`
	Gender      int     `json:"gender"`
	Pronouns    *string `json:"pronouns"`
	BirthYear   *int    `json:"birthYear"`
	BirthMonth  *int    `json:"birthMonth"`
	BirthDay    *int    `json:"birthDay"`
	Name        *string `json:"name"`
	NameNative  *string `json:"nameNative"`
}

// Organization solved.ac 조직 정보를 나타냅니다
type Organization struct {
	OrganizationID int    `json:"organizationId"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Rating         int    `json:"rating"`
	UserCount      int    `json:"userCount"`
	VoteCount      int    `json:"voteCount"`
	SolvedCount    int    `json:"solvedCount"`
	Color          string `json:"color"`
}

// NewSolvedACClient 새로운 SolvedACClient 인스턴스를 생성합니다
func NewSolvedACClient() *SolvedACClient {
	utils.Debug("Creating new SolvedAC API client")
	return &SolvedACClient{
		client: &http.Client{
			Timeout: constants.APITimeout,
		},
		baseURL: constants.SolvedACBaseURL,
	}
}

// GetUserInfo 지정된 핸들의 사용자 정보를 가져옵니다
func (client *SolvedACClient) GetUserInfo(ctx context.Context, handle string) (*UserInfo, error) {
	if !utils.IsValidBaekjoonID(handle) {
		return nil, fmt.Errorf("잘못된 핸들 형식: %s", handle)
	}

	url := fmt.Sprintf("%s/user/show?handle=%s", client.baseURL, handle)
	return client.getUserInfoWithRetry(ctx, url, handle)
}

// doRequest 공통 HTTP 요청 및 재시도 로직
func (client *SolvedACClient) doRequest(ctx context.Context, url, requestType, handle string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < constants.MaxRetries; attempt++ {
		if attempt > 0 {
			utils.Debug("Retrying %s fetch for %s (attempt %d/%d)", requestType, handle, attempt+1, constants.MaxRetries)
			time.Sleep(constants.RetryDelay * time.Duration(attempt))
		}

		utils.Debug("Fetching %s from: %s", requestType, url)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = fmt.Errorf("요청 생성 실패: %w", err)
			continue
		}

		resp, err := client.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("%s 조회 실패: %w", requestType, err)
			utils.Warn("Attempt %d failed for %s %s: %v", attempt+1, requestType, handle, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("요청 한도 초과")
			utils.Warn("Rate limited for %s %s, attempt %d", requestType, handle, attempt+1)
			time.Sleep(constants.RetryDelay * constants.APIRetryMultiplier)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API가 상태 코드 %d를 반환했습니다", resp.StatusCode)
			utils.Warn("API returned non-200 status for %s %s: %d", requestType, handle, resp.StatusCode)
			if resp.StatusCode >= constants.HTTPServerErrorThreshold {
				continue // 서버 에러는 재시도
			}
			break // 클라이언트 에러는 즉시 반환
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("응답 읽기 실패: %w", err)
			utils.Error("Failed to read %s response body for %s: %v", requestType, handle, err)
			continue
		}

		return body, nil
	}

	utils.Error("Failed to fetch %s for %s after %d attempts: %v", requestType, handle, constants.MaxRetries, lastErr)
	return nil, lastErr
}

// 재시도 로직을 포함한 사용자 정보 조회
func (client *SolvedACClient) getUserInfoWithRetry(ctx context.Context, url, handle string) (*UserInfo, error) {
	body, err := client.doRequest(ctx, url, "user info", handle)
	if err != nil {
		return nil, err
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		utils.Error("Failed to parse user info for %s: %v", handle, err)
		return nil, fmt.Errorf("사용자 정보 파싱 실패: %w", err)
	}

	utils.Debug("Successfully fetched user info for %s (tier: %d, rating: %d)",
		handle, userInfo.Tier, userInfo.Rating)
	return &userInfo, nil
}

// GetUserTop100 지정된 사용자의 TOP 100 문제를 가져옵니다
func (client *SolvedACClient) GetUserTop100(ctx context.Context, handle string) (*Top100Response, error) {
	if !utils.IsValidBaekjoonID(handle) {
		return nil, fmt.Errorf("잘못된 핸들 형식: %s", handle)
	}

	url := fmt.Sprintf("%s/user/top_100?handle=%s", client.baseURL, handle)
	return client.getUserTop100WithRetry(ctx, url, handle)
}

// 재시도 로직을 포함한 TOP 100 조회
func (client *SolvedACClient) getUserTop100WithRetry(ctx context.Context, url, handle string) (*Top100Response, error) {
	body, err := client.doRequest(ctx, url, "top 100", handle)
	if err != nil {
		return nil, err
	}

	var top100 Top100Response
	if err := json.Unmarshal(body, &top100); err != nil {
		utils.Error("Failed to parse top 100 for %s: %v", handle, err)
		return nil, fmt.Errorf("TOP 100 파싱 실패: %w", err)
	}

	utils.Debug("Successfully fetched %d top problems for %s", top100.Count, handle)
	return &top100, nil
}

// GetUserAdditionalInfo 지정된 사용자의 추가 정보를 가져옵니다
func (client *SolvedACClient) GetUserAdditionalInfo(ctx context.Context, handle string) (*UserAdditionalInfo, error) {
	if !utils.IsValidBaekjoonID(handle) {
		return nil, fmt.Errorf("잘못된 핸들 형식: %s", handle)
	}

	url := fmt.Sprintf("%s/user/additional_info?handle=%s", client.baseURL, handle)
	return client.getUserAdditionalInfoWithRetry(ctx, url, handle)
}

// 재시도 로직을 포함한 사용자 추가 정보 조회
func (client *SolvedACClient) getUserAdditionalInfoWithRetry(ctx context.Context, url, handle string) (*UserAdditionalInfo, error) {
	body, err := client.doRequest(ctx, url, "additional info", handle)
	if err != nil {
		return nil, err
	}

	var additionalInfo UserAdditionalInfo
	// API 응답 디버깅
	utils.Debug("Raw API response for additional info %s: %s", handle, string(body))

	if err := json.Unmarshal(body, &additionalInfo); err != nil {
		utils.Error("Failed to parse additional info for %s: %v", handle, err)
		return nil, fmt.Errorf("추가 정보 파싱 실패: %w", err)
	}

	var nameNativeStr string
	if additionalInfo.NameNative != nil {
		nameNativeStr = *additionalInfo.NameNative
	}
	utils.Debug("Successfully fetched additional info for %s (nameNative: %s)",
		handle, nameNativeStr)
	return &additionalInfo, nil
}

// GetUserOrganizations 지정된 사용자의 소속 조직 목록을 가져옵니다
func (client *SolvedACClient) GetUserOrganizations(ctx context.Context, handle string) ([]Organization, error) {
	if !utils.IsValidBaekjoonID(handle) {
		return nil, fmt.Errorf("잘못된 핸들 형식: %s", handle)
	}

	url := fmt.Sprintf("%s/user/organizations?handle=%s", client.baseURL, handle)
	return client.getUserOrganizationsWithRetry(ctx, url, handle)
}

// 재시도 로직을 포함한 사용자 조직 목록 조회
func (client *SolvedACClient) getUserOrganizationsWithRetry(ctx context.Context, url, handle string) ([]Organization, error) {
	body, err := client.doRequest(ctx, url, "user organizations", handle)
	if err != nil {
		return nil, err
	}

	var organizations []Organization
	if err := json.Unmarshal(body, &organizations); err != nil {
		utils.Error("Failed to parse organizations for %s: %v", handle, err)
		return nil, fmt.Errorf("조직 정보 파싱 실패: %w", err)
	}

	utils.Debug("Successfully fetched %d organizations for %s", len(organizations), handle)
	return organizations, nil
}
