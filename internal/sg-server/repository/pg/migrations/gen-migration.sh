#!/usr/bin/env bash
#
# gen-migration.sh — Assembles a single goose SQL migration file from
#                     multiple SQL fragment files located in subdirectories.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── defaults ────────────────────────────────────────────────────────
OUTPUT_DIR="$SCRIPT_DIR"
SOURCE_DIRS=()    # filled later; empty ⇒ scan every subdir in SCRIPT_DIR

# ── usage ───────────────────────────────────────────────────────────
usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Assemble goose SQL migrations from fragment files.

Options:
  -d DIR   Process only this directory (may be repeated).
           If omitted, every immediate subdirectory of the script's
           directory is processed.
  -o DIR   Write the resulting .sql file(s) into DIR.
           Default: the directory where this script lives.
  -h       Show this help and exit.
EOF
  exit 0
}

# ── parse flags ─────────────────────────────────────────────────────
while getopts ":d:o:h" opt; do
  case "$opt" in
    d) SOURCE_DIRS+=("$OPTARG") ;;
    o) OUTPUT_DIR="$OPTARG" ;;
    h) usage ;;
    :) echo "Error: -$OPTARG requires an argument" >&2; exit 1 ;;
    *) echo "Error: unknown option -$OPTARG" >&2; exit 1 ;;
  esac
done
shift $((OPTIND - 1))

# If no -d given, discover all immediate subdirectories.
if [[ ${#SOURCE_DIRS[@]} -eq 0 ]]; then
  for d in "$SCRIPT_DIR"/*/; do
    [[ -d "$d" ]] && SOURCE_DIRS+=("$d")
  done
fi

if [[ ${#SOURCE_DIRS[@]} -eq 0 ]]; then
  echo "No subdirectories found to process." >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"

# ── helpers ─────────────────────────────────────────────────────────

# extract_section FILE SECTION
#   Prints the content between "-- +goose <SECTION>" / "-- +goose StatementBegin"
#   and "-- +goose StatementEnd" that belongs to the given SECTION (Up or Down).
#   Handles the pattern:
#     -- +goose Up
#     -- +goose StatementBegin
#     <content>
#     -- +goose StatementEnd
extract_section() {
  local file="$1"
  local section="$2"        # "Up" or "Down"
  local in_section=0
  local in_statement=0
  local content=""

  while IFS= read -r line || [[ -n "$line" ]]; do
    # Detect section start
    if [[ "$line" =~ ^--\ \+goose\ Up[[:space:]]*$ ]] && [[ "$section" == "Up" ]]; then
      in_section=1
      in_statement=0
      continue
    fi
    if [[ "$line" =~ ^--\ \+goose\ Down[[:space:]]*$ ]]; then
      if [[ "$section" == "Down" ]]; then
        in_section=1
        in_statement=0
      else
        # We were reading Up — stop now.
        in_section=0
        in_statement=0
      fi
      continue
    fi

    # Inside the desired section
    if [[ $in_section -eq 1 ]]; then
      if [[ "$line" =~ ^--\ \+goose\ StatementBegin[[:space:]]*$ ]]; then
        in_statement=1
        continue
      fi
      if [[ "$line" =~ ^--\ \+goose\ StatementEnd[[:space:]]*$ ]]; then
        in_statement=0
        continue
      fi
      if [[ $in_statement -eq 1 ]]; then
        content+="$line"$'\n'
      fi
    fi
  done < "$file"

  # Trim trailing blank lines but keep one trailing newline
  printf '%s' "$content"
}

# ── main loop ───────────────────────────────────────────────────────
for src_dir in "${SOURCE_DIRS[@]}"; do
  src_dir="${src_dir%/}"              # strip trailing slash
  dir_name="$(basename "$src_dir")"
  out_file="$OUTPUT_DIR/${dir_name}.sql"

  # Collect .sql files sorted by numeric prefix
  mapfile -t sql_files < <(
    find "$src_dir" -maxdepth 1 -name '*.sql' -type f | sort -t/ -k+999 -V
  )

  if [[ ${#sql_files[@]} -eq 0 ]]; then
    echo "Warning: no .sql files in $src_dir — skipping." >&2
    continue
  fi

  up_body=""
  down_body=""

  # Collect all Up sections in file order
  for f in "${sql_files[@]}"; do
    section="$(extract_section "$f" "Up")"
    if [[ -n "$section" ]]; then
      up_body+="$section"$'\n'
    fi
  done

  # Collect all Down sections in REVERSE file order
  # (the conventional approach: undo in the opposite order)
  for (( i=${#sql_files[@]}-1; i>=0; i-- )); do
    section="$(extract_section "${sql_files[$i]}" "Down")"
    if [[ -n "$section" ]]; then
      down_body+="$section"$'\n'
    fi
  done

  # Strip excessive trailing newlines (keep at most one)
  up_body="$(printf '%s' "$up_body" | sed -e :a -e '/^[[:space:]]*$/{ $d; N; ba; }')"
  down_body="$(printf '%s' "$down_body" | sed -e :a -e '/^[[:space:]]*$/{ $d; N; ba; }')"

  # ── write resulting migration ──────────────────────────────────
  {
    echo "-- +goose Up"
    echo "-- +goose StatementBegin"
    if [[ -n "$up_body" ]]; then
      echo ""
      echo "$up_body"
      echo ""
    fi
    echo "-- +goose StatementEnd"
    echo ""
    echo "-- +goose Down"
    echo "-- +goose StatementBegin"
    if [[ -n "$down_body" ]]; then
      echo ""
      echo "$down_body"
      echo ""
    fi
    echo "-- +goose StatementEnd"
  } > "$out_file"

  echo "Generated: $out_file"
done
