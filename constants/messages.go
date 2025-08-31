package constants

// 사용자 인터페이스 메시지
const (
	// 등록 관련
	MsgRegisterSuccess        = "%s%s(%s)%s님 성공적으로 등록되었습니다!"
	MsgRegisterUsage          = "사용법: `!등록 <이름> <백준ID>`"
	MsgRegisterNotStarted     = "대회가 아직 시작되지 않았습니다. 등록은 %s부터 가능합니다."
	MsgRegisterNoSolvedacName = "solved.ac에 이름이 등록되지 않았습니다. solved.ac 프로필에서 본명 또는 영문 이름을 등록한 후 다시 시도해주세요."
	MsgRegisterNameMismatch   = "입력한 이름 '%s'이 solved.ac에 등록된 이름 '%s'과 일치하지 않습니다."

	// 스코어보드 관련
	MsgScoreboardTitle           = "🏆 %s 스코어보드"
	MsgScoreboardDMOnly          = "❌ 스코어보드는 서버에서만 확인할 수 있습니다."
	MsgScoreboardBlackout        = "🔒 스코어보드 비공개"
	MsgScoreboardBlackoutDesc    = "마지막 3일간 스코어보드가 비공개됩니다"
	MsgScoreboardNoParticipants  = "참가자가 없습니다."
	MsgScoreboardNoScores        = "아직 점수가 계산된 참가자가 없습니다."
	MsgScoreboardBlackoutWarning = "⚠️ %d일 후 스코어보드가 비공개됩니다"

	// 참가자 관련
	MsgParticipantsEmpty = "참가자가 없습니다."

	// 삭제 관련
	MsgRemoveSuccess           = "**참가자 삭제 완료**\n🎯 백준ID: %s"
	MsgRemoveUsage             = "사용법: `!삭제 <백준ID>`"
	MsgRemoveInvalidBaekjoonID = "유효하지 않은 백준 ID 형식입니다."

	// 권한 관련
	MsgInsufficientPermissions = "❌ 관리자 권한이 필요합니다."

	// 기본 응답
	MsgPong = "Pong! 🏓"

	// 대회 관리 관련
	MsgCompetitionCreateUsage   = "사용법: `!대회 create <대회명> <시작일> <종료일>` (날짜 형식: YYYY-MM-DD)"
	MsgCompetitionCreateSuccess = "**대회 생성 완료**\n🏆 대회명: %s\n📅 기간: %s ~ %s\n🔒 블랙아웃: %s부터"
	MsgCompetitionUpdateSuccess = "**대회 정보 수정 완료**\n🎯 수정 항목: %s"
	MsgCompetitionStatus        = "🏆 **대회 정보**\n📝 대회명: %s\n📅 시작일: %s\n📅 종료일: %s\n🔒 블랙아웃: %s\n📊 스코어보드: %s\n👥 참가자: %d명"

	// 상태 표시
	StatusActive   = "활성"
	StatusInactive = "비활성"
	StatusVisible  = "공개"
	StatusHidden   = "비공개"
)

// 도움말 메시지
const HelpMessage = `🤖 **알고리즘 경진대회 봇 명령어**

**참가자 명령어:**
• ` + "`!등록 <이름> <백준ID>`" + ` - 대회 등록 신청 (대회 시작 후, solved.ac 등록 이름과 일치해야 함)

**관리자 명령어:**
• ` + "`!스코어보드`" + ` - 현재 스코어보드 확인
• ` + "`!참가자`" + ` - 참가자 목록 확인
• ` + "`!대회 create <대회명> <시작일> <종료일>`" + ` - 대회 생성 (YYYY-MM-DD 형식)
• ` + "`!대회 status`" + ` - 대회 상태 확인
• ` + "`!대회 blackout <on/off>`" + ` - 스코어보드 공개/비공개 설정
• ` + "`!대회 update <필드> <값>`" + ` - 대회 정보 수정 (name, start, end)
• ` + "`!삭제 <백준ID>`" + ` - 참가자 삭제

**기타:**
• ` + "`!ping`" + ` - 봇 응답 확인
• ` + "`!도움말`" + ` - 도움말 표시`

// 에러 코드와 메시지 매핑
var ErrorMessages = map[string]string{
	"REGISTER_INVALID_PARAMS":           MsgRegisterUsage,
	"REGISTRATION_NOT_STARTED":          "대회가 아직 시작되지 않았습니다.",
	"NO_SOLVEDAC_NAME":                  MsgRegisterNoSolvedacName,
	"NAME_MISMATCH":                     "입력한 이름이 solved.ac 이름과 일치하지 않습니다.",
	"REMOVE_INVALID_PARAMS":             MsgRemoveUsage,
	"REMOVE_INVALID_BAEKJOON_ID":        MsgRemoveInvalidBaekjoonID,
	"COMPETITION_CREATE_INVALID_PARAMS": MsgCompetitionCreateUsage,
	"INSUFFICIENT_PERMISSIONS":          MsgInsufficientPermissions,
}
