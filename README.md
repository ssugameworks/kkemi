<h1 align="center">깨미</h1>

<p align="center">백준 알고리즘 문제 풀이 점수를 집계하여 스코어보드를 제공하는 디스코드 봇입니다.

---
[![Go Version](https://img.shields.io/badge/Go-1.25.0-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Discord.js](https://img.shields.io/badge/DiscordGo-v0.29.0-7289da?style=flat&logo=discord)](https://github.com/bwmarrin/discordgo)
[![Firebase](https://img.shields.io/badge/Firebase-v3.13.0-ffca28?style=flat&logo=firebase)](https://firebase.google.com/)
[![Google Cloud](https://img.shields.io/badge/Google%20Cloud-Monitoring-4285f4?style=flat&logo=googlecloud)](https://cloud.google.com/monitoring)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Test Coverage](https://img.shields.io/badge/Coverage-85%25-brightgreen.svg)](https://github.com/your-username/Discord-Bot/actions)


## ✨ 주요 기능

- **참가자 관리**: 백준 사용자 자동 등록 및 티어 확인
- **실시간 집계**: solved.ac API를 활용한 병렬 점수 계산  
- **블랙아웃 모드**: 대회 종료 전 N일간 스코어보드 비공개
- **차등 점수**: 도전/기본/연습 문제에 따른 차등 점수
- ️**대회 관리**: 대회 생성, 수정, 상태 관리
- **자동화**: 설정한 시간에 스코어보드 송신
- **다중 채널**: DM 및 서버 채널 지원
- **텔레메트리**: 실시간 메트릭 수집

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

## 🚀 배포 및 실행

### 1. 환경 설정
`.env` 파일을 생성하고 다음 내용을 추가하세요:

```bash
# Discord Bot Configuration (필수)
export DISCORD_BOT_TOKEN="your_discord_bot_token_here"
export DISCORD_CHANNEL_ID="your_channel_id_here"

# Firebase/Firestore Configuration (프로덕션 환경용)
export FIREBASE_CREDENTIALS_JSON='{"type":"service_account",...}'

# Competition Initialization (선택사항 - 자동 대회 생성)
export COMPETITION_NAME="Test Competition"
export COMPETITION_START_DATE="2025-01-01"
export COMPETITION_END_DATE="2025-01-31"

# Scoreboard Schedule Configuration (선택사항)
export SCOREBOARD_HOUR="9"      # 스코어보드 전송 시간 (0-23)
export SCOREBOARD_MINUTE="0"    # 스코어보드 전송 분 (0-59)

# Telemetry Configuration (선택사항 - 메트릭 수집)
export TELEMETRY_ENABLED="true"           # 텔레메트리 활성화
export GOOGLE_CLOUD_PROJECT="your-project-id"  # Google Cloud 프로젝트 ID

# 기타 설정 (선택사항)
export LOG_LEVEL="INFO"         # 로그 레벨 (DEBUG, INFO, WARN, ERROR)
export DEBUG_MODE="false"       # 디버그 모드
export JSON_LOGGING="false"     # JSON 형식 로깅 (프로덕션용)
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

## Discord Developer Portal 설정

### Bot 권한 설정
다음 권한이 필요합니다:
- `Send Messages` - 메시지 전송
- `Read Message History` - 메시지 기록 읽기
- `View Channels` - 채널 보기

### 인텐트 설정
다음 인텐트들이 활성화되어야 합니다:
- `Message Content Intent` - 메시지 내용 읽기
- `Server Members Intent` - 서버 멤버 정보

## 사용법

### 참가자 명령어
- `!등록 <이름> <백준ID>` - 대회 등록 신청 (대회 시작 후, solved.ac 등록 이름과 일치해야 함)
- `!도움말` - 도움말 표시
- `!ping` - 봇 응답 확인

### 관리자 전용 명령어
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

## 데이터 저장소

### 프로덕션 환경 (Firebase/Firestore)
- **참가자 데이터**: Firestore의 `participants` 컬렉션에 저장
- **대회 정보**: Firestore의 `competitions` 컬렉션에 저장
- **자동 재연결**: 네트워크 장애 시 자동 복구
- **헬스체크**: 연결 상태 실시간 모니터링

### 개발/테스트 환경 (In-Memory)
- **임시 저장소**: 메모리 기반 비영구 데이터 저장
- **빠른 개발**: Firebase 설정 없이 즉시 테스트 가능
- **자동 폴백**: Firebase 인증 정보가 없으면 자동 전환

## API 사용

### solved.ac API
- **사용자 정보**: `https://solved.ac/api/v3/user/show?handle={백준ID}`
- **TOP 100**: `https://solved.ac/api/v3/user/top_100?handle={백준ID}`
- **추가 정보**: `https://solved.ac/api/v3/user/additional_info?handle={백준ID}` - 본명 및 개인정보 조회
- **조직 정보**: `https://solved.ac/api/v3/user/organizations?handle={백준ID}` - 대학 소속 확인

### 성능 최적화
- **적응형 동시성**: API 응답 시간에 따른 자동 동시성 제어 (1-10개)
- **메모리 풀**: 객체 재사용으로 GC 압박 감소
- **효율적 캐시**: 15분 TTL 기반 자동 캐시 관리
- **재시도 로직**: 네트워크 오류 시 자동 재시도 (최대 3회)
- **레이트 리미팅**: API 과부하 방지를 위한 요청 제한

## 텔레메트리 및 모니터링

### 메트릭 수집
- **명령어 사용량**: 각 명령어별 호출 횟수와 관리자/일반 사용자 분류
- **캐시 성능**: API 캐시 히트율, 총 호출 수, 캐시된 데이터 수
- **대회 활동**: 대회 생성, 참가자 등록 등의 이벤트 추적
- **성능 메트릭**: 스코어보드 생성 시간, API 응답 시간 등

### Google Cloud Monitoring 연동
- **커스텀 메트릭**: `custom.googleapis.com/discord_bot/*` 네임스페이스 사용
- **실시간 대시보드**: 봇 사용량과 성능 지표 실시간 모니터링
- **알람 설정**: 임계값 초과 시 자동 알림 (선택사항)
- **로그 통합**: Cloud Logging과 연동된 구조화된 로그

### 설정 방법
1. Google Cloud Console에서 Monitoring API 활성화
2. 서비스 계정에 Monitoring Metric Writer 권한 부여
3. `TELEMETRY_ENABLED=true` 및 `GOOGLE_CLOUD_PROJECT` 환경변수 설정
4. Firebase 인증 정보로 자동 인증 (동일 프로젝트 권장)

## 아키텍처

### 프로젝트 구조

```
discord-bot/
├── main.go                    # 애플리케이션 진입점
├── app/
│   └── app.go                 # 애플리케이션 생명주기 및 의존성 관리
├── interfaces/
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
│   ├── tier.go                # 통합된 티어 관리 시스템
│   └── score_data.go          # 점수 데이터 모델
├── api/
│   └── solvedac.go            # 캐시된 solved.ac API 클라이언트
├── cache/
│   └── api_cache.go           # TTL 기반 API 캐시 시스템
├── performance/
│   ├── memory_pool.go         # 메모리 풀 및 채널 관리
│   └── adaptive_concurrency.go # 적응형 동시성 관리
├── health/
│   └── health_checker.go      # Firebase/Firestore 헬스체크
├── scoring/
│   └── calculator.go          # 인터페이스 기반 점수 계산
├── storage/
│   └── storage.go             # Firebase/In-Memory 하이브리드 저장소
├── telemetry/
│   └── metrics.go             # Google Cloud Monitoring 메트릭 전송
├── bot/
│   ├── commands.go            # 권한 기반 명령어 처리 및 실명 검증
│   ├── command_deps.go        # 명령어 의존성 관리 및 봇 상태 업데이트
│   ├── competition_handler.go # 날짜 포맷팅 개선된 대회 관리
│   └── scoreboard.go          # 병렬 처리 기반 스코어보드
├── errors/
│   └── errors.go              # 포괄적인 타입별 에러 시스템
└── scheduler/
    └── scheduler.go           # 고루틴 리크 수정된 스케줄러
```

## 🚀 최신 개선사항 (v2.0)

### 성능 최적화
- **적응형 동시성 관리**: API 응답 시간에 따른 자동 동시성 조절
- **메모리 풀링**: 객체 재사용으로 GC 압박 최소화
- **효율적 캐시**: 자동 TTL 관리와 메모리 최적화
- **데드락 방지**: 채널 반환 시 안전한 메모리 풀 관리

### 사용자 경험
- **동적 봇 상태**: 현재 활성 대회 이름이 봇 상태로 자동 표시
- **실시간 업데이트**: 대회 생성/수정 시 즉시 봇 상태 반영
- **향상된 로깅**: 구조화된 JSON 로깅 시스템
- **포괄적 에러 처리**: 타입별 에러 분류와 사용자 친화적 메시지
- **실시간 모니터링**: Google Cloud Monitoring을 통한 성능 및 사용량 추적

### 코드 품질
- **의존성 주입**: 테스트 가능한 깔끔한 아키텍처
- **인터페이스 분리**: 모듈 간 낮은 결합도
- **타입 안전성**: 포괄적인 타입 검사
- **단위 테스트**: 핵심 로직 테스트 커버리지

## 라이선스

MIT License
