# KRDS 네이밍 규칙 스냅샷

출처: https://www.krds.go.kr/html/site/utility/utility_03.html
수집일: 2026-07-25 (내용이 바뀐 것 같으면 사용자에게 "규칙 갱신"을 요청해 WebFetch로 재수집할 것)

## 1. 핵심 원칙 (5가지)

1. **논리 구조** — 예측 가능하고 일관된 방식으로 작성
2. **시각적 속성 배제** — "파란색 버튼" 대신 기능 중심 명칭 사용
3. **확장가능성** — "primary", "secondary" 구조로 향후 추가 대비
4. **일관성** — 단어 구분 기호 통일 (하이픈 우선)
5. **명확성** — 약어 제한. "bg" 대신 "background", "xs" 대신 "xsmall"

## 2. 작업 파일 규칙 (Figma 기준)

### 토큰에 영향을 미치는 요소

**컴포넌트(Component / Component Set)**
- 모두 소문자
- 내부 띄어쓰기는 언더바(`_`)
- 유형 구분은 더블언더바(`__`)
- 예: `button__primary`

**로컬 변수 / 로컬 스타일 (Local Variable / Style)**
- 모두 소문자
- 띄어쓰기는 언더바(`_`)
- 계층 분리는 하이픈(`-`)
- 예: `size-height`, `spacing_small`

### 토큰에 영향을 미치지 않는 요소

**페이지 / 프레임 (Page / Frame)**
- 띄어쓰기 가능, 기호 제한 없음
- 영문+한글 혼용 가능: "Main menu (메인 메뉴)"
- 영문 첫 글자만 대문자
- 유니코드 사용 가능 (✍, ✅, 🎨, 📌)
- 위반 시 severity: **low** (강제 규칙 아님, 참고용 경고만)

**프로퍼티 (Property)**
- 띄어쓰기 가능
- Name은 첫 글자 대문자
- Value는 전체 소문자

## 3. 토큰 네이밍 구조

- 구분자는 **하이픈(-)** 사용 (CSS 일관성, 점/슬래시보다 가독성 우수)
- 색상: 역할 기반 이름(`primary`, `secondary`, `gray`) 사용, 색상명(`blue`, `red`, `green` 등) 금지
  - 명도 표현은 10 단위 증가(`primary-10`, `primary-20`, ...), 세밀도 필요 시 5 단위 포함 가능
- 타이포그래피: `font-family`, `font-size`, `line-height` 등 유형별 구분, 단위 통일(px 또는 rem)
- 숫자(간격 등): 4px, 8px 배수가 기본, 중간 단위로 2px·10px 추가 가능, 시맨틱 표현(`small`/`medium`/`large`) 또는 역할 없는 경우 단계(`1`, `2`) 사용

## 4. 토큰 8단계 속성 구조

```
namespace > theme > category > component > type > variant > element > state
```

| 단계 | 예시 | 설명 |
|---|---|---|
| Namespace | `krds` | 코드 구분 식별자 |
| Theme | `light`, `dark` | 모드/용도 구분 |
| Category | `color`, `spacing` | 큰 범주 |
| Component | `button`, `card` | UI 요소 |
| Type | `background`, `surface` | 역할/기능 |
| Variant | `primary`, `danger` | 계층/시스템 구조 |
| Element | `label`, `body`, `text` | 하위 구성요소 |
| State | `default`, `hover`, `pressed` | 상호작용 상태 |

## 5. 토큰 속성 세부 값

**모드(Mode)**: `-light`(기본), `-high-contrast`(선명한 화면 모드)

**용도별 속성**: `-background`, `-surface`, `-divider`, `-element`, `-action`, `-gap`, `-padding`, `-fill`, `-line`, `-alpha`, `-neutral`, `-on(배경)`, `-dim`, `-inverse`, `-static`

**밝기**: `-lighter`, `-light`, `-normal`, `-dark`, `-darker`

**강조**: `-subtler`, `-subtle`, `-bold`, `-bolder`

**상태(허용 목록)**: `-default`, `-hover`, `-pressed`, `-focused`, `-disabled`, `-error`, `-active`, `-completed`, `-selected`, `-unselected`, `-indeterminate`
- **금지**: `-view` (사용 금지로 명시됨)

**크기(허용 목록)**: `xxsmall`, `xsmall`, `small`, `medium`, `large`, `xlarge`, `xxlarge`
- **금지**: `xs`, `sm`, `md`, `lg`, `xl` 등 약어형

## 6. 결정론적 검사로 옮기기 어려운 항목

- "논리 구조", "확장가능성" 같은 원칙은 문자열 패턴만으로 완벽히 판정 불가 → 린트 스크립트는 아래 항목만 규칙화한다:
  1. 컴포넌트명 소문자 + `_`/`__` 패턴
  2. 로컬 변수/스타일명 소문자 + `_`/`-` 패턴
  3. 페이지/프레임명 첫 글자 대문자 (low severity)
  4. 토큰 하이픈 구분자 사용 여부
  5. 색상 토큰의 색상명(blue/red/green/yellow/purple/orange/pink/black/white) 사용 금지
  6. 금지 약어(`bg`, `xs`, `sm`, `md`, `lg`, `xl`) 검출
  7. 상태 접미사 허용 목록 검사 + `-view` 금지
  8. 사이즈 키워드 허용 목록 검사

## 7. 확장(참고용) 값 검사 — 원문이 강제하지 않는 권장 사항

- spacing 계열 변수의 실제 값이 4 또는 8의 배수가 아니면 참고 항목으로 표시 (2, 10은 예외 허용)
- color 계열 변수의 이름에 포함된 명도 단계 숫자가 10 단위 규칙에서 벗어나면 참고 항목으로 표시
- 이 섹션은 항상 "권장 사항(원문 강제 아님)" 라벨을 붙여 본 규칙 위반과 분리 표시한다
