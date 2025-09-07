package scoring

import (
	"discord-bot/api"
	"discord-bot/models"
	"fmt"
	"testing"
)

type mockAPIClient struct {
	userInfo *api.UserInfo
	top100   *api.Top100Response
	err      error
}

func (m *mockAPIClient) GetUserInfo(handle string) (*api.UserInfo, error) {
	return m.userInfo, m.err
}

func (m *mockAPIClient) GetUserTop100(handle string) (*api.Top100Response, error) {
	return m.top100, m.err
}

func (m *mockAPIClient) GetUserAdditionalInfo(handle string) (*api.UserAdditionalInfo, error) {
	return nil, nil
}

func TestScoreCalculator_CalculateScore(t *testing.T) {
	tierManager := models.NewTierManager()

	mockTop100 := &api.Top100Response{
		Count: 3,
		Items: []api.ProblemInfo{
			{ProblemID: 1, Level: 11}, // Gold V (1100 points)
			{ProblemID: 2, Level: 12}, // Gold IV (1300 points)
			{ProblemID: 3, Level: 13}, // Gold III (1600 points)
		},
	}

	mockClient := &mockAPIClient{
		top100: mockTop100,
	}

	calculator := NewScoreCalculator(mockClient, tierManager)

	t.Run("Basic score calculation with no starting problems", func(t *testing.T) {
		startTier := 11 // Gold V (프로 리그)
		startProblemIDs := []int{}

		score, err := calculator.CalculateScore("testuser", startTier, startProblemIDs)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Gold V (Level 11, same tier): 11 * 1.0 = 11 (프로 리그 동일 티어)
		// Gold IV (Level 12, higher tier): 12 * 1.2 = 14.4 (프로 리그 상위 티어)
		// Gold III (Level 13, higher tier): 13 * 1.2 = 15.6 (프로 리그 상위 티어)
		// Total: 11 + 14.4 + 15.6 = 41
		expected := float64(41)
		if score != expected {
			t.Errorf("Expected score %f, got %f", expected, score)
		}
	})

	t.Run("Score calculation excluding starting problems", func(t *testing.T) {
		startTier := 11             // Gold V (프로 리그)
		startProblemIDs := []int{1} // Exclude problem 1 (Gold V)

		score, err := calculator.CalculateScore("testuser", startTier, startProblemIDs)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Gold IV (Level 12, higher tier): 12 * 1.2 = 14.4 (프로 리그 상위 티어)
		// Gold III (Level 13, higher tier): 13 * 1.2 = 15.6 (프로 리그 상위 티어)
		// Total: 14.4 + 15.6 = 30
		expected := float64(30)
		if score != expected {
			t.Errorf("Expected score %f, got %f", expected, score)
		}
	})

	t.Run("API error handling", func(t *testing.T) {
		mockClient := &mockAPIClient{
			err: fmt.Errorf("API error"),
		}
		calculator := NewScoreCalculator(mockClient, tierManager)

		_, err := calculator.CalculateScore("testuser", 11, []int{})
		if err == nil {
			t.Error("Expected error but got nil")
		}
	})
}

func TestScoreCalculator_CalculateScoreWithTop100(t *testing.T) {
	tierManager := models.NewTierManager()
	calculator := &ScoreCalculator{
		tierManager: tierManager,
	}

	tests := []struct {
		name            string
		top100          *api.Top100Response
		startTier       int
		startProblemIDs []int
		expected        float64
	}{
		{
			name: "Empty problems list",
			top100: &api.Top100Response{
				Count: 0,
				Items: []api.ProblemInfo{},
			},
			startTier:       11,
			startProblemIDs: []int{},
			expected:        0,
		},
		{
			name: "Problems with different multipliers",
			top100: &api.Top100Response{
				Count: 3,
				Items: []api.ProblemInfo{
					{ProblemID: 1, Level: 10}, // Silver I (프로 리그 하위 티어)
					{ProblemID: 2, Level: 11}, // Gold V (프로 리그 동일 티어)
					{ProblemID: 3, Level: 12}, // Gold IV (프로 리그 상위 티어)
				},
			},
			startTier:       11, // Gold V (프로 리그)
			startProblemIDs: []int{},
			expected:        33, // 10*0.8 + 11*1.0 + 12*1.2 = 8 + 11 + 14.4 = 33.4 → 33 (반올림)
		},
		{
			name: "All problems excluded by starting problems",
			top100: &api.Top100Response{
				Count: 2,
				Items: []api.ProblemInfo{
					{ProblemID: 1, Level: 11},
					{ProblemID: 2, Level: 12},
				},
			},
			startTier:       11,
			startProblemIDs: []int{1, 2},
			expected:        0,
		},
		{
			name: "Problems with zero points (unranked)",
			top100: &api.Top100Response{
				Count: 2,
				Items: []api.ProblemInfo{
					{ProblemID: 1, Level: 0},  // Unranked
					{ProblemID: 2, Level: 11}, // Gold V
				},
			},
			startTier:       11, // Gold V (프로 리그)
			startProblemIDs: []int{},
			expected:        11, // Only Gold V contributes: 11 * 1.0 (프로 리그 동일 티어)
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			score := calculator.CalculateScoreWithTop100(test.top100, test.startTier, test.startProblemIDs)
			if score != test.expected {
				t.Errorf("Expected score %f, got %f", test.expected, score)
			}
		})
	}
}

func TestScoreCalculator_Integration(t *testing.T) {
	tierManager := models.NewTierManager()

	// 실제 데이터와 유사한 시나리오 테스트
	top100 := &api.Top100Response{
		Count: 5,
		Items: []api.ProblemInfo{
			{ProblemID: 1000, Level: 8},  // Silver III (12 points)
			{ProblemID: 1001, Level: 11}, // Gold V (18 points)
			{ProblemID: 1002, Level: 12}, // Gold IV (20 points)
			{ProblemID: 1003, Level: 13}, // Gold III (22 points)
			{ProblemID: 1004, Level: 16}, // Platinum V (28 points)
		},
	}

	mockClient := &mockAPIClient{
		top100: top100,
	}

	calculator := NewScoreCalculator(mockClient, tierManager)

	t.Run("Gold V starter with some starting problems", func(t *testing.T) {
		startTier := 11                      // Gold V (프로 리그)
		startProblemIDs := []int{1000, 1001} // Already solved Silver III and Gold V

		score, err := calculator.CalculateScore("testuser", startTier, startProblemIDs)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Only Gold IV, Gold III, Platinum V should be counted (프로 리그 가중치)
		// Gold IV (Level 12, higher): 12 * 1.2 = 14.4
		// Gold III (Level 13, higher): 13 * 1.2 = 15.6
		// Platinum V (Level 16, higher): 16 * 1.2 = 19.2
		// Total: 14.4 + 15.6 + 19.2 = 49.2 → 49 (반올림)
		expected := float64(49)
		if score != expected {
			t.Errorf("Expected score %f, got %f", expected, score)
		}
	})

	t.Run("Platinum starter solving lower tier problems", func(t *testing.T) {
		startTier := 16 // Platinum V (맥스 리그)
		startProblemIDs := []int{}

		score, err := calculator.CalculateScore("testuser", startTier, startProblemIDs)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// 맥스 리그는 모든 문제에 동일 가중치 (1.0) 적용
		// Silver III (Level 8): 8 * 1.0 = 8
		// Gold V (Level 11): 11 * 1.0 = 11
		// Gold IV (Level 12): 12 * 1.0 = 12
		// Gold III (Level 13): 13 * 1.0 = 13
		// Platinum V (Level 16): 16 * 1.0 = 16
		// Total: 8 + 11 + 12 + 13 + 16 = 60
		expected := float64(60)
		if score != expected {
			t.Errorf("Expected score %f, got %f", expected, score)
		}
	})
}
