#!/usr/bin/env bash
set -euo pipefail

for i in $(find -name '*_test.go'); do
    dirname $i
done | sort | uniq
