package models

import (
	"testing"
)

func TestTierManager_GetTierInfo(t *testing.T) {
	tm := NewTierManager()

	tests := []struct {
		tier           int
		expectedName   string
		expectedPoints int
	}{
		{0, "Unranked", 0},
		{1, "Bronze V", 1},
		{5, "Bronze I", 5},
		{6, "Silver V", 8},
		{10, "Silver I", 16},
		{11, "Gold V", 18},
		{15, "Gold I", 25},
		{16, "Platinum V", 28},
		{20, "Platinum I", 37},
		{21, "Diamond V", 40},
		{25, "Diamond I", 50},
		{26, "Ruby V", 55},
		{30, "Ruby I", 75},
		{31, "Master", 80},
		{35, "Master", 80}, // Above 31 should return Master
	}

	for _, test := range tests {
		t.Run(test.expectedName, func(t *testing.T) {
			tierInfo := tm.GetTierInfo(test.tier)
			if tierInfo.Name != test.expectedName {
				t.Errorf("Expected name '%s' for tier %d, got '%s'", test.expectedName, test.tier, tierInfo.Name)
			}
			if tierInfo.Points != test.expectedPoints {
				t.Errorf("Expected points %d for tier %d, got %d", test.expectedPoints, test.tier, tierInfo.Points)
			}
		})
	}
}

func TestTierManager_GetTierName(t *testing.T) {
	tm := NewTierManager()

	tests := []struct {
		tier     int
		expected string
	}{
		{0, "Unranked"},
		{11, "Gold V"},
		{20, "Platinum I"},
		{31, "Master"},
		{-1, "Unranked"}, // Invalid tier should return Unranked
	}

	for _, test := range tests {
		result := tm.GetTierName(test.tier)
		if result != test.expected {
			t.Errorf("Expected tier name '%s' for tier %d, got '%s'", test.expected, test.tier, result)
		}
	}
}

func TestTierManager_GetTierPoints(t *testing.T) {
	tm := NewTierManager()

	tests := []struct {
		tier     int
		expected int
	}{
		{0, 0},   // Unranked
		{1, 1},   // Bronze V
		{11, 18}, // Gold V
		{31, 80}, // Master
		{35, 80}, // Above Master should return Master points
		{-1, 0},  // Invalid tier should return Unranked points
	}

	for _, test := range tests {
		result := tm.GetTierPoints(test.tier)
		if result != test.expected {
			t.Errorf("Expected tier points %d for tier %d, got %d", test.expected, test.tier, result)
		}
	}
}

func TestTierManager_GetTierColor(t *testing.T) {
	tm := NewTierManager()

	tests := []struct {
		tier     int
		expected int
	}{
		{0, 0x36393F},  // Unranked
		{1, 0xA25B1F},  // Bronze V
		{6, 0x495E78},  // Silver V
		{11, 0xE09E37}, // Gold V
		{16, 0x6DDFA8}, // Platinum V
		{21, 0x50B1F6}, // Diamond V
		{26, 0xEA3364}, // Ruby V
		{31, 0x8A2BE2}, // Master
	}

	for _, test := range tests {
		result := tm.GetTierColor(test.tier)
		if result != test.expected {
			t.Errorf("Expected color 0x%X for tier %d, got 0x%X", test.expected, test.tier, result)
		}
	}
}

func TestTierManager_GetTierCategory(t *testing.T) {
	tm := NewTierManager()

	tests := []struct {
		tier     int
		expected TierCategory
	}{
		{0, CategoryUnranked},
		{1, CategoryBronze},
		{5, CategoryBronze},
		{6, CategorySilver},
		{10, CategorySilver},
		{11, CategoryGold},
		{15, CategoryGold},
		{16, CategoryPlatinum},
		{20, CategoryPlatinum},
		{21, CategoryDiamond},
		{25, CategoryDiamond},
		{26, CategoryRuby},
		{30, CategoryRuby},
		{31, CategoryMaster},
		{35, CategoryMaster},
	}

	for _, test := range tests {
		result := tm.GetTierCategory(test.tier)
		if result != test.expected {
			t.Errorf("Expected category %d for tier %d, got %d", test.expected, test.tier, result)
		}
	}
}

func TestTierManager_GetANSIColor(t *testing.T) {
	tm := NewTierManager()

	tests := []struct {
		tier     int
		expected string
	}{
		{0, "\x1b[0m"},     // Unranked
		{1, "\x1b[1;33m"},  // Bronze V
		{6, "\x1b[1;37m"},  // Silver V
		{11, "\x1b[1;33m"}, // Gold V
		{16, "\x1b[1;36m"}, // Platinum V
		{21, "\x1b[1;34m"}, // Diamond V
		{26, "\x1b[1;31m"}, // Ruby V
		{31, "\x1b[1;35m"}, // Master
	}

	for _, test := range tests {
		result := tm.GetTierANSIColor(test.tier)
		if result != test.expected {
			t.Errorf("Expected ANSI color '%s' for tier %d, got '%s'", test.expected, test.tier, result)
		}
	}
}

func TestTierManager_GetANSIReset(t *testing.T) {
	tm := NewTierManager()
	expected := "\x1b[0m"
	result := tm.GetANSIReset()
	if result != expected {
		t.Errorf("Expected ANSI reset '%s', got '%s'", expected, result)
	}
}

func TestTierManager_Singleton(t *testing.T) {
	// 두 개의 인스턴스가 같은 객체인지 확인
	tm1 := GetTierManager()
	tm2 := GetTierManager()

	if tm1 != tm2 {
		t.Error("TierManager should be a singleton")
	}

	// 같은 데이터를 반환하는지 확인
	name1 := tm1.GetTierName(11)
	name2 := tm2.GetTierName(11)
	if name1 != name2 {
		t.Errorf("Both instances should return same name, got '%s' and '%s'", name1, name2)
	}
}

func TestTierManager_EdgeCases(t *testing.T) {
	tm := NewTierManager()

	t.Run("Negative tier", func(t *testing.T) {
		tierInfo := tm.GetTierInfo(-1)
		if tierInfo.Name != "Unranked" {
			t.Errorf("Expected 'Unranked' for negative tier, got '%s'", tierInfo.Name)
		}
	})

	t.Run("Very high tier", func(t *testing.T) {
		tierInfo := tm.GetTierInfo(100)
		if tierInfo.Name != "Master" {
			t.Errorf("Expected 'Master' for very high tier, got '%s'", tierInfo.Name)
		}
	})

	t.Run("Boundary tiers", func(t *testing.T) {
		// Test boundaries between categories
		boundaryTests := []struct {
			tier     int
			category TierCategory
		}{
			{5, CategoryBronze},    // Highest Bronze
			{6, CategorySilver},    // Lowest Silver
			{10, CategorySilver},   // Highest Silver
			{11, CategoryGold},     // Lowest Gold
			{15, CategoryGold},     // Highest Gold
			{16, CategoryPlatinum}, // Lowest Platinum
		}

		for _, test := range boundaryTests {
			category := tm.GetTierCategory(test.tier)
			if category != test.category {
				t.Errorf("Expected category %d for tier %d, got %d", test.category, test.tier, category)
			}
		}
	})
}
