# 깨미 🤖

<p align="center">
<strong>백준 알고리즘 문제 풀이 점수를 집계하여 스코어보드를 제공하는 Discord 봇</strong>
</p>

<p align="center">
<a href="#주요-기능">기능</a> •
<a href="#빠른-시작">빠른 시작</a> •
<a href="#사용법">사용법</a> •
<a href="#점수-계산">점수 계산</a> •
<a href="#문서">문서</a> •
<a href="#기여">기여</a>
</p>

---

[![Go Version](https://img.shields.io/badge/Go-1.25.0-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Code Quality](https://img.shields.io/badge/Code%20Quality-A+-brightgreen?style=flat&logo=go&logoColor=white)](https://github.com/ssugameworks/kkemi)
[![DiscordGo](https://img.shields.io/badge/DiscordGo-v0.29.0-7289da?style=flat&logo=discord)](https://github.com/bwmarrin/discordgo)
[![Firebase](https://img.shields.io/badge/Firebase-Firestore-ffca28?style=flat&logo=firebase)](https://firebase.google.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

## 주요 기능

### 🎯 핵심 기능
- **리그 시스템**: 실력별 리그 분류 (루키/프로/마스터)
- **차등 점수**: 도전/기본/연습 문제에 따른 가중치 적용
- **실시간 집계**: solved.ac API를 활용한 병렬 점수 계산
- **블랙아웃 모드**: 대회 종료 전 N일간 스코어보드 비공개

### ⚙️ 관리 기능
- **대회 관리**: 대회 생성, 수정, 상태 관리
- **참가자 관리**: 자동 등록 및 실명 검증
- **자동화**: 설정 시간에 자동 스코어보드 전송
- **다중 채널**: DM 및 서버 채널 지원

### 📊 성능 & 모니터링
- **적응형 동시성**: API 응답 시간 기반 자동 조절
- **메모리 풀링**: 객체 재사용으로 GC 압박 최소화
- **캐싱**: 15분 TTL 기반 자동 캐시 관리
- **텔레메트리**: Google Cloud Monitoring 연동

---

## 빠른 시작

### 필수 요구사항

- Go 1.25.0 이상
- Discord Bot Token
- Firebase 프로젝트 (프로덕션 환경)

### 설치 및 실행

```bash
# 1. 저장소 클론
git clone https://github.com/ssugameworks/kkemi.git
cd kkemi

# 2. 의존성 설치
go mod tidy

# 3. 환경변수 설정
cp .env.example .env
# .env 파일을 편집하여 필요한 값 입력

# 4. 봇 실행
source .env
go run main.go
```

### 환경변수 설정

`.env` 파일에 다음 환경변수를 설정하세요:

#### 필수 설정
```bash
# Discord Bot 설정
export DISCORD_BOT_TOKEN="your_discord_bot_token_here"
export DISCORD_CHANNEL_ID="your_channel_id_here"
```

#### 데이터베이스 (프로덕션)
```bash
# Firebase/Firestore 설정
export FIREBASE_CREDENTIALS_JSON='{"type":"service_account",...}'
```

#### 대회 자동 생성 (선택)
```bash
export COMPETITION_NAME="잔디심기 챌린지 2025"
export COMPETITION_START_DATE="2025-01-01"
export COMPETITION_END_DATE="2025-01-31"
```

#### 스케줄링 (선택)
```bash
export SCOREBOARD_HOUR="9"      # 스코어보드 전송 시간 (0-23)
export SCOREBOARD_MINUTE="0"    # 스코어보드 전송 분 (0-59)
```

#### Google Sheets 연동 (선택)
```bash
export PARTICIPANT_SPREADSHEET_ID="your_spreadsheet_id"  # 참가자 명단 시트
export SCOREBOARD_SPREADSHEET_ID="your_spreadsheet_id"   # 스코어보드 시트
```

#### 텔레메트리 (선택)
```bash
export TELEMETRY_ENABLED="true"
export GOOGLE_CLOUD_PROJECT="your-project-id"
```

#### 기타 설정 (선택)
```bash
export LOG_LEVEL="INFO"         # DEBUG, INFO, WARN, ERROR
export DEBUG_MODE="false"
export JSON_LOGGING="false"     # 프로덕션용 JSON 로깅
```

---

## Discord Bot 설정

### Bot 권한
다음 권한이 필요합니다:
- ✅ Send Messages
- ✅ Read Message History
- ✅ View Channels
- ✅ Use Slash Commands (향후)

### 인텐트
다음 인텐트를 활성화하세요:
- ✅ Message Content Intent
- ✅ Server Members Intent

### 초대 링크
```
https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=274877909056&scope=bot
```

---

## 사용법

### 일반 사용자 명령어

#### `!등록 <이름> <백준ID>`
대회에 등록합니다.

```
!등록 홍길동 baekjoon123
```

**조건**:
- 대회가 진행 중이어야 함
- solved.ac에 등록된 이름과 일치해야 함
- 숭실대학교 소속이어야 함 (organization_id: 323)

#### `!ping`
봇 응답 확인

#### `!도움말`
도움말 표시

---

### 관리자 전용 명령어

#### 대회 관리

```bash
# 대회 생성
!대회 create <대회명> <시작일> <종료일>
예시: !대회 create "알고리즘 챌린지 2025" 2025-01-01 2025-01-31

# 대회 상태 확인
!대회 status

# 대회 정보 수정
!대회 update name <새이름>
!대회 update start <새시작일>
!대회 update end <새종료일>

# 블랙아웃 모드
!대회 blackout on   # 스코어보드 비공개
!대회 blackout off  # 스코어보드 공개
```

#### 참가자 관리

```bash
# 참가자 목록 확인
!참가자

# 참가자 삭제
!삭제 <백준ID>
예시: !삭제 baekjoon123
```

#### 스코어보드

```bash
# 현재 스코어보드 확인
!스코어보드
```

---

## 점수 계산

### 리그 분류

참가자는 **등록 시점의 백준 티어**에 따라 3개 리그로 분류됩니다.

| 리그 | 티어 범위 | 설명 |
|------|----------|------|
| 🌱 루키 | Unrated ~ Silver V (0-6) | 알고리즘 입문자 |
| ⚡ 프로 | Silver IV ~ Gold V (7-11) | 중급 실력자 |
| 👑 마스터 | Gold IV 이상 (12+) | 고급 실력자 |

### 가중치 시스템

문제 난이도와 참가자 티어를 비교하여 가중치를 적용합니다.

#### 루키 리그
- 도전 문제 (티어 > 등록 티어): **×1.4**
- 기본 문제 (티어 = 등록 티어): **×1.0**
- 연습 문제 (티어 < 등록 티어): **×0.5**

#### 프로 리그
- 도전 문제: **×1.2**
- 기본 문제: **×1.0**
- 연습 문제: **×0.8**

#### 마스터 리그
- 모든 문제: **×1.0**

### 점수 공식

```
최종 점수 = Σ(문제 난이도 점수 × 가중치)
```

**상세 정보**: [점수 산정 시스템 문서](./SCORING.md)

---

## 자동 스코어보드

### 전송 시간
- 기본: 매일 오전 9시
- 설정: `SCOREBOARD_HOUR`, `SCOREBOARD_MINUTE`

### 전송 조건
- 대회 기간 내에만 전송
- 블랙아웃 기간에는 전송 중단
  - 대회 종료 3일 전부터 비공개
  - 마지막 날은 예외로 전송

### 채널 설정
`DISCORD_CHANNEL_ID` 환경변수로 지정된 채널로 전송

---

## 데이터 저장소

### 프로덕션 환경 (Firestore)
- **참가자 데이터**: `competitions/{competitionId}/participants/{baekjoonId}`
- **대회 정보**: `competitions/{competitionId}`
- **자동 재연결**: 네트워크 장애 시 자동 복구
- **헬스체크**: 연결 상태 실시간 모니터링

### 개발/테스트 환경 (In-Memory)
- Firebase 설정 없이 즉시 테스트 가능
- 비영구 메모리 기반 저장
- `FIREBASE_CREDENTIALS_JSON` 미설정 시 자동 전환

---

## 성능 최적화

### API 최적화
- **적응형 동시성**: 1~20개 동적 조절
- **캐싱**: 15분 TTL 자동 캐시
- **재시도**: 최대 3회 exponential backoff
- **병렬 처리**: Goroutine 워커 풀

### 메모리 최적화
- **메모리 풀**: sync.Pool로 객체 재사용
- **사전 할당**: slice capacity 미리 지정
- **효율적 GC**: 불필요한 할당 최소화

### 동시성 제어
- **세마포어**: 동시 요청 수 제한
- **RWMutex**: 읽기/쓰기 잠금 분리
- **WaitGroup**: Goroutine 동기화

**상세 정보**: [아키텍처 문서](./ARCHITECTURE.md)

---

## 텔레메트리

### 수집 메트릭
- **명령어 사용량**: 명령어별 호출 횟수
- **캐시 성능**: 히트율, 총 호출 수
- **대회 활동**: 참가자 등록, 대회 생성
- **성능**: 스코어보드 생성 시간

### Google Cloud Monitoring
- 커스텀 메트릭: `custom.googleapis.com/discord_bot/*`
- 실시간 대시보드
- 알람 설정 가능

### 설정 방법
1. Google Cloud Console에서 Monitoring API 활성화
2. 서비스 계정에 **Monitoring Metric Writer** 권한 부여
3. 환경변수 설정:
   ```bash
   export TELEMETRY_ENABLED="true"
   export GOOGLE_CLOUD_PROJECT="your-project-id"
   ```

---

## 문서

### 주요 문서
- 📖 [아키텍처 문서](./ARCHITECTURE.md) - 시스템 설계 및 구조
- 📊 [점수 산정 시스템](./SCORING.md) - 점수 계산 알고리즘
- 📝 [API 명세](./document.yaml) - OpenAPI 스펙

### 기술 문서
- [DiscordGo Documentation](https://pkg.go.dev/github.com/bwmarrin/discordgo)
- [Firebase Go SDK](https://firebase.google.com/docs/admin/setup)
- [solved.ac API](https://solvedac.github.io/unofficial-documentation/)

---

## 프로젝트 구조

```
kkemi/
├── main.go                    # 진입점
├── app/                       # 애플리케이션 생명주기
├── bot/                       # Discord 명령어 처리
│   ├── commands.go
│   ├── competition_handler.go
│   └── scoreboard.go
├── api/                       # solved.ac API 클라이언트
├── cache/                     # TTL 캐시 시스템
├── scoring/                   # 점수 계산 로직
├── storage/                   # Firestore/InMemory 저장소
├── performance/               # 성능 최적화
│   ├── memory_pool.go
│   └── adaptive_concurrency.go
├── telemetry/                 # Google Cloud Monitoring
├── scheduler/                 # 자동 스케줄링
├── models/                    # 데이터 모델
├── interfaces/                # 인터페이스 정의
├── constants/                 # 상수 정의
├── utils/                     # 유틸리티
└── errors/                    # 에러 처리
```

---

## 개발

### 테스트 실행
```bash
# 전체 테스트
go test ./...

# 특정 패키지
go test ./scoring -v

# 커버리지
go test -cover ./...
```

### 로컬 개발
```bash
# 개발 모드 (In-Memory Storage)
export DISCORD_BOT_TOKEN="your-token"
export DISCORD_CHANNEL_ID="your-channel"
go run main.go
```

### 코드 포맷팅
```bash
# gofmt 실행
gofmt -w .

# 린트
golangci-lint run
```

---

## 배포

### Railway 배포
1. Railway 프로젝트 생성
2. GitHub 저장소 연결
3. 환경변수 설정
4. 자동 배포

### 헬스체크
```bash
# 헬스 엔드포인트
curl http://localhost:8080/health

# 응답 예시
{
  "status": "healthy",
  "timestamp": "2025-01-12T10:00:00Z",
  "checks": {
    "firestore": true
  }
}
```

---

## 기여

### 기여 방법
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### 코딩 컨벤션
- Go 표준 컨벤션 준수
- gofmt로 포맷팅
- 의미있는 커밋 메시지
- 테스트 커버리지 유지

---

## 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

---

## 문의

- **GitHub Issues**: [여기서 이슈 생성](https://github.com/ssugameworks/kkemi/issues)
- **Discord DM**: stdlib_h

---