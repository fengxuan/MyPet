#!/usr/bin/env bash

set -euo pipefail

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
output_dir="$project_dir/dist"
output_file="$output_dir/mypet"

cd "$project_dir"
mkdir -p "$output_dir"

go build -trimpath -o "$output_file" .

echo "Build complete: $output_file"
