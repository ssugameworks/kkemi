package models

import "sync"

// TierInfo 특정 티어에 대한 모든 정보를 포함합니다
type TierInfo struct {
	Level     int    // 티어 레벨 (1-31+)
	Name      string // 표시 이름 (예: "Bronze V", "Gold I")
	Points    int    // 해당 티어 문제 해결 시 기본 점수
	ColorCode int    // Discord embed 색상 코드
	ANSIColor string // 터미널 표시용 ANSI 색상 코드
}

// TierCategory 주요 티어 카테고리를 나타냅니다
type TierCategory int

const (
	CategoryUnranked TierCategory = iota // 언랭크
	CategoryBronze                       // 브론즈
	CategorySilver                       // 실버
	CategoryGold                         // 골드
	CategoryPlatinum                     // 플래티넘
	CategoryDiamond                      // 다이아몬드
	CategoryRuby                         // 루비
	CategoryMaster                       // 마스터
)

// TierManager 모든 티어 관련 기능을 관리합니다
type TierManager struct {
	tiers map[int]*TierInfo
}

var (
	globalTierManager *TierManager
	once              sync.Once
)

// GetTierManager 전역 TierManager 인스턴스를 반환합니다 (싱글톤)
func GetTierManager() *TierManager {
	once.Do(func() {
		globalTierManager = &TierManager{
			tiers: make(map[int]*TierInfo),
		}
		globalTierManager.initializeTiers()
	})
	return globalTierManager
}

// NewTierManager 새로운 TierManager 인스턴스를 생성합니다
// Deprecated: GetTierManager() 사용을 권장합니다
func NewTierManager() *TierManager {
	return GetTierManager()
}

// initializeTiers 티어 정보를 초기화합니다
func (tm *TierManager) initializeTiers() {
	// 언랭크
	tm.tiers[0] = &TierInfo{0, "Unranked", 0, 0x36393F, "\x1b[0m"}

	// 브론즈 (1-5)
	tm.tiers[1] = &TierInfo{1, "Bronze V", 1, 0xA25B1F, "\x1b[1;33m"}
	tm.tiers[2] = &TierInfo{2, "Bronze IV", 2, 0xA25B1F, "\x1b[1;33m"}
	tm.tiers[3] = &TierInfo{3, "Bronze III", 3, 0xA25B1F, "\x1b[1;33m"}
	tm.tiers[4] = &TierInfo{4, "Bronze II", 4, 0xA25B1F, "\x1b[1;33m"}
	tm.tiers[5] = &TierInfo{5, "Bronze I", 5, 0xA25B1F, "\x1b[1;33m"}

	// 실버 (6-10)
	tm.tiers[6] = &TierInfo{6, "Silver V", 8, 0x495E78, "\x1b[1;37m"}
	tm.tiers[7] = &TierInfo{7, "Silver IV", 10, 0x495E78, "\x1b[1;37m"}
	tm.tiers[8] = &TierInfo{8, "Silver III", 12, 0x495E78, "\x1b[1;37m"}
	tm.tiers[9] = &TierInfo{9, "Silver II", 14, 0x495E78, "\x1b[1;37m"}
	tm.tiers[10] = &TierInfo{10, "Silver I", 16, 0x495E78, "\x1b[1;37m"}

	// 골드 (11-15)
	tm.tiers[11] = &TierInfo{11, "Gold V", 18, 0xE09E37, "\x1b[1;33m"}
	tm.tiers[12] = &TierInfo{12, "Gold IV", 20, 0xE09E37, "\x1b[1;33m"}
	tm.tiers[13] = &TierInfo{13, "Gold III", 22, 0xE09E37, "\x1b[1;33m"}
	tm.tiers[14] = &TierInfo{14, "Gold II", 23, 0xE09E37, "\x1b[1;33m"}
	tm.tiers[15] = &TierInfo{15, "Gold I", 25, 0xE09E37, "\x1b[1;33m"}

	// 플래티넘 (16-20)
	tm.tiers[16] = &TierInfo{16, "Platinum V", 28, 0x6DDFA8, "\x1b[1;36m"}
	tm.tiers[17] = &TierInfo{17, "Platinum IV", 30, 0x6DDFA8, "\x1b[1;36m"}
	tm.tiers[18] = &TierInfo{18, "Platinum III", 32, 0x6DDFA8, "\x1b[1;36m"}
	tm.tiers[19] = &TierInfo{19, "Platinum II", 35, 0x6DDFA8, "\x1b[1;36m"}
	tm.tiers[20] = &TierInfo{20, "Platinum I", 37, 0x6DDFA8, "\x1b[1;36m"}

	// 다이아몬드 (21-25)
	tm.tiers[21] = &TierInfo{21, "Diamond V", 40, 0x50B1F6, "\x1b[1;34m"}
	tm.tiers[22] = &TierInfo{22, "Diamond IV", 42, 0x50B1F6, "\x1b[1;34m"}
	tm.tiers[23] = &TierInfo{23, "Diamond III", 45, 0x50B1F6, "\x1b[1;34m"}
	tm.tiers[24] = &TierInfo{24, "Diamond II", 47, 0x50B1F6, "\x1b[1;34m"}
	tm.tiers[25] = &TierInfo{25, "Diamond I", 50, 0x50B1F6, "\x1b[1;34m"}

	// 루비 (26-30)
	tm.tiers[26] = &TierInfo{26, "Ruby V", 55, 0xEA3364, "\x1b[1;31m"}
	tm.tiers[27] = &TierInfo{27, "Ruby IV", 60, 0xEA3364, "\x1b[1;31m"}
	tm.tiers[28] = &TierInfo{28, "Ruby III", 65, 0xEA3364, "\x1b[1;31m"}
	tm.tiers[29] = &TierInfo{29, "Ruby II", 70, 0xEA3364, "\x1b[1;31m"}
	tm.tiers[30] = &TierInfo{30, "Ruby I", 75, 0xEA3364, "\x1b[1;31m"}

	// 마스터 (31+)
	tm.tiers[31] = &TierInfo{31, "Master", 80, 0x8A2BE2, "\x1b[1;35m"}
}

// GetTierInfo 티어의 완전한 정보를 반환합니다
func (tm *TierManager) GetTierInfo(tier int) *TierInfo {
	if info, exists := tm.tiers[tier]; exists {
		return info
	}
	// 31보다 높은 티어는 마스터로 처리
	if tier > 31 {
		return tm.tiers[31]
	}
	// 기본적으로 언랭크 반환
	return tm.tiers[0]
}

// GetTierName 티어의 표시 이름을 반환합니다
func (tm *TierManager) GetTierName(tier int) string {
	return tm.GetTierInfo(tier).Name
}

// GetTierPoints 티어의 기본 점수를 반환합니다
func (tm *TierManager) GetTierPoints(tier int) int {
	return tm.GetTierInfo(tier).Points
}

// GetTierColor 티어의 Discord embed 색상을 반환합니다
func (tm *TierManager) GetTierColor(tier int) int {
	return tm.GetTierInfo(tier).ColorCode
}

// GetTierANSIColor 티어의 ANSI 색상 코드를 반환합니다
func (tm *TierManager) GetTierANSIColor(tier int) string {
	return tm.GetTierInfo(tier).ANSIColor
}

// GetTierCategory 티어의 주요 카테고리를 반환합니다
func (tm *TierManager) GetTierCategory(tier int) TierCategory {
	switch {
	case tier == 0:
		return CategoryUnranked
	case tier >= 1 && tier <= 5:
		return CategoryBronze
	case tier >= 6 && tier <= 10:
		return CategorySilver
	case tier >= 11 && tier <= 15:
		return CategoryGold
	case tier >= 16 && tier <= 20:
		return CategoryPlatinum
	case tier >= 21 && tier <= 25:
		return CategoryDiamond
	case tier >= 26 && tier <= 30:
		return CategoryRuby
	default:
		return CategoryMaster
	}
}

// GetANSIReset ANSI 리셋 코드를 반환합니다
func (tm *TierManager) GetANSIReset() string {
	return "\x1b[0m"
}
