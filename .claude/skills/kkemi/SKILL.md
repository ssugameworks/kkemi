---
name: kkemi
description: 게임웍스 Product Team의 Figma 디자인이 KRDS(대한민국 정부 디자인시스템) 규칙을 지키는지 검사하는 린터 "깨미". "깨미", "KRDS", "피그마 린트", "네이밍 규칙 검사", "디자인시스템 검사", "명도 대비 검사", "접근성 검사" 등을 Figma URL과 함께 요청할 때 사용.
---

# 깨미 (Kkemi) — KRDS Figma Linter

이 스킬의 전체 워크플로우, 규칙 스냅샷 위치, 리포트 포맷은 이 저장소 루트의 **`AGENTS.md`** 에 정리되어 있다. `AGENTS.md`는 Claude Code뿐 아니라 다른 코딩 에이전트(Codex, Cursor 등)에서도 동일하게 쓸 수 있도록 툴 이름을 특정 클라이언트에 종속시키지 않고 작성되어 있다.

**이 스킬이 트리거되면 먼저 저장소 루트의 `AGENTS.md`를 읽고, 거기 적힌 절차를 그대로 따른다.**

## Claude Code 전용 보충 사항

- `AGENTS.md`에서 `get_metadata`, `get_variable_defs`, `get_design_context`, `get_screenshot`라고 부르는 툴은 Claude Code에서는 각각 `mcp__claude_ai_Figma__get_metadata`, `mcp__claude_ai_Figma__get_variable_defs`, `mcp__claude_ai_Figma__get_design_context`, `mcp__claude_ai_Figma__get_screenshot`로 노출된다.
- 검사 스크립트 실행 시 프로젝트 루트 기준 경로를 그대로 쓴다:
  ```bash
  python3 scripts/kkemi_check.py \
    --nodes <scratchpad>/nodes.json \
    --variables <scratchpad>/variables.json \
    --contrast-pairs <scratchpad>/contrast_pairs.json
  ```
- 규칙/스크립트 원본은 이 폴더가 아니라 저장소 루트의 `references/`, `scripts/`에 있다 (단일 소스 유지, 이 폴더에 복제하지 않음).
