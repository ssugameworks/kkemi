# 🐜 깨미 (Kkemi)

**Figma의 컨벤션 확인을 자동화합니다.**

Product Team이 만든 Figma 린터입니다. Figma 파일의 네이밍 규칙과 명도 대비를 [KRDS](https://www.krds.go.kr) 기준에 맞춰 자동으로 점검하고, 코딩 에이전트 안에서 바로 리포트를 받아볼 수 있습니다.

[![python](https://img.shields.io/badge/python-3.x-blue)](scripts/kkemi_check.py)
[![license](https://img.shields.io/badge/license-Internal-lightgrey)](#)
[![agent](https://img.shields.io/badge/works%20with-Claude%20Code%20%7C%20Codex%20%7C%20Cursor-8a2be2)](#호환-에이전트)

---

디자인 리뷰에서 네이밍 컨벤션과 명도 대비를 사람이 눈으로 확인하는 건 느리고, 리뷰어마다 판정이 달라집니다.  
깨미는 이 중 **결정론적으로 판정 가능한 항목만** 정규식과 WCAG 공식으로 정확히 계산하고, 코드 구현 단계에서만 확인 가능한 항목(키보드 탐색, 포커스 트랩 등)은 검증했다고 우기지 않고 검사 범위에서 아예 제외합니다.

- ✅ 애매하게 "괜찮아 보이는데요"가 아니라 표와 수치로 답합니다.
- ✅ 불확실한 전경/배경 쌍은 억지로 계산하지 않고 "수동 확인 필요"로 분리합니다.
- ✅ 어떤 코딩 에이전트에서 실행해도 같은 절차, 같은 결과를 냅니다.

## 검사 규칙

| 영역 | 방식 | 판정 |
|---|---|---|
| 컴포넌트/변수 네이밍 | 정규식 (`kkemi_check.py`) | 확정 (severity: high/medium/low) |
| 페이지/프레임 네이밍 | 정규식, 참고용 | low (강제 아님) |
| 간격·색상 토큰 확장 규칙 | 4/8px 배수, 10단위 색상 스텝 | 참고용 경고 |
| 명도 대비 (텍스트) | WCAG relative luminance 공식 | 확정, 기준 4.5:1 |
| 명도 대비 (아이콘·비텍스트) | WCAG relative luminance 공식 | 확정, 기준 3:1 |

키보드 탐색, 포커스 트랩, alt 텍스트 등 코드 구현 단계에서만 확인 가능한 항목은 정적 디자인으로 검증할 수 없어 **검사 범위에서 제외**합니다 (리포트에 출력하지 않음).

자세한 규칙 원문 스냅샷은 [`references/naming-rules.md`](references/naming-rules.md), [`references/accessibility-rules.md`](references/accessibility-rules.md)에 있습니다.

## Quick start

### 사전 준비사항

1. Figma 데스크톱 앱이 켜져 있고 로컬 MCP 서버(기본 `http://127.0.0.1:3845/mcp`)가 연결되어 있어야 합니다.
2. 코딩 에이전트가 해당 MCP의 `get_metadata`, `get_variable_defs`, `get_design_context`(선택: `get_screenshot`) 툴을 사용할 수 있어야 합니다.
3. Python 3가 설치되어 있어야 합니다.

### 사용법

에이전트 채팅에서 검사하고 싶은 Figma 프레임 URL과 함께 요청합니다.

```
깨미야 이 프레임 검사해줘
https://figma.com/design/:fileKey/:fileName?node-id=1-2
```

에이전트가 `AGENTS.md`에 정의된 절차를 그대로 수행합니다. 스크립트만 직접 돌려보고 싶다면:

```bash
python3 scripts/kkemi_check.py \
  --nodes nodes.json \
  --variables variables.json \
  --contrast-pairs contrast_pairs.json
```

Figma 데이터 없이 대비 계산 로직만 검증하려면:

```bash
python3 scripts/kkemi_check.py --selftest
```

## 동작 방식

```
Figma 파일
   │
   ├─ get_metadata          → 노드 트리(XML)         → scripts/parse_metadata.py       → nodes.json
   ├─ get_variable_defs     → 로컬 변수/스타일 토큰   →                                 → variables.json
   └─ get_design_context    → 생성된 React/Tailwind → scripts/parse_design_context.py → contrast_pairs.json
              │
              ↓
     scripts/kkemi_check.py   (정규식 및 WCAG 대비 계산)
              ↓
   naming_violations / value_extensions / contrast_results (JSON 정규화)
              ↓
          보고서 생성
```

규칙과 검사 로직의 단일 소스는 `references/`와 `scripts/`이며, 다른 곳에 복제하지 않습니다.  
에이전트별 설정(`.claude/skills/kkemi/SKILL.md` 등)은 툴 이름 매핑만 보충하고, 실제 절차는 항상 [`AGENTS.md`](AGENTS.md)를 가리킵니다.

## 리포트 예시

```markdown
## 🐜 깨미 린트 리포트 — Main menu (메인 메뉴)

### 요약
- 검사 노드 42개, 변수 18개 / 위반 3건 (high 1, medium 1, low 1)

### 1. 네이밍 규칙 위반 (KRDS 네이밍 원칙)
| 이름 | node id | 위반 규칙 | 문제 | 수정 제안 |
|---|--------------|---|---|---|
| `button_blue` | 12:34        | 시각적 속성 배제 | 색상명이 이름에 노출됨 | `button__primary` |

### 3. 접근성 자동검증 — 명도 대비
| 요소 | 전경색 | 배경색 | 계산된 대비율 | 기준 | 결과 |
|---|---|---|---|---|---|
| 카드 본문 텍스트 | #767676 | #FFFFFF | 4.48:1 | 4.5:1 | **Fail** |
```

## 프로젝트 구조

```
.
├── AGENTS.md                        # 에이전트 공용 작업 지침 (단일 소스)
├── references/
│   ├── naming-rules.md              # KRDS 네이밍 규칙 스냅샷
│   └── accessibility-rules.md       # KRDS 접근성 규칙 스냅샷
├── scripts/
│   ├── kkemi_check.py               # 정규식 검사 + WCAG 대비 계산 (Python 3 표준 라이브러리만)
│   ├── parse_metadata.py            # get_metadata XML → nodes.json 변환 헬퍼
│   └── parse_design_context.py      # get_design_context(React/Tailwind) → contrast_pairs.json 변환 헬퍼
└── .claude/skills/kkemi/SKILL.md    # Claude Code 전용 툴 이름 매핑
```

## 호환 에이전트

`AGENTS.md`는 특정 클라이언트에 종속되지 않도록 작성되어 있어, Figma MCP를 지원하는 어떤 코딩 에이전트에서도 동일하게 동작합니다.

- Claude Code
- Codex CLI
- Cursor
- Windsurf

## 규칙 갱신

규칙 원문은 수집일 기준 스냅샷으로 내장되어 있습니다. KRDS 원문이 바뀐 것 같다면 에이전트에게 **"규칙 갱신해줘"**라고 명시적으로 요청하세요 — 그 외에는 자동으로 재조회하지 않습니다.

## 설계 원칙

- **판정 가능한 것만 판정한다.** 정규식·WCAG 공식으로 계산되지 않는 항목은 "검사했다"고 말하지 않는다.
- **단일 소스 유지.** 규칙과 로직은 `references/`, `scripts/`에만 있고 에이전트별 설정 파일은 이를 가리키기만 한다.
- **애매하면 건너뛴다.** 전경/배경 페어를 확신 있게 특정할 수 없으면 억지로 계산하지 않고 수동 확인 항목으로 넘긴다.
