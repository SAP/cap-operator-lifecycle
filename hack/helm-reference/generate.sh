#!/bin/bash

set -eo pipefail

# Change dir to the root dir of this repo (should point to cap-operator)
cd $(dirname "${BASH_SOURCE[0]}")/../..

echo "PWD: ${PWD}"

TEMPDIR=$(mktemp -d)
trap 'rm -rf "$TEMPDIR"' EXIT

cp -r "$PWD"/chart "$TEMPDIR"

helm-docs -c "$TEMPDIR"/chart -s file
cp "$TEMPDIR"/chart/README.md "$PWD"/chart