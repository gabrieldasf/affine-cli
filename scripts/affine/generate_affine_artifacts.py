#!/usr/bin/env python3
"""Generate AFFiNE Printing Press spec and API inventory from qtz-affine."""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import subprocess
from datetime import datetime, timezone
from pathlib import Path


DEFAULT_SOURCE = Path(r"D:\Apps\QTZ-Apps\qtz-affine")
DEFAULT_OUTPUT = Path("specs") / "affine"


def read_text(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def git_head(repo: Path) -> str:
    try:
        return subprocess.check_output(
            ["git", "-C", str(repo), "rev-parse", "HEAD"],
            text=True,
            stderr=subprocess.DEVNULL,
        ).strip()
    except Exception:
        return ""


def extract_root_fields(schema: str, root_name: str) -> list[dict[str, object]]:
    block = extract_type_block(schema, root_name)
    fields: list[dict[str, object]] = []
    for line in block.splitlines():
        clean = line.split("#", 1)[0].strip()
        if not clean or clean.startswith(("}", "@")):
            continue
        match = re.match(r"^([A-Za-z_][A-Za-z0-9_]*)\s*(?:\((.*?)\))?\s*:\s*(.+)$", clean)
        if not match:
            continue
        args = parse_args(match.group(2) or "")
        fields.append(
            {
                "name": match.group(1),
                "args": args,
                "return_type": match.group(3).strip(),
            }
        )
    return fields


def extract_type_block(schema: str, root_name: str) -> str:
    match = re.search(rf"^type\s+{re.escape(root_name)}\b[^\{{]*\{{", schema, re.MULTILINE)
    if not match:
        return ""
    start = match.end()
    depth = 1
    i = start
    while i < len(schema):
        char = schema[i]
        if char == "{":
            depth += 1
        elif char == "}":
            depth -= 1
            if depth == 0:
                return schema[start:i]
        i += 1
    return ""


def parse_args(raw: str) -> list[dict[str, object]]:
    if not raw.strip():
        return []
    args: list[dict[str, object]] = []
    current = []
    depth = 0
    for char in raw:
        if char in "([{":
            depth += 1
        elif char in ")]}":
            depth -= 1
        if char == "," and depth == 0:
            add_arg(args, "".join(current).strip())
            current = []
        else:
            current.append(char)
    add_arg(args, "".join(current).strip())
    return args


def add_arg(args: list[dict[str, object]], item: str) -> None:
    if not item:
        return
    match = re.match(r"^([A-Za-z_][A-Za-z0-9_]*)\s*:\s*([^=]+?)(?:\s*=\s*(.+))?$", item)
    if not match:
        return
    gql_type = match.group(2).strip()
    args.append(
        {
            "name": match.group(1),
            "type": gql_type,
            "required": gql_type.endswith("!"),
            "default": (match.group(3) or "").strip(),
        }
    )


def strip_comments(document: str) -> str:
    return "\n".join(line for line in document.splitlines() if not line.lstrip().startswith("#"))


def split_definitions(document: str) -> list[tuple[str, str, str]]:
    text = strip_comments(document)
    definitions: list[tuple[str, str, str]] = []
    pattern = re.compile(r"(?m)^\s*(query|mutation|fragment)\s+([A-Za-z_][A-Za-z0-9_]*)\b")
    matches = list(pattern.finditer(text))
    for idx, match in enumerate(matches):
        end = matches[idx + 1].start() if idx + 1 < len(matches) else len(text)
        definitions.append((match.group(1), match.group(2), text[match.start() : end].strip()))
    return definitions


def used_fragments(document: str) -> set[str]:
    return {
        name
        for name in re.findall(r"\.\.\.\s*([A-Za-z_][A-Za-z0-9_]*)", document)
        if name != "on"
    }


def with_fragments(document: str, fragment_map: dict[str, str]) -> str:
    selected: list[str] = []
    seen: set[str] = set()

    def add_needed(doc: str) -> None:
        for name in sorted(used_fragments(doc)):
            if name in seen or name not in fragment_map:
                continue
            seen.add(name)
            frag = fragment_map[name]
            selected.append(frag)
            add_needed(frag)

    add_needed(document)
    if not selected:
        return document.strip()
    return document.strip() + "\n\n" + "\n\n".join(selected)


def operation_root_field(document: str) -> str:
    first_brace = document.find("{")
    if first_brace < 0:
        return ""
    i = first_brace + 1
    while i < len(document):
        char = document[i]
        if char.isspace():
            i += 1
            continue
        if char == "#":
            while i < len(document) and document[i] != "\n":
                i += 1
            continue
        break
    match = re.match(r"([A-Za-z_][A-Za-z0-9_]*)", document[i:])
    return match.group(1) if match else ""


def operation_group(path: Path, operation_name: str, root_field: str) -> str:
    stem = path.stem.lower().replace("_", "-")
    prefixes = [
        "admin",
        "blob",
        "comment",
        "copilot",
        "doc",
        "invite",
        "member",
        "notification",
        "payment",
        "reply",
        "subscription",
        "user",
        "workspace",
    ]
    for prefix in prefixes:
        if stem == prefix or stem.startswith(prefix + "-"):
            return pluralize(prefix)
    for prefix in prefixes:
        if operation_name.lower().startswith(prefix):
            return pluralize(prefix)
    if root_field:
        snake = re.sub(r"(?<!^)([A-Z])", r"-\1", root_field).lower()
        for prefix in prefixes:
            if snake == prefix or snake.startswith(prefix + "-"):
                return pluralize(prefix)
    return "graphql"


def pluralize(value: str) -> str:
    if value.endswith("y"):
        return value[:-1] + "ies"
    if value.endswith("s"):
        return value
    return value + "s"


def command_name(operation_name: str) -> str:
    kebab = re.sub(r"(?<!^)([A-Z])", r"-\1", operation_name).replace("_", "-")
    return re.sub(r"[^a-z0-9-]+", "-", kebab.lower()).strip("-") or "run"


def yaml_scalar(value: object, indent: int = 0) -> list[str]:
    pad = " " * indent
    if value is None:
        return [pad + '""']
    if isinstance(value, bool):
        return [pad + ("true" if value else "false")]
    if isinstance(value, (int, float)):
        return [pad + str(value)]
    if isinstance(value, dict):
        return [pad + "{}"] if not value else [pad + json.dumps(value, ensure_ascii=False)]
    if isinstance(value, list):
        return [pad + "[]"] if not value else [pad + json.dumps(value, ensure_ascii=False)]
    text = str(value)
    if "\n" in text:
        lines = [pad + "|-"]
        lines.extend(pad + "  " + line for line in text.splitlines())
        return lines
    escaped = text.replace("\\", "\\\\").replace('"', '\\"')
    return [pad + f'"{escaped}"']


def emit_param(lines: list[str], param: dict[str, object], indent: int) -> None:
    pad = " " * indent
    lines.append(f"{pad}- name: {param['name']}")
    lines.append(f"{pad}  type: {param.get('type', 'string')}")
    if param.get("required"):
        lines.append(f"{pad}  required: true")
    if "default" in param:
        lines.append(f"{pad}  default:")
        lines.extend(yaml_scalar(param["default"], indent + 4))
    desc = param.get("description")
    if desc:
        lines.append(f"{pad}  description: \"{str(desc).replace(chr(34), chr(39))}\"")


def generate_yaml(operations: list[dict[str, object]], base_url: str) -> str:
    groups: dict[str, list[dict[str, object]]] = {}
    for op in operations:
        groups.setdefault(str(op["group"]), []).append(op)

    lines = [
        "name: affine",
        "display_name: AFFiNE",
        "description: AFFiNE self-hosted GraphQL API wrapper generated from qtz-affine schema.gql and client .gql documents.",
        "cli_description: Operate AFFiNE GraphQL from the terminal.",
        'version: "0.1.0"',
        "spec_source: official",
        "category: productivity",
        'website_url: "https://affine.quartzo.ai"',
        f'base_url: "{base_url}"',
        'graphql_endpoint_path: "/graphql"',
        "",
        "auth:",
        "  type: bearer_token",
        "  header: Authorization",
        "  in: header",
        "  env_vars:",
        "    - AFFINE_TOKEN",
        '  verify_query: "{ currentUser { id email name } }"',
        "",
        "config:",
        "  format: toml",
        '  path: "~/.config/affine-cli/config.toml"',
        "",
        "resources:",
        "  graphql:",
        "    description: Raw AFFiNE GraphQL operations.",
        "    endpoints:",
        "      execute:",
        "        method: POST",
        "        path: /graphql",
        "        description: Execute any AFFiNE GraphQL document.",
        "        body:",
    ]
    emit_param(
        lines,
        {
            "name": "operationName",
            "type": "string",
            "required": False,
            "description": "Optional GraphQL operation name.",
        },
        10,
    )
    emit_param(
        lines,
        {
            "name": "query",
            "type": "string",
            "required": True,
            "description": "GraphQL query or mutation document.",
        },
        10,
    )
    emit_param(
        lines,
        {
            "name": "variables",
            "type": "object",
            "required": False,
            "description": "GraphQL variables as JSON.",
        },
        10,
    )
    lines.extend(
        [
            "        response:",
            "          type: object",
            "          item: GraphQLResponse",
            "",
        ]
    )

    for group in sorted(groups):
        if group == "graphql":
            endpoint_indent = "      "
            group_lines = lines
        else:
            lines.extend(
                [
                    f"  {group}:",
                    f"    description: AFFiNE {group.replace('-', ' ')} GraphQL operations.",
                    "    endpoints:",
                ]
            )
            endpoint_indent = "      "
            group_lines = lines
        used_names: set[str] = set()
        for op in sorted(groups[group], key=lambda item: str(item["name"]).lower()):
            name = command_name(str(op["name"]))
            if name in used_names:
                suffix = 2
                while f"{name}-{suffix}" in used_names:
                    suffix += 1
                name = f"{name}-{suffix}"
            used_names.add(name)
            group_lines.extend(
                [
                    f"{endpoint_indent}{name}:",
                    "        method: POST",
                    "        path: /graphql",
                    f"        description: Run AFFiNE {op['kind']} {op['name']}.",
                    "        body:",
                ]
            )
            emit_param(
                group_lines,
                {
                    "name": "operationName",
                    "type": "string",
                    "required": True,
                    "default": op["name"],
                    "description": "GraphQL operation name.",
                },
                10,
            )
            emit_param(
                group_lines,
                {
                    "name": "query",
                    "type": "string",
                    "required": False,
                    "default": op["document"],
                    "description": "GraphQL query document.",
                },
                10,
            )
            emit_param(
                group_lines,
                {
                    "name": "variables",
                    "type": "object",
                    "required": False,
                    "description": "GraphQL variables as JSON.",
                },
                10,
            )
            group_lines.extend(
                [
                    "        response:",
                    "          type: object",
                    f"          item: {op['name']}Response",
                    "",
                ]
            )

    lines.extend(
        [
            "types:",
            "  GraphQLResponse:",
            "    fields:",
            "      - name: data",
            "        type: object",
            "      - name: errors",
            "        type: array",
        ]
    )
    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, default=DEFAULT_SOURCE)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--base-url", default="https://qtz-affine.fly.dev")
    args = parser.parse_args()

    source = args.source
    output = args.output
    schema_path = source / "packages" / "backend" / "server" / "src" / "schema.gql"
    gql_root = source / "packages" / "common" / "graphql" / "src" / "graphql"
    if not schema_path.exists():
        raise SystemExit(f"schema not found: {schema_path}")
    if not gql_root.exists():
        raise SystemExit(f"graphql documents not found: {gql_root}")

    output.mkdir(parents=True, exist_ok=True)
    schema = read_text(schema_path)
    shutil.copyfile(schema_path, output / "schema.gql")

    query_fields = extract_root_fields(schema, "Query")
    mutation_fields = extract_root_fields(schema, "Mutation")

    gql_files = sorted(gql_root.rglob("*.gql"))
    fragment_map: dict[str, str] = {}
    definitions_by_file: dict[Path, list[tuple[str, str, str]]] = {}
    for file_path in gql_files:
        definitions = split_definitions(read_text(file_path))
        definitions_by_file[file_path] = definitions
        for kind, name, document in definitions:
            if kind == "fragment":
                fragment_map[name] = document

    operations: list[dict[str, object]] = []
    fragments: list[dict[str, str]] = []
    for file_path, definitions in definitions_by_file.items():
        rel = file_path.relative_to(gql_root).as_posix()
        for kind, name, document in definitions:
            if kind == "fragment":
                fragments.append({"name": name, "file": rel})
                continue
            full_doc = with_fragments(document, fragment_map)
            root_field = operation_root_field(document)
            operations.append(
                {
                    "name": name,
                    "kind": kind,
                    "file": rel,
                    "root_field": root_field,
                    "group": operation_group(file_path, name, root_field),
                    "variables_declared": re.findall(r"\$([A-Za-z_][A-Za-z0-9_]*)", document.split("{", 1)[0]),
                    "uses_upload": "Upload" in document,
                    "document": full_doc,
                }
            )

    inventory = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "source_repo": str(source),
        "source_git_head": git_head(source),
        "schema": {
            "path": str(schema_path),
            "query_field_count": len(query_fields),
            "mutation_field_count": len(mutation_fields),
            "query_fields": query_fields,
            "mutation_fields": mutation_fields,
        },
        "documents": {
            "root": str(gql_root),
            "file_count": len(gql_files),
            "operation_count": len(operations),
            "query_operation_count": sum(1 for op in operations if op["kind"] == "query"),
            "mutation_operation_count": sum(1 for op in operations if op["kind"] == "mutation"),
            "fragment_count": len(fragments),
            "operations": [
                {k: v for k, v in op.items() if k != "document"}
                for op in sorted(operations, key=lambda item: (str(item["kind"]), str(item["name"])))
            ],
            "fragments": sorted(fragments, key=lambda item: item["name"]),
        },
        "printing_press": {
            "spec": str(output / "affine-graphql.yaml"),
            "base_url": args.base_url,
            "graphql_endpoint_path": "/graphql",
            "auth_env": "AFFINE_TOKEN",
        },
    }
    (output / "affine-api-inventory.json").write_text(
        json.dumps(inventory, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8",
    )
    (output / "affine-graphql.yaml").write_text(
        generate_yaml(operations, args.base_url),
        encoding="utf-8",
    )
    (output / "README.md").write_text(
        "# AFFiNE Printing Press Inputs\n\n"
        "- `schema.gql`: copy of the SDL generated in `qtz-affine`.\n"
        "- `affine-api-inventory.json`: inventory of the schema and client GraphQL documents.\n"
        "- `affine-graphql.yaml`: spec interna consumida pelo Printing Press.\n\n"
        "The spec exposes `graphql execute` for any GraphQL document and named commands for the existing `.gql` operations.\n",
        encoding="utf-8",
    )
    print(json.dumps({"output": str(output), "operations": len(operations), "queries": len(query_fields), "mutations": len(mutation_fields)}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
