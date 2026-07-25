#!/usr/bin/env python3
"""parse_design_context.py: convert a Figma MCP get_design_context response into
contrast_pairs.json for kkemi_check.py.

get_design_context does NOT return structured color data — it returns generated
React+Tailwind JSX code, e.g.:

  <div className="bg-[var(--color/gray/50,#f2f2f3)]" data-node-id="1:1" data-name="card">
    <p className="text-[color:var(--color/blue/800,#004bb2)]" data-node-id="1:2" data-name="body">
      hello
    </p>
  </div>

This script extracts, for each element carrying a `data-node-id`, its own text/background
color and — since Figma text nodes rarely set their own fill — walks up the JSX using
indentation as an approximation of DOM nesting to find the nearest ancestor's background.
This is a heuristic, not a real JSX/DOM parser: it assumes the code is pretty-printed with
consistent indentation per nesting level (true for what this MCP tool emits in practice) and
does not handle a single element's opening tag spanning multiple lines.

Backgrounds come in three flavors, and each is handled differently:
  - solid: a plain hex (`bg-[...#f2f2f3]`, `bg-white`) or a Tailwind bracket whose CSS var
    fallback is a hex or a named color (`bg-[var(--x,white)]`) -> used directly.
  - flat translucent: a single-color rgba()/rgb() fill (`bg-[var(--x,rgba(255,255,255,0.05))]`)
    or an inline `style={{ backgroundImage: "linear-gradient(...)" }}` whose gradient stops
    are all the *same* color (Figma's export shape for a plain fill+opacity) -> composited
    with true alpha blending over whatever solid color is found by continuing the ancestor
    walk from that element's parent.
  - real gradient / unknown image: an inline background whose stops actually differ, or a
    style background with no parsable color function at all -> cannot be reduced to one
    color. The ancestor walk stops here (does NOT fall through to a grandparent's color,
    which would silently produce a confidently-wrong pair) and the element is reported as
    unresolved with a reason, for a human to check with get_screenshot instead.

Pairs are only emitted when both a foreground and a composited background hex can be
resolved. Anything else is skipped and returned in a separate `unresolved` list with a
reason, per the "애매하면 건너뛴다" rule this project follows elsewhere — the difference
from a plain skip is that the caller gets a reason string to fold into the report's
"계산 불가" section instead of the element silently vanishing.

Usage:
  python3 scripts/parse_design_context.py \
    --input <get_design_context response or saved tool-result file> \
    --nodes <nodes.json produced by parse_metadata.py> \
    --output contrast_pairs.json

`--nodes` is optional but recommended: node ids present in it are classified precisely as
TEXT / icon-like (kind="text"/"nontext") vs. containers (skipped). get_metadata does not
expand component-instance internals, so most real text in a design that uses components
won't have an entry in nodes.json — for any id NOT found in nodes_map, this script falls
back to a structural heuristic (an element with its own resolvable text color is treated as
text; an <svg> with its own color is treated as nontext) rather than skipping it outright.

Run with --selftest to verify parsing against embedded samples without needing Figma data.
"""

import argparse
import json
import re
import sys

HEX_IN_BRACKETS_RE = re.compile(r"#[0-9a-fA-F]{3,8}")
NAMED_COLORS = {"white": "#ffffff", "black": "#000000"}
NAMED_FALLBACK_RE = re.compile(r",\s*(white|black)\s*\)?\s*$")

TAG_RE = re.compile(r"<([A-Za-z][\w.]*)\b([^>]*)>")
NODE_ID_RE = re.compile(r'data-node-id="([^"]+)"')
DATA_NAME_RE = re.compile(r'data-name="([^"]+)"')
CLASSNAME_RE = re.compile(r'className=[^"]*?"([^"]*)"')

STYLE_ATTR_RE = re.compile(r"style=\{\{([^}]*)\}\}")
BG_IMAGE_RE = re.compile(r'backgroundImage:\s*"([^"]*)"')
COLOR_FN_RE = re.compile(r"rgba?\(([^)]+)\)")
# CSS multiple-background layers are comma separated at the top level. A stop boundary
# inside one layer always looks like "<fn>(...) <percent>, <fn>(...)" (a space before the
# comma), while the boundary *between* layers looks like "...)),  <fn>(" (no space before
# the comma, since the previous layer's outer gradient function just closed). Requiring the
# comma to immediately follow ")" with no space is what keeps this from splitting inside a
# single layer's color stops.
LAYER_SPLIT_RE = re.compile(r"\),\s*(?=linear-gradient|radial-gradient|rgba?\()")

TEXT_TYPE = "TEXT"
NONTEXT_TYPES = {"VECTOR", "BOOLEAN_OPERATION", "LINE"}


def hex_to_rgb(hex_color):
    h = hex_color.lstrip("#")
    if len(h) == 3:
        h = "".join(c * 2 for c in h)
    if len(h) == 8:
        h = h[:6]
    r, g, b = (int(h[i:i + 2], 16) for i in (0, 2, 4))
    return r, g, b


def rgb_to_hex(rgb):
    return "#%02x%02x%02x" % tuple(rgb)


def parse_color_fn(text):
    """'26, 122, 255, 0.9' or '0, 53, 128' -> (r, g, b, alpha)."""
    parts = [p.strip() for p in text.split(",")]
    r, g, b = (int(float(p)) for p in parts[:3])
    a = float(parts[3]) if len(parts) > 3 else 1.0
    return r, g, b, a


def parse_background_image(value):
    """Split a backgroundImage value into layers and classify it.

    Returns ("flat", [(r,g,b,a), ...]) top-to-bottom paint order if every layer's color
    stops are a single repeated color (Figma's shape for a flat translucent fill); returns
    ("gradient", (first_hex, last_hex)) if any layer's stops actually differ (a real
    gradient that can't be reduced to one color); returns (None, None) if no rgb()/rgba()
    color function was found at all (e.g. a pure image url background).
    """
    layers = LAYER_SPLIT_RE.split(value)
    flat_layers = []
    for layer in layers:
        stops = COLOR_FN_RE.findall(layer)
        if not stops:
            continue
        parsed = [parse_color_fn(s) for s in stops]
        distinct_rgb = {p[:3] for p in parsed}
        if len(distinct_rgb) > 1:
            return "gradient", (rgb_to_hex(parsed[0][:3]), rgb_to_hex(parsed[-1][:3]))
        flat_layers.append(parsed[0])
    if not flat_layers:
        return None, None
    return "flat", flat_layers


def composite_over(base_rgb, layers_top_to_bottom):
    """layers_top_to_bottom[0] is the topmost (painted last per CSS multi-background rules,
    i.e. first-listed); the last entry is the bottom-most (painted first, right on the
    base). Composite bottom-to-top so the topmost layer ends up on top of the result."""
    result = base_rgb
    for r, g, b, a in reversed(layers_top_to_bottom):
        result = tuple(round(a * c + (1 - a) * base_c) for c, base_c in zip((r, g, b), result))
    return result


def parse_class_fill(class_str, prefix):
    """Own solid/flat fill declared via a Tailwind class (`{prefix}-white`, `{prefix}-[...]`)."""
    for name, hexval in NAMED_COLORS.items():
        if re.search(rf"\b{prefix}-{name}\b", class_str):
            return {"type": "solid", "hex": hexval}
    for bracket in re.findall(rf"{prefix}-\[([^\]]*)\]", class_str):
        m = HEX_IN_BRACKETS_RE.search(bracket)
        if m:
            return {"type": "solid", "hex": m.group(0).lower()}
        m = NAMED_FALLBACK_RE.search(bracket)
        if m:
            return {"type": "solid", "hex": NAMED_COLORS[m.group(1)]}
        m = COLOR_FN_RE.search(bracket)
        if m:
            return {"type": "flat", "layers": [parse_color_fn(m.group(1))]}
    return None


def parse_style_bg_fill(attrs):
    """Own fill declared via `style={{ backgroundImage: "..." }}`, if any."""
    style_match = STYLE_ATTR_RE.search(attrs)
    if not style_match:
        return None
    bg_match = BG_IMAGE_RE.search(style_match.group(1))
    if not bg_match:
        return None
    kind, value = parse_background_image(bg_match.group(1))
    if kind == "flat":
        return {"type": "flat", "layers": value}
    if kind == "gradient":
        return {"type": "gradient", "first": value[0], "last": value[1]}
    return {"type": "unknown"}


def simple_fg_hex(class_str):
    fill = parse_class_fill(class_str, "text")
    return fill["hex"] if fill and fill["type"] == "solid" else None


def own_bg_fill(class_str, attrs):
    return parse_style_bg_fill(attrs) or parse_class_fill(class_str, "bg")


def extract_code_text(raw):
    """raw is the file content: either the [{"type","text"}, ...] tool-result shape or
    the plain generated code string."""
    stripped = raw.strip()
    try:
        data = json.loads(raw)
    except json.JSONDecodeError:
        return stripped
    if isinstance(data, list):
        candidates = [item.get("text", "") for item in data if isinstance(item, dict)]
        with_node_ids = [t for t in candidates if "data-node-id" in t]
        if with_node_ids:
            return max(with_node_ids, key=len)
        raise ValueError("no text field containing data-node-id found in tool-result JSON array")
    raise ValueError("unrecognized get_design_context output shape")


def parse_elements(code_text):
    elements = []
    for line in code_text.splitlines():
        indent = len(line) - len(line.lstrip(" "))
        for tag_match in TAG_RE.finditer(line):
            tag = tag_match.group(1)
            attrs = tag_match.group(2)
            node_id_match = NODE_ID_RE.search(attrs)
            name_match = DATA_NAME_RE.search(attrs)
            class_match = CLASSNAME_RE.search(attrs)
            class_str = class_match.group(1) if class_match else ""
            elements.append({
                "indent": indent,
                "tag": tag,
                "id": node_id_match.group(1) if node_id_match else None,
                "name": name_match.group(1) if name_match else None,
                "fg": simple_fg_hex(class_str),
                "bg_fill": own_bg_fill(class_str, attrs),
            })
    return elements


def nearest_fg(elements, idx):
    if elements[idx]["fg"]:
        return elements[idx]["fg"]
    indent = elements[idx]["indent"]
    i = idx - 1
    while i >= 0:
        if elements[i]["indent"] < indent:
            if elements[i]["fg"]:
                return elements[i]["fg"]
            indent = elements[i]["indent"]
        i -= 1
    return None


def resolve_bg(elements, idx, memo):
    """Returns (hex_or_None, reason_or_None). `reason` is only set when hex is None, so
    callers can report *why* (real gradient, unknown image, ...) instead of a bare skip."""
    if idx in memo:
        return memo[idx]

    fill = elements[idx]["bg_fill"]
    if fill is None:
        result = resolve_ancestor_bg(elements, idx, memo)
    elif fill["type"] == "solid":
        result = (fill["hex"], None)
    elif fill["type"] == "gradient":
        result = (None, f"그라디언트 배경(~{fill['first']} ~ {fill['last']}) — 값이 구간마다 달라 자동 계산 불가")
    elif fill["type"] == "unknown":
        result = (None, "배경 스타일에서 색상 값을 판독할 수 없음 (이미지 등)")
    else:  # "flat": translucent, needs a resolvable base beneath it
        base_hex, base_reason = resolve_ancestor_bg(elements, idx, memo)
        if base_hex is None:
            result = (None, base_reason or "반투명 배경의 바탕색을 찾지 못함")
        else:
            composited = composite_over(hex_to_rgb(base_hex), fill["layers"])
            result = (rgb_to_hex(composited), None)

    memo[idx] = result
    return result


def resolve_ancestor_bg(elements, idx, memo):
    indent = elements[idx]["indent"]
    i = idx - 1
    while i >= 0:
        if elements[i]["indent"] < indent:
            return resolve_bg(elements, i, memo)
        i -= 1
    return None, None


def classify(el, nodes_map):
    if nodes_map is not None and el["id"] in nodes_map:
        ntype = nodes_map[el["id"]]
        if ntype == TEXT_TYPE:
            return "text"
        if ntype in NONTEXT_TYPES:
            return "nontext"
        return None
    # Not in nodes_map at all -> get_metadata never expanded this node (typically because
    # it lives inside a component instance). Fall back to a structural guess instead of
    # skipping it outright.
    if el["tag"].lower() == "svg" and el["fg"]:
        return "nontext"
    if el["fg"]:
        return "text"
    return None


def build_pairs(elements, nodes_map):
    pairs = []
    unresolved = []
    memo = {}
    for idx, el in enumerate(elements):
        if not el["id"]:
            continue
        kind = classify(el, nodes_map)
        if kind is None:
            continue

        fg = nearest_fg(elements, idx)
        bg, reason = resolve_ancestor_bg(elements, idx, memo)
        label = el["name"] or el["id"]

        if fg and bg:
            pairs.append({"label": label, "fg": fg, "bg": bg, "kind": kind})
        else:
            why = reason if bg is None and reason else ("전경색 판독 불가" if not fg else "배경 판독 불가")
            unresolved.append({"label": label, "reason": why})
    return pairs, unresolved


def selftest():
    sample = """
    <div className="bg-[var(--color/gray/50,#f2f2f3)]" data-node-id="1:1" data-name="card">
      <p className="text-[color:var(--color/blue/800,#004bb2)]" data-node-id="1:2" data-name="body">
        hello
      </p>
      <div className="text-[length:var(--font/body,14px)]" data-node-id="1:3" data-name="no-color-text">
        inherits color from card? no ancestor text color set, should be skipped
      </div>
      <svg data-node-id="1:4" data-name="icon" className="text-[#5b6472]" />
      <div className="bg-white" data-node-id="1:5" data-name="unrelated-frame" />
      <p className="text-[color:var(--color/static/white,white)]" data-node-id="1:6" data-name="word-fallback-text">
        word fallback
      </p>
    </div>
    """
    nodes_map = {
        "1:1": "FRAME", "1:2": "TEXT", "1:3": "TEXT", "1:4": "VECTOR", "1:5": "FRAME", "1:6": "TEXT",
    }
    elements = parse_elements(sample)
    ok = True

    pairs, unresolved = build_pairs(elements, nodes_map)
    by_label = {p["label"]: p for p in pairs}

    expected = {
        "body": {"fg": "#004bb2", "bg": "#f2f2f3", "kind": "text"},
        "icon": {"fg": "#5b6472", "bg": "#f2f2f3", "kind": "nontext"},
        "word-fallback-text": {"fg": "#ffffff", "bg": "#f2f2f3", "kind": "text"},
    }
    for label, exp in expected.items():
        got = by_label.get(label)
        if got != {**exp, "label": label}:
            print(f"[FAIL] {label}: got {got}, expected {exp}")
            ok = False
        else:
            print(f"[OK] {label}: {got}")

    if "no-color-text" in by_label or "unrelated-frame" in by_label:
        print("[FAIL] no-color-text/unrelated-frame should have been skipped")
        ok = False
    else:
        print("[OK] no-color-text and unrelated-frame correctly skipped")

    # --- translucent style background composited over a resolvable solid ancestor ---
    button_sample = """
    <div className="bg-[#fafafa]" data-node-id="2:1" data-name="section">
      <button style={{ backgroundImage: "linear-gradient(90deg, rgba(26, 122, 255, 0.9) 0%, rgba(26, 122, 255, 0.9) 100%), linear-gradient(90deg, rgba(255, 255, 255, 0.7) 0%, rgba(255, 255, 255, 0.7) 100%)" }} data-node-id="2:2" data-name="detail-button">
        <p className="text-[color:var(--color/gray/50,#f2f2f3)]" data-node-id="2:3" data-name="detail-button-label">
          click me
        </p>
      </button>
    </div>
    """
    button_nodes_map = {"2:1": "FRAME", "2:2": "INSTANCE", "2:3": "TEXT"}
    button_elements = parse_elements(button_sample)
    button_pairs, button_unresolved = build_pairs(button_elements, button_nodes_map)
    by_label2 = {p["label"]: p for p in button_pairs}
    got = by_label2.get("detail-button-label")
    expected_bg = "#3187ff"  # hand-verified: 0.7 white then 0.9 blue over #fafafa
    if got and got["bg"] == expected_bg and got["fg"] == "#f2f2f3":
        print(f"[OK] detail-button-label: translucent style background composited to {expected_bg}")
    else:
        print(f"[FAIL] detail-button-label: got {got}, expected bg={expected_bg}")
        ok = False

    # --- translucent Tailwind bracket fill (rgba fallback, no hex) over a solid ancestor ---
    card_sample = """
    <div className="bg-[#fafafa]" data-node-id="3:1" data-name="section">
      <div className="bg-[var(--color/semantic/shadow/white-a5,rgba(255,255,255,0.05))]" data-node-id="3:2" data-name="card">
        <p className="text-[color:var(--color/blue/950,#000b1a)]" data-node-id="3:3" data-name="card-title">
          title
        </p>
      </div>
    </div>
    """
    card_nodes_map = {"3:1": "FRAME", "3:2": "INSTANCE", "3:3": "TEXT"}
    card_elements = parse_elements(card_sample)
    card_pairs, _ = build_pairs(card_elements, card_nodes_map)
    got = {p["label"]: p for p in card_pairs}.get("card-title")
    if got and got["bg"] == "#fafafa":
        print("[OK] card-title: rgba(255,255,255,0.05) over #fafafa correctly composites back to #fafafa")
    else:
        print(f"[FAIL] card-title: got {got}, expected bg=#fafafa")
        ok = False

    # --- real (non-uniform) gradient: must stop the climb and report a reason, not fall
    #     through to whatever solid color happens to sit further up ---
    gradient_sample = """
    <div className="bg-[#fafafa]" data-node-id="4:1" data-name="page-root">
      <div style={{ backgroundImage: "linear-gradient(180deg, rgb(12, 12, 13) 0%, rgb(34, 48, 99) 100%)" }} data-node-id="4:2" data-name="hero">
        <p className="text-[color:var(--color/static/white,white)]" data-node-id="4:3" data-name="hero-heading">
          headline
        </p>
      </div>
    </div>
    """
    gradient_nodes_map = {"4:1": "FRAME", "4:2": "FRAME", "4:3": "TEXT"}
    gradient_elements = parse_elements(gradient_sample)
    gradient_pairs, gradient_unresolved = build_pairs(gradient_elements, gradient_nodes_map)
    if any(p["label"] == "hero-heading" for p in gradient_pairs):
        print("[FAIL] hero-heading should NOT resolve to a pair (real gradient background)")
        ok = False
    else:
        reason = next((u["reason"] for u in gradient_unresolved if u["label"] == "hero-heading"), None)
        if reason and "그라디언트" in reason:
            print(f"[OK] hero-heading correctly unresolved: {reason}")
        else:
            print(f"[FAIL] hero-heading unresolved but with unexpected/missing reason: {reason}")
            ok = False

    # --- id missing from nodes_map entirely (instance-internal node) -> fallback heuristic ---
    fallback_pairs, _ = build_pairs(elements, {"1:1": "FRAME"})  # only card is mapped
    if {p["label"] for p in fallback_pairs} >= {"body", "icon", "word-fallback-text"}:
        print("[OK] ids missing from nodes_map fall back to own-fg heuristic instead of being skipped")
    else:
        print(f"[FAIL] fallback classification incomplete: {[p['label'] for p in fallback_pairs]}")
        ok = False

    return ok


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--input", help="path to get_design_context output (raw tool-result file or plain code)")
    parser.add_argument("--nodes", help="path to nodes.json from parse_metadata.py (recommended)")
    parser.add_argument("--output", default="contrast_pairs.json", help="where to write contrast_pairs.json")
    parser.add_argument("--selftest", action="store_true", help="run built-in self-test and exit")
    args = parser.parse_args()

    if args.selftest:
        ok = selftest()
        sys.exit(0 if ok else 1)

    if not args.input:
        parser.error("--input is required unless --selftest is given")

    with open(args.input, encoding="utf-8") as f:
        raw = f.read()
    code_text = extract_code_text(raw)
    elements = parse_elements(code_text)

    nodes_map = None
    if args.nodes:
        nodes = json.load(open(args.nodes))
        nodes_map = {n["id"]: n["type"] for n in nodes}

    pairs, unresolved = build_pairs(elements, nodes_map)

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(pairs, f, ensure_ascii=False, indent=2)

    print(f"wrote {len(pairs)} contrast pairs to {args.output}")
    print(json.dumps({"resolved": len(pairs), "unresolved": len(unresolved)}, ensure_ascii=False, indent=2))
    if unresolved:
        print("unresolved detail (fold into report's 계산 불가 section):")
        print(json.dumps(unresolved, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
