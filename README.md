# 알고리즘 경진대회 디스코드 봇

백준 알고리즘 문제 풀이 점수를 집계하여 스코어보드를 제공하는 디스코드 봇입니다.

## ✨ 주요 기능

- 🎯 **참가자 관리**: 백준 사용자 자동 등록 및 티어 확인
- 📊 **실시간 집계**: solved.ac API를 활용한 병렬 점수 계산  
- 🔒 **블랙아웃 모드**: 대회 종료 전 N일간 스코어보드 비공개
- ⚡ **차등 점수**: 도전/기본/연습 문제에 따른 차등 점수
- 🛠️ **대회 관리**: 대회 생성, 수정, 상태 관리 기능
- ⏰ **자동화**: 설정 가능한 시간에 자동 스코어보드 전송
- 💬 **다중 채널**: DM 및 서버 채널 지원
- 🚀 **고성능**: 병렬 API 호출로 빠른 응답 시간
- 🛡️ **안정성**: 포괄적인 에러 처리 및 복구 시스템

## 점수 계산 방식

### 점수 공식
```
최종점수 = Σ(문제 난이도 점수 × 가중치)
```

### 가중치 적용
- **도전 문제** (현재 티어보다 높은 문제): 1.4배
- **기본 문제** (현재 티어와 같은 문제): 1.0배  
- **연습 문제** (현재 티어보다 낮은 문제): 0.5배

### 난이도별 점수표
| 티어 | 점수 | 티어 | 점수 | 티어 | 점수 |
|------|------|------|------|------|------|
| Bronze V-I | 1-5점 | Silver V-I | 8-16점 | Gold V-I | 18-25점 |
| Platinum V-I | 28-37점 | Diamond V-I | 40-50점 | Ruby V-I | 55-75점 |

## 설치 및 실행

### 1. 환경 설정
`.env` 파일을 생성하고 다음 내용을 추가하세요:

```bash
# Discord Bot Configuration (필수)
export DISCORD_BOT_TOKEN="your_discord_bot_token_here"
export DISCORD_CHANNEL_ID="your_channel_id_here"

# Scoreboard Schedule Configuration (선택사항)
export SCOREBOARD_HOUR="9"      # 스코어보드 전송 시간 (0-23)
export SCOREBOARD_MINUTE="0"    # 스코어보드 전송 분 (0-59)

# 기타 설정 (선택사항)
export LOG_LEVEL="INFO"         # 로그 레벨 (DEBUG, INFO, WARN, ERROR)
export DEBUG_MODE="false"       # 디버그 모드
```

### 2. 의존성 설치
```bash
go mod tidy
```

### 3. 환경 변수 로드 및 봇 실행
```bash
# 환경 변수 로드
source .env

# 봇 실행
go run main.go
```

## Discord Bot 설정

### Bot 권한 설정
Discord Developer Portal에서 봇 생성 시 다음 권한이 필요합니다:
- `Send Messages` - 메시지 전송
- `Read Message History` - 메시지 기록 읽기
- `View Channels` - 채널 보기

### 인텐트 설정
다음 인텐트들이 활성화되어야 합니다:
- `Message Content Intent` - 메시지 내용 읽기
- `Server Members Intent` - 서버 멤버 정보

## 사용법

### 참가자 명령어 (모든 사용자)
- `!등록 <이름> <백준ID>` - 대회 등록 신청 (대회 시작 후, solved.ac 등록 이름과 일치해야 함)
- `!도움말` - 도움말 표시
- `!ping` - 봇 응답 확인

### 관리자 전용 명령어 (관리자 권한 필요)
- `!스코어보드` - 현재 스코어보드 확인
- `!참가자` - 참가자 목록 확인
- `!대회 create <대회명> <시작일> <종료일>` - 대회 생성
  - 예시: `!대회 create 알고리즘대회 2025-01-01 2025-01-21`
- `!대회 status` - 대회 상태 확인
- `!대회 blackout <on/off>` - 스코어보드 공개/비공개 설정
- `!대회 update <필드> <값>` - 대회 정보 수정
  - 필드: name, start, end
  - 예시: `!대회 update name 대회명`
- `!삭제 <백준ID>` - 참가자 삭제

## 자동 스코어보드

- **전송 시간**: 기본 오전 9시 (`SCOREBOARD_HOUR`, `SCOREBOARD_MINUTE`로 설정 가능)
- **전송 조건**: 대회 기간 내 매일 자동 전송
- **블랙아웃**: 대회 종료 3일 전부터 자동 전송 중단 (단, 마지막 날은 예외로 전송)
- **채널 설정**: `DISCORD_CHANNEL_ID` 환경변수로 지정
- **활성화 조건**: `DISCORD_CHANNEL_ID`가 설정된 경우에만 활성화

## 권한 시스템

### 관리자 권한 확인
- **서버 소유자**: 자동으로 관리자 권한 부여
- **관리자 역할**: Discord의 ADMINISTRATOR 권한을 가진 역할 보유자
- DM에선 보안상 관리자 명령어 사용 불가

### 명령어 권한 분류
- **공개 명령어**: `!등록`, `!ping`, `!도움말` - 모든 사용자 사용 가능
- **관리자 전용**: `!스코어보드`, `!참가자`, `!대회`, `!삭제` - 서버 관리자만 사용 가능

## 데이터 저장

봇은 JSON 파일을 사용하여 데이터를 저장합니다:
- `participants.json` - 참가자 정보 및 시작 시점 문제 기록
- `competition.json` - 대회 설정 및 블랙아웃 정보
- 파일 손상 시 자동으로 `.corrupted` 백업본을 생성합니다.

## API 사용

### solved.ac API
- **사용자 정보**: `https://solved.ac/api/v3/user/show?handle={백준ID}`
- **TOP 100**: `https://solved.ac/api/v3/user/top_100?handle={백준ID}`
- **추가 정보**: `https://solved.ac/api/v3/user/additional_info?handle={백준ID}` - 본명 및 개인정보 조회
- **재시도 로직**: 네트워크 오류 시 자동 재시도 (최대 3회)
- **레이트 리미팅**: API 과부하 방지를 위한 요청 제한
- **병렬 처리**: 다중 사용자 점수 계산 시 동시 요청 (최대 5개)
- **이름 검증**: 등록 시 solved.ac 프로필의 실명과 입력 이름 일치 확인

## 아키텍처

### 프로젝트 구조

```
discord-bot/
├── main.go                    # 애플리케이션 진입점
├── app/
│   └── app.go                 # 애플리케이션 생명주기 및 의존성 관리
├── interfaces/                # 의존성 역전을 위한 인터페이스 정의
│   ├── api.go                 # API 클라이언트 인터페이스
│   ├── storage.go             # 스토리지 리포지토리 인터페이스
│   └── scoring.go             # 점수 계산 인터페이스
├── config/
│   └── config.go              # 구조화된 환경 설정 관리
├── constants/
│   ├── constants.go           # 중앙화된 상수 정의
│   └── messages.go            # 사용자 메시지 및 다국어 지원 상수
├── utils/
│   ├── logger.go              # 구조화된 로깅 시스템
│   ├── validation.go          # 통합된 유효성 검사 및 날짜 처리
│   ├── error_helpers.go       # 타입별 에러 헬퍼 팩토리
│   ├── command_context.go     # 명령어 컨텍스트 헬퍼
│   └── date_formatter.go      # 날짜 포맷팅 유틸리티
├── models/
│   ├── participant.go         # 참가자 데이터 모델
│   ├── competition.go         # 대회 데이터 모델  
│   └── tier.go                # 통합된 티어 관리 시스템
├── api/
│   └── solvedac.go            # 한글 에러 메시지 및 재시도 로직
├── scoring/
│   └── calculator.go          # 인터페이스 기반 점수 계산
├── storage/
│   └── storage.go             # 인터페이스 기반 데이터 저장소
├── bot/
│   ├── commands.go            # 권한 기반 명령어 처리 및 실명 검증
│   ├── competition_handler.go # 날짜 포맷팅 개선된 대회 관리
│   └── scoreboard.go          # 스코어보드
├── errors/
│   └── errors.go              # 포괄적인 타입별 에러 시스템
├── scheduler/
│   └── scheduler.go           # 고루틴 리크 수정된 스케줄러
├── participants.json          # 참가자 데이터 (실행 시 생성)
└── competition.json           # 대회 데이터 (실행 시 생성)
```

## 라이선스

MIT License