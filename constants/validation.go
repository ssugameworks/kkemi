package constants

// 검증 관련 상수
const (
	// 사용자명 검증
	MaxUsernameDisplayWidth = 40 // 사용자명 최대 표시 너비
	MaxCharacterRepeats     = 5  // 허용되는 최대 문자 반복 횟수

	// 시간대 관련
	KSTOffsetSeconds = 9 * 60 * 60 // 한국 표준시(KST) UTC 오프셋 (초)

	// HTTP 관련
	HTTPServerErrorThreshold = 500 // 서버 오류 임계값 (5xx)

	// 제어 문자 관련
	ControlCharTab = 9  // 탭 문자
	ControlCharLF  = 10 // 줄 바꿈
	ControlCharCR  = 13 // 캐리지 리턴
	ControlCharMin = 32 // 허용되는 최소 제어 문자

	// 유니코드 범위 - 한글 문자 너비 계산용
	UnicodeHangulJamoStart         = 0x1100
	UnicodeHangulJamoEnd           = 0x11FF
	UnicodeHangulCompatStart       = 0x3130
	UnicodeHangulCompatEnd         = 0x318F
	UnicodeHangulSyllableStart     = 0xAC00
	UnicodeHangulSyllableEnd       = 0xD7AF
	UnicodeCJKStart                = 0x4E00
	UnicodeCJKEnd                  = 0x9FFF
	UnicodeFullwidthPrintableStart = 0xFF01 // 전각 인쇄 가능 문자 시작
	UnicodeFullwidthPrintableEnd   = 0xFF5E // 전각 인쇄 가능 문자 끝
)

// SQL 인젝션 및 악성 패턴 목록
var SecurityMaliciousPatterns = []string{
	"union", "select", "insert", "update", "delete", "drop", "create", "alter",
	"exec", "execute", "script", "javascript", "vbscript", "onload", "onerror",
	"<script", "</script>", "eval(", "alert(", "document.cookie",
	"--", "/*", "*/", ";--", "';", "\";", "')", "\")",
}

// 예약어 목록 - Baekjoon ID 검증용
var ReservedBaekjoonIDs = []string{
	"admin", "administrator", "root", "system", "null", "undefined",
	"test", "demo", "sample", "guest", "anonymous", "user",
	"baekjoon", "boj", "solved", "acmicpc",
}
