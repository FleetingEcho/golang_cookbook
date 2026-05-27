#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DB_PATH="$ROOT_DIR/data/issue_tracker.db"
UPLOAD_DIR="$ROOT_DIR/uploads"

echo "Recreating database..."
rm -f "$DB_PATH"
mkdir -p "$ROOT_DIR/data" "$UPLOAD_DIR"

sqlite3 "$DB_PATH" < "$ROOT_DIR/migrations/0001_init.sql"
sqlite3 "$DB_PATH" < "$ROOT_DIR/scripts/seed.sql"

find "$UPLOAD_DIR" -type f ! -name ".gitkeep" -delete

echo "Seeded $DB_PATH"
echo "Cleaned $UPLOAD_DIR"
