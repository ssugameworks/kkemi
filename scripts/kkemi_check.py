#!/usr/bin/env python3
"""kkemi (깨미): KRDS Figma lint checker — deterministic naming-rule regex checks + WCAG contrast math.

Input JSON shapes (all optional, pass what you have):

nodes.json (from parsed get_metadata XML):
  [{"id": "12:34", "name": "button__primary", "type": "COMPONENT"}, ...]
  type is one of: COMPONENT, COMPONENT_SET, FRAME, PAGE, INSTANCE, TEXT, ...
  Only COMPONENT/COMPONENT_SET and top-level FRAME/PAGE nodes are checked;
  other types are ignored for naming rules.

variables.json (from get_variable_defs):
  {"color/primary/40": "#1B64DA", "spacing/small": 12, "size-height": 44, ...}
  Values that look like hex colors are treated as color tokens; numeric
  values are treated as spacing/size tokens for the value-extension checks.

contrast_pairs.json (assembled by the caller from get_design_context):
  [{"label": "body text on card", "fg": "#111111", "bg": "#FFFFFF", "kind": "text"},
   {"label": "select icon on field", "fg": "#5B6472", "bg": "#FFFFFF", "kind": "nontext"}]
  kind is "text" (4.5:1 threshold) or "nontext" (3:1 threshold).

Run with --selftest to verify the contrast math against known reference values
without needing any Figma data.
"""

import argparse
import json
import re
import sys

FORBIDDEN_COLOR_WORDS = [
    "blue", "red", "green", "yellow", "purple", "orange", "pink",
    "cyan", "magenta", "teal", "indigo", "violet", "brown",
]

FORBIDDEN_ABBREVIATIONS = {
    "bg": "background",
    "xs": "xsmall",
    "sm": "small",
    "md": "medium",
    "lg": "large",
    "xl": "xlarge",
}

ALLOWED_STATES = {
    "default", "hover", "pressed", "focused", "disabled", "error",
    "active", "completed", "selected", "unselected", "indeterminate",
}
FORBIDDEN_STATE = "view"

ALLOWED_SIZES = {"xxsmall", "xsmall", "small", "medium", "large", "xlarge", "xxlarge"}

HEX_RE = re.compile(r"^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$")
COMPONENT_NAME_RE = re.compile(r"^[a-z0-9]+(_{1,2}[a-z0-9]+)*$")
SEGMENT_BAD_CHARS_RE = re.compile(r"[ A-Z]")


def word_hits(name, vocabulary):
    """Return vocabulary words that appear as whole tokens in name (split on _/-/ /)."""
    tokens = re.split(r"[/_\-\s]+", name.lower())
    return [t for t in tokens if t in vocabulary]


def check_component_name(node):
    name = node["name"]
    issues = []
    if not COMPONENT_NAME_RE.match(name):
        issues.append({
            "node_id": node.get("id"),
            "name": name,
            "rule": "component-naming",
            "current": name,
            "issue": "컴포넌트명은 모두 소문자 + 언더바(_)로 공백 대체, 더블언더바(__)로 유형 구분해야 함",
            "suggestion": re.sub(r"[\s\-]+", "_", name.lower()),
            "severity": "high",
        })
    for abbr, full in FORBIDDEN_ABBREVIATIONS.items():
        if abbr in word_hits(name, {abbr}):
            issues.append({
                "node_id": node.get("id"),
                "name": name,
                "rule": "forbidden-abbreviation",
                "current": name,
                "issue": f"금지된 약어 '{abbr}' 사용 — '{full}'로 표기해야 함",
                "suggestion": name.lower().replace(abbr, full),
                "severity": "medium",
            })
    return issues


def check_frame_name(node):
    name = node["name"]
    issues = []
    first_alpha = next((c for c in name if c.isalpha()), None)
    if first_alpha is not None and not first_alpha.isupper():
        issues.append({
            "node_id": node.get("id"),
            "name": name,
            "rule": "frame-naming",
            "current": name,
            "issue": "페이지/프레임명은 영문 첫 글자만 대문자로 시작 권장",
            "suggestion": name[:1].upper() + name[1:] if name else name,
            "severity": "low",
        })
    return issues


def check_variable_name(name, value):
    issues = []
    if SEGMENT_BAD_CHARS_RE.search(name):
        issues.append({
            "name": name,
            "rule": "variable-naming",
            "current": name,
            "issue": "로컬 변수/토큰명은 모두 소문자여야 하며 공백 대신 언더바(_)를 사용해야 함",
            "suggestion": re.sub(r"\s+", "_", name).lower(),
            "severity": "high",
        })

    hits = word_hits(name, set(FORBIDDEN_COLOR_WORDS))
    is_color_value = isinstance(value, str) and HEX_RE.match(value)
    if hits and is_color_value:
        issues.append({
            "name": name,
            "rule": "color-role-naming",
            "current": name,
            "issue": f"색상 토큰명에 색상명({', '.join(hits)}) 사용 금지 — primary/secondary/gray 등 역할 기반 이름 사용",
            "suggestion": None,
            "severity": "high",
        })

    for abbr, full in FORBIDDEN_ABBREVIATIONS.items():
        if abbr in word_hits(name, {abbr}):
            issues.append({
                "name": name,
                "rule": "forbidden-abbreviation",
                "current": name,
                "issue": f"금지된 약어 '{abbr}' 사용 — '{full}'로 표기해야 함",
                "suggestion": name.lower().replace(abbr, full),
                "severity": "medium",
            })

    tokens = re.split(r"[/_\-\s]+", name.lower())
    if FORBIDDEN_STATE in tokens:
        issues.append({
            "name": name,
            "rule": "forbidden-state",
            "current": name,
            "issue": "상태 접미사 '-view'는 사용 금지",
            "suggestion": None,
            "severity": "medium",
        })

    size_like = re.compile(r"^(xs|sm|md|lg|xl)$")
    for t in tokens:
        if size_like.match(t):
            issues.append({
                "name": name,
                "rule": "size-abbreviation",
                "current": name,
                "issue": f"크기 표현 약어 '{t}' 사용 금지 — {sorted(ALLOWED_SIZES)} 중 완전한 표현 사용",
                "suggestion": None,
                "severity": "medium",
            })

    return issues


def check_value_extensions(name, value):
    """Non-mandatory reference checks: spacing 4/8px multiples, color 10-step scale."""
    issues = []
    tokens = re.split(r"[/_\-\s]+", name.lower())
    is_spacing_like = any(t in ("spacing", "space", "gap", "padding", "size") for t in tokens)
    is_color_like = any(t in ("color", "primary", "secondary", "gray") for t in tokens)

    if is_spacing_like and isinstance(value, (int, float)):
        if value % 4 != 0 and value not in (2, 10):
            issues.append({
                "name": name,
                "issue": f"간격 값 {value}가 4px/8px 배수(예외 2, 10)를 따르지 않음 — 권장 사항, 원문 강제 아님",
                "severity": "info",
            })

    if is_color_like:
        m = re.search(r"(\d+)$", name)
        if m:
            step = int(m.group(1))
            if step % 10 != 0 and step % 5 != 0:
                issues.append({
                    "name": name,
                    "issue": f"색상 명도 단계 {step}이 10단위(또는 5단위 세밀도) 규칙에서 벗어남 — 권장 사항, 원문 강제 아님",
                    "severity": "info",
                })

    return issues


def hex_to_rgb(hex_color):
    h = hex_color.lstrip("#")
    if len(h) == 3:
        h = "".join(c * 2 for c in h)
    if len(h) == 8:
        h = h[:6]
    r, g, b = (int(h[i:i + 2], 16) for i in (0, 2, 4))
    return r, g, b


def relative_luminance(hex_color):
    r, g, b = hex_to_rgb(hex_color)

    def channel(c):
        c = c / 255.0
        return c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4

    R, G, B = channel(r), channel(g), channel(b)
    return 0.2126 * R + 0.7152 * G + 0.0722 * B


def contrast_ratio(hex_fg, hex_bg):
    l1 = relative_luminance(hex_fg)
    l2 = relative_luminance(hex_bg)
    lighter, darker = max(l1, l2), min(l1, l2)
    return (lighter + 0.05) / (darker + 0.05)


def check_contrast_pair(pair):
    ratio = contrast_ratio(pair["fg"], pair["bg"])
    threshold = 4.5 if pair.get("kind", "text") == "text" else 3.0
    return {
        "label": pair.get("label"),
        "fg": pair["fg"],
        "bg": pair["bg"],
        "ratio": round(ratio, 2),
        "threshold": threshold,
        "pass": ratio >= threshold,
    }


def run(nodes, variables, contrast_pairs):
    naming_violations = []
    value_extensions = []

    for node in nodes:
        ntype = node.get("type", "")
        if ntype in ("COMPONENT", "COMPONENT_SET"):
            naming_violations.extend(check_component_name(node))
        elif ntype in ("FRAME", "PAGE") and node.get("is_top_level"):
            naming_violations.extend(check_frame_name(node))

    for name, value in variables.items():
        naming_violations.extend(check_variable_name(name, value))
        value_extensions.extend(check_value_extensions(name, value))

    contrast_results = [check_contrast_pair(p) for p in contrast_pairs]

    return {
        "naming_violations": naming_violations,
        "value_extensions": value_extensions,
        "contrast_results": contrast_results,
    }


def selftest():
    cases = [
        ("#000000", "#FFFFFF", 21.0),
        ("#FFFFFF", "#FFFFFF", 1.0),
        ("#767676", "#FFFFFF", 4.54),
    ]
    ok = True
    for fg, bg, expected in cases:
        ratio = round(contrast_ratio(fg, bg), 2)
        status = "OK" if abs(ratio - expected) < 0.05 else "FAIL"
        if status == "FAIL":
            ok = False
        print(f"[{status}] contrast({fg}, {bg}) = {ratio} (expected ~{expected})")

    sample_nodes = [
        {"id": "1:1", "name": "Button__Primary", "type": "COMPONENT"},
        {"id": "1:2", "name": "card_footer", "type": "COMPONENT"},
        {"id": "1:3", "name": "main menu", "type": "FRAME", "is_top_level": True},
    ]
    sample_vars = {
        "color/blue/40": "#1B64DA",
        "color-primary-40": "#1B64DA",
        "spacing-bg": 12,
        "size-xs": 4,
        "state-view": "#000000",
    }
    result = run(sample_nodes, sample_vars, [])
    print(json.dumps(result, ensure_ascii=False, indent=2))
    expected_rules_hit = {
        "component-naming",
        "color-role-naming",
        "forbidden-abbreviation",
        "size-abbreviation",
        "forbidden-state",
        "frame-naming",
    }
    hit_rules = {v["rule"] for v in result["naming_violations"]}
    missing = expected_rules_hit - hit_rules
    if missing:
        print(f"[FAIL] expected rule types not triggered: {missing}")
        ok = False
    else:
        print("[OK] all expected naming rule types triggered on sample data")
    return ok


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--nodes", help="path to nodes.json")
    parser.add_argument("--variables", help="path to variables.json")
    parser.add_argument("--contrast-pairs", help="path to contrast_pairs.json")
    parser.add_argument("--output", default="kkemi_result.json", help="where to write the full result JSON")
    parser.add_argument("--selftest", action="store_true", help="run built-in self-test and exit")
    args = parser.parse_args()

    if args.selftest:
        ok = selftest()
        sys.exit(0 if ok else 1)

    nodes = json.load(open(args.nodes)) if args.nodes else []
    variables = json.load(open(args.variables)) if args.variables else {}
    contrast_pairs = json.load(open(args.contrast_pairs)) if args.contrast_pairs else []

    result = run(nodes, variables, contrast_pairs)

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(result, f, ensure_ascii=False, indent=2)

    severity_counts = {}
    for v in result["naming_violations"]:
        severity_counts[v["severity"]] = severity_counts.get(v["severity"], 0) + 1
    contrast_results = result["contrast_results"]
    contrast_fail = sum(1 for c in contrast_results if not c["pass"])

    total = len(result["naming_violations"]) + len(result["value_extensions"]) + len(contrast_results)
    print(f"wrote {total} check results to {args.output}")
    print(json.dumps({
        "naming_violations": severity_counts,
        "value_extensions": len(result["value_extensions"]),
        "contrast_checked": len(contrast_results),
        "contrast_fail": contrast_fail,
    }, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
