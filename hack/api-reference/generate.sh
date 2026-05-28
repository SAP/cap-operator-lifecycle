#!/bin/bash
# Generates includes/api-reference.html from the Go types on the given branch (default: main).
# Usage: ./generate.sh [branch]
# Run this from the repo root on the website branch.

set -eo pipefail

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")
REPO_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)
BRANCH=${1:-main}
TEMPDIR=$(mktemp -d)
trap 'rm -rf "$TEMPDIR"' EXIT

echo "Extracting Go types from $BRANCH branch..."
git archive "$BRANCH" -- api/ go.mod go.sum | tar -x -C "$TEMPDIR"

echo "Generating API reference..."
cd "$TEMPDIR"
"$REPO_ROOT/bin/gen-crd-api-reference-docs" \
  -config "$REPO_ROOT/hack/api-reference/config.json" \
  -template-dir "$REPO_ROOT/hack/api-reference/template" \
  -api-dir github.com/sap/cap-operator-lifecycle/api/v1alpha1 \
  -out-file "$REPO_ROOT/includes/api-reference.html"

echo "Done: $REPO_ROOT/includes/api-reference.html"
