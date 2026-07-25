# 깨미 (Kkemi) — Figma Linter

이 저장소는 게임웍스 Product Team이 Figma 디자인 파일이 KRDS(대한민국 정부 디자인시스템, https://www.krds.go.kr) 규칙을 지키는지 점검하기 위한 린터 "깨미"의 규칙 스냅샷과 검사 스크립트를 담고 있다.

이 문서는 Claude Code, Codex CLI, Cursor, Windsurf 등 **어떤 코딩 에이전트에서 작업하든 동일하게 적용되는 지침**이다. 에이전트별 전용 기능(예: Claude Code의 Skill 자동 트리거)은 각 도구의 설정 파일(예: `.claude/skills/kkemi/SKILL.md`)에 별도로 있으며, 그 파일들도 결국 이 문서와 `kkemi/` 폴더를 가리킨다. **규칙과 검사 로직의 단일 소스는 `kkemi/` 폴더이며, 다른 곳에 복제하지 않는다.**

## 전제 조건: Figma MCP 연결

이 작업을 수행하려면 에이전트가 Figma MCP 서버(Figma 데스크톱 앱의 로컬 MCP, 기본적으로 `http://127.0.0.1:3845/mcp`)에 연결되어 있어야 한다. 노출되는 툴 이름은 클라이언트마다 접두사가 다를 수 있다 (예: Claude Code에서는 `mcp__claude_ai_Figma__get_metadata`, 다른 클라이언트에서는 접두사 없이 `get_metadata`로 노출될 수 있음). 아래 지침에서는 접두사 없이 **기능 이름만** 표기하니, 실제 호출 시에는 연결된 MCP 클라이언트가 노출하는 정확한 툴 이름으로 바꿔 호출한다. 필요한 기능:

- `get_metadata(fileKey, nodeId)` — 노드 트리(이름/타입/구조) 조회
- `get_variable_defs(fileKey, nodeId)` — 바인딩된 로컬 변수/토큰 이름과 값 조회
- `get_design_context(fileKey, nodeId)` — 코드/스타일 컨텍스트 조회 (대비 계산용 전경/배경 색상 페어링)
- `get_screenshot(fileKey, nodeId)` — (선택) 전경/배경 관계를 육안으로 확인할 때

## 언제 이 워크플로우를 실행하는가

사용자가 "깨미", "KRDS 검사", "피그마 린트", "네이밍 규칙 검사", "명도 대비 검사", "접근성 검사" 등을 **Figma URL과 함께** 요청할 때 아래 절차를 수행한다.

**이 스킬은 Figma URL 필수다.** URL이 없으면 다른 단계를 진행하지 말고 "검사할 Figma 파일/프레임의 URL을 알려주세요"라고 요청한다. (예: `https://figma.com/design/:fileKey/:fileName?node-id=1-2`)

## 실행 절차

### 0. URL 파싱
- `fileKey`: `/design/` 다음 경로 세그먼트 (URL이 `/branch/:branchKey/` 형식이면 `branchKey`를 `fileKey`로 사용)
- `nodeId`: `node-id` 쿼리 파라미터, `-`를 `:`로 변환 (예: `1-2` → `1:2`)

### 1. 노드 트리 조회
`get_metadata(fileKey, nodeId)` 호출 → 응답(XML)을 파싱해 `{id, name, type, is_top_level}` 목록으로 정리한다. nodeId로 지정된 루트 바로 아래 자식은 `is_top_level: true`로 표시한다(프레임/페이지 네이밍 규칙은 최상위 노드에만 적용). 임시 작업 디렉터리에 `nodes.json`으로 저장한다.

**이 변환은 직접 파싱 코드를 새로 짜지 말고 `scripts/parse_metadata.py`를 사용한다.** 특히 프레임이 커서 `get_metadata` 응답이 토큰 한도를 넘어 파일로 저장되는 경우, 그 파일 경로를 그대로 입력으로 넘기면 된다 (도구 결과 JSON 배열 형태와 순수 XML 문자열 둘 다 지원):
```bash
python3 scripts/parse_metadata.py --input <get_metadata 응답 또는 저장된 tool-result 파일 경로> --output <임시경로>/nodes.json
```
`--selftest`로 Figma 데이터 없이 파싱 로직 자체를 검증할 수 있다. 이 스크립트가 다루지 못하는 새로운 태그를 만나면 스크립트에 매핑을 추가하고, 매번 즉석 파이썬으로 재구현하지 않는다(단일 소스 유지).

### 2. 변수/토큰 조회
`get_variable_defs(fileKey, nodeId)` 호출 → `{변수명: 값}` 형태로 `variables.json`에 저장한다. 색상 값은 hex 문자열, 간격/크기 값은 숫자로 저장한다.

### 3. 대비 계산용 전경/배경 페어 수집
`get_design_context(fileKey, nodeId)` 호출 → 텍스트 노드의 전경색과, 그 텍스트를 감싸는 가장 가까운 컨테이너의 배경색(fill)을 짝지어 `contrast_pairs.json`으로 저장한다. **이 응답은 XML이 아니라 React+Tailwind로 생성된 코드(`data-node-id`, `className="text-[color:var(--...,#hex)]"` 등)이므로, 직접 정규식/파싱 코드나 색상 합성(alpha blending) 코드를 새로 짜지 말고 `scripts/parse_design_context.py`를 사용한다:**
```bash
python3 scripts/parse_design_context.py \
  --input <get_design_context 응답 또는 저장된 tool-result 파일 경로> \
  --nodes <임시경로>/nodes.json \
  --output <임시경로>/contrast_pairs.json
```
- `--nodes`에 1단계에서 만든 `nodes.json`을 넘기면 TEXT 타입 노드는 `kind: "text"`(기준 4.5:1), VECTOR/BOOLEAN_OPERATION/LINE 타입은 `kind: "nontext"`(기준 3:1)로 자동 분류된다. `get_metadata`가 컴포넌트 인스턴스 내부를 펼치지 않아 `nodes.json`에 없는 id는, 자기 자신에게 텍스트 색상이 있으면 `kind: "text"`로 추정하는 폴백이 적용된다.
- 반투명 배경(`rgba(...)`, 여러 겹 오버레이)은 스크립트가 알파 합성까지 자동으로 계산해 최종 hex를 뽑아준다. 진짜 그라디언트(구간마다 색이 다른 경우)나 이미지 배경처럼 하나의 색으로 환원할 수 없는 경우에만 페어 생성을 건너뛰고 `unresolved` 목록에 사유(예: "그라디언트 배경(~#0c0c0d ~ #223063)")를 붙여 반환한다. **이 판단은 스크립트가 하는 것이고, 에이전트가 즉석 코드로 배경색을 손수 계산하거나 반대로 애매한 경우를 억지로 확정 짓지 않는다.**
- `--selftest`로 Figma 데이터 없이 파싱/합성 로직 자체를 검증할 수 있다.
- 실행 후 표준출력에 찍히는 `unresolved` 상세 목록(`label`, `reason`)을 그대로 리포트 3번 섹션 하단 "계산 불가" 표에 옮겨 적는다 — 사유를 새로 지어내지 않는다 (5단계 참고).
- 그래도 눈으로 확인하고 싶으면 `get_screenshot`을 추가로 써도 된다.

### 4. 결정론적 검사 실행
정규식 매칭과 WCAG 대비 계산은 눈대중이 아니라 스크립트로 정확히 처리한다. 저장소 루트 기준 다음 명령을 실행한다 (Python 3 필요, 표준 라이브러리만 사용하므로 별도 설치 불필요):
```bash
python3 scripts/kkemi_check.py \
  --nodes <임시경로>/nodes.json \
  --variables <임시경로>/variables.json \
  --contrast-pairs <임시경로>/contrast_pairs.json \
  --output <임시경로>/kkemi_result.json
```
`--output`(기본값 `kkemi_result.json`)에 `naming_violations`, `value_extensions`, `contrast_results` 세 배열을 담은 전체 결과 JSON을 저장한다. `--selftest` 플래그로 Figma 데이터 없이 스크립트 자체 동작을 검증할 수 있다.

### 출력 규격 (공통)
`parse_metadata.py`, `parse_design_context.py`, `kkemi_check.py` 세 스크립트 모두 표준출력에 같은 모양으로 요약을 찍는다:
```
wrote <N> <항목> to <경로>
{ <개수 위주의 짧은 breakdown JSON> }
[해당 시] 상세 목록(예: unresolved detail)
```
전체 결과 JSON을 다시 표준출력에 통째로 덤프하지 않는다 — `--output` 파일에서 필요한 항목만 읽는다. 리포트를 쓸 때는 이 표준출력 요약으로 몇 건인지 먼저 확인하고, 표 작성에 필요한 상세 데이터는 `--output` 파일(`nodes.json`/`contrast_pairs.json`/`kkemi_result.json`)에서 가져온다.

### 5. 리포트 작성
스크립트 출력을 아래 형식의 마크다운으로 정리해 채팅 답변으로 제공한다 (별도 요청이 없는 한 파일로 저장하지 않음):

```markdown
## 🐜 깨미가 Lint 작업을 완료했어요. <파일명/노드명>

### 요약
- 검사한 노드/변수 수, 위반 건수(심각도: high/medium/low/info별 개수)

### 1. 네이밍 규칙 위반
| 이름 | 위치(node id) | 위반 규칙 | 문제 | 수정 제안 |
|---|---|---|---|---|
(naming_violations를 severity high→medium→low 순으로 정렬해서 표로)

### 2. 네이밍 확장 검사 (참고용)
(value_extensions 항목들. 간격 4/8px 배수, 색상 10단위 스텝 관련)

### 3. 접근성 검증
| 요소 | 전경색 | 배경색 | 계산된 대비율 | 기준 | 결과 |
|---|---|---|---|---|---|
(contrast_results를 표로. pass=false인 항목은 굵게 강조)
- 대비 계산이 불가능했던 노드는 하단에 "계산 불가" 목록으로 별도 표기
```

위반이 하나도 없는 섹션은 "위반 없음"이라고 명시하고 표를 생략한다.

## 규칙 갱신

규칙 원문은 `references/`에 스냅샷으로 내장되어 있다 (수집일 2026-07-25). **사용자가 명시적으로 "규칙 갱신해줘"라고 요청한 경우에만** https://www.krds.go.kr/html/site/utility/utility_03.html 과 utility_04.html 을 다시 읽고 두 reference 파일을 갱신한다. 그 외에는 재조회하지 않는다.

## 참고 파일
- `references/naming-rules.md` — 네이밍 규칙 전체 스냅샷 및 어떤 항목을 정규식화했는지
- `references/accessibility-rules.md` — 접근성 규칙 스냅샷, 자동검증 항목과 수동확인 항목 구분
- `scripts/kkemi_check.py` — 실제 검사 로직 (Python 3, 표준 라이브러리만 사용). `--selftest`로 동작 확인 가능
- `scripts/parse_metadata.py` — `get_metadata` XML 응답을 `nodes.json`으로 변환하는 헬퍼 (Python 3, 표준 라이브러리만 사용). `--selftest`로 동작 확인 가능
- `scripts/parse_design_context.py` — `get_design_context`가 반환하는 React+Tailwind 코드를 `contrast_pairs.json`으로 변환하는 헬퍼. 반투명 배경(rgba 오버레이) 알파 합성까지 처리하고, 진짜 그라디언트/이미지 배경처럼 환원 불가능한 경우만 사유와 함께 `unresolved`로 분리한다 (Python 3, 표준 라이브러리만 사용). `--selftest`로 동작 확인 가능
