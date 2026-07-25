#!/usr/bin/env python3
"""parse_metadata.py: convert a Figma MCP get_metadata response into nodes.json for kkemi_check.py.

get_metadata returns pseudo-XML like:
  <section id="194:30255" name="web" x="-2382" y="3984" width="35479" height="22124">
    <frame id="221:12411" name="home작업실" ...>
      <text id="289:19949" name="..." .../>
      <instance id="289:19950" name="..." .../>
    </frame>
    <symbol id="221:12675" name="activity" .../>
  </section>

When the response is too large for the agent's context window, Claude Code saves it to a
tool-results/*.txt file shaped as a JSON array: [{"type": "text", "text": "<section ...>...\
"}, {"type": "text", "text": "IMPORTANT: ..."}]. This script accepts either that raw file or
a plain XML file/string directly, so the agent does not need to hand-roll parsing code for
every large frame.

Usage:
  python3 scripts/parse_metadata.py --input <path to get_metadata output> --output nodes.json

Run with --selftest to verify parsing against an embedded sample without needing Figma data.
"""

import argparse
import json
import sys
import xml.etree.ElementTree as ET

# Figma's REST API node types are upper snake case (FRAME, COMPONENT, ...). This MCP's
# get_metadata XML uses lowercase tag names that mostly match 1:1; "symbol" is this tool's
# tag for a master component (not a standard Figma API type), observed empirically rather
# than documented, so it's mapped to COMPONENT here.
TAG_TYPE_MAP = {
    "document": "DOCUMENT",
    "canvas": "PAGE",
    "page": "PAGE",
    "section": "SECTION",
    "frame": "FRAME",
    "group": "GROUP",
    "vector": "VECTOR",
    "boolean_operation": "BOOLEAN_OPERATION",
    "star": "STAR",
    "line": "LINE",
    "ellipse": "ELLIPSE",
    "regular_polygon": "REGULAR_POLYGON",
    "rectangle": "RECTANGLE",
    "rounded-rectangle": "RECTANGLE",
    "rounded_rectangle": "RECTANGLE",
    "text": "TEXT",
    "slice": "SLICE",
    "component": "COMPONENT",
    "component_set": "COMPONENT_SET",
    "symbol": "COMPONENT",
    "symbol_set": "COMPONENT_SET",
    "instance": "INSTANCE",
    "sticky": "STICKY",
    "shape_with_text": "SHAPE_WITH_TEXT",
    "connector": "CONNECTOR",
    "washi_tape": "WASHI_TAPE",
    "table": "TABLE",
    "table_cell": "TABLE_CELL",
    "highlight": "HIGHLIGHT",
    "widget": "WIDGET",
}


def node_type(tag):
    return TAG_TYPE_MAP.get(tag.lower(), tag.upper())


def extract_xml_text(raw):
    """raw is the file content. Handle both the [{"type","text"}, ...] tool-result shape
    and a plain XML string."""
    stripped = raw.strip()
    if stripped.startswith("<"):
        return stripped
    try:
        data = json.loads(raw)
    except json.JSONDecodeError:
        return stripped
    if isinstance(data, list):
        for item in data:
            text = item.get("text", "") if isinstance(item, dict) else ""
            if text.strip().startswith("<"):
                return text.strip()
        raise ValueError("no XML-shaped text field found in tool-result JSON array")
    raise ValueError("unrecognized get_metadata output shape")


def parse_nodes(xml_text):
    root = ET.fromstring(xml_text)
    nodes = []

    def walk(el, is_top_level):
        nodes.append(
            {
                "id": el.get("id"),
                "name": el.get("name", ""),
                "type": node_type(el.tag),
                "is_top_level": is_top_level,
            }
        )
        for child in el:
            walk(child, is_top_level=(el is root))

    walk(root, is_top_level=False)
    return nodes


def selftest():
    sample = """
    <section id="1:1" name="web">
      <frame id="1:2" name="main menu">
        <text id="1:3" name="hello" />
        <instance id="1:4" name="Button/text" />
      </frame>
      <symbol id="1:5" name="Button__Primary" />
    </section>
    """
    nodes = parse_nodes(sample)
    by_id = {n["id"]: n for n in nodes}
    ok = True

    checks = [
        ("1:1", "SECTION", False),
        ("1:2", "FRAME", True),
        ("1:3", "TEXT", False),
        ("1:4", "INSTANCE", False),
        ("1:5", "COMPONENT", True),
    ]
    for node_id, expected_type, expected_top in checks:
        n = by_id.get(node_id)
        if not n:
            print(f"[FAIL] missing node {node_id}")
            ok = False
            continue
        if n["type"] != expected_type or n["is_top_level"] != expected_top:
            print(f"[FAIL] {node_id}: got {n}, expected type={expected_type} is_top_level={expected_top}")
            ok = False
        else:
            print(f"[OK] {node_id}: type={n['type']} is_top_level={n['is_top_level']}")

    wrapped = json.dumps([{"type": "text", "text": sample}, {"type": "text", "text": "IMPORTANT: ..."}])
    if parse_nodes(extract_xml_text(wrapped)) != nodes:
        print("[FAIL] tool-result JSON array unwrapping produced different nodes")
        ok = False
    else:
        print("[OK] tool-result JSON array unwrapping matches raw XML parsing")

    return ok


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", help="path to get_metadata output (raw tool-result file or plain XML)")
    parser.add_argument("--output", default="nodes.json", help="where to write nodes.json (default: ./nodes.json)")
    parser.add_argument("--selftest", action="store_true", help="run built-in self-test and exit")
    args = parser.parse_args()

    if args.selftest:
        ok = selftest()
        sys.exit(0 if ok else 1)

    if not args.input:
        parser.error("--input is required unless --selftest is given")

    with open(args.input, encoding="utf-8") as f:
        raw = f.read()

    xml_text = extract_xml_text(raw)
    nodes = parse_nodes(xml_text)

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(nodes, f, ensure_ascii=False, indent=2)

    type_counts = {}
    for n in nodes:
        type_counts[n["type"]] = type_counts.get(n["type"], 0) + 1
    print(f"wrote {len(nodes)} nodes to {args.output}")
    print(json.dumps(type_counts, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
