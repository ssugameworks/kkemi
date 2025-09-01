package scoring

import (
	"discord-bot/api"
	"discord-bot/constants"
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
		startTier := 11 // Gold V
		startProblemIDs := []int{}

		score, err := calculator.CalculateScore("testuser", startTier, startProblemIDs)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Gold V (same tier): 18 * 1.0 = 18
		// Gold IV (higher tier): 20 * 1.4 = 28
		// Gold III (higher tier): 22 * 1.4 = 31 (rounded)
		// Total: 77
		expected := float64(77)
		if score != expected {
			t.Errorf("Expected score %f, got %f", expected, score)
		}
	})

	t.Run("Score calculation excluding starting problems", func(t *testing.T) {
		startTier := 11 // Gold V
		startProblemIDs := []int{1} // Exclude problem 1

		score, err := calculator.CalculateScore("testuser", startTier, startProblemIDs)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Gold IV (higher tier): 20 * 1.4 = 28
		// Gold III (higher tier): 22 * 1.4 = 31 (rounded)
		// Total: 59
		expected := float64(59)
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
					{ProblemID: 1, Level: 10}, // Silver I (16 points, penalty multiplier)
					{ProblemID: 2, Level: 11}, // Gold V (18 points, base multiplier)
					{ProblemID: 3, Level: 12}, // Gold IV (20 points, challenge multiplier)
				},
			},
			startTier:       11,
			startProblemIDs: []int{},
			expected:        54, // 16*0.5 + 18*1.0 + 20*1.4 = 8 + 18 + 28 = 54
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
			startTier:       11,
			startProblemIDs: []int{},
			expected:        18, // Only Gold V contributes: 18 * 1.0
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

func TestScoreCalculator_getWeight(t *testing.T) {
	calculator := &ScoreCalculator{}

	tests := []struct {
		name        string
		problemTier int
		startTier   int
		expected    float64
	}{
		{
			name:        "Challenge tier (higher than start)",
			problemTier: 12,
			startTier:   11,
			expected:    constants.ChallengeMultiplier,
		},
		{
			name:        "Base tier (same as start)",
			problemTier: 11,
			startTier:   11,
			expected:    constants.BaseMultiplier,
		},
		{
			name:        "Penalty tier (lower than start)",
			problemTier: 10,
			startTier:   11,
			expected:    constants.PenaltyMultiplier,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			weight := calculator.getWeight(test.problemTier, test.startTier)
			if weight != test.expected {
				t.Errorf("Expected weight %f, got %f", test.expected, weight)
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
		startTier := 11 // Gold V
		startProblemIDs := []int{1000, 1001} // Already solved Silver III and Gold V

		score, err := calculator.CalculateScore("testuser", startTier, startProblemIDs)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Only Gold IV, Gold III, Platinum V should be counted
		// Gold IV: 20 * 1.4 = 28
		// Gold III: 22 * 1.4 = 31 (rounded)  
		// Platinum V: 28 * 1.4 = 39 (rounded)
		// Total: 98
		expected := float64(98)
		if score != expected {
			t.Errorf("Expected score %f, got %f", expected, score)
		}
	})

	t.Run("Platinum starter solving lower tier problems", func(t *testing.T) {
		startTier := 16 // Platinum III
		startProblemIDs := []int{}

		score, err := calculator.CalculateScore("testuser", startTier, startProblemIDs)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// All problems are lower tier except Platinum V
		// Silver III: 12 * 0.5 = 6
		// Gold V: 18 * 0.5 = 9
		// Gold IV: 20 * 0.5 = 10
		// Gold III: 22 * 0.5 = 11
		// Platinum V: 28 * 1.0 = 28
		// Total: 64
		expected := float64(64)
		if score != expected {
			t.Errorf("Expected score %f, got %f", expected, score)
		}
	})
}