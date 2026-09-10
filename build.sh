#!/usr/bin/env bash

set -euo pipefail

# Usage:
#   ./build.sh                 build and package all desktop targets
#   ./build.sh local           current platform only
#   ./build.sh windows linux   all architectures of those OSes
#   ./build.sh darwin/arm64    one target

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
output_dir="$project_dir/dist"

export PATH="/usr/local/go/bin:/opt/homebrew/bin:$PATH"

if ! command -v go >/dev/null 2>&1; then
  echo "go is required" >&2
  exit 1
fi

cd "$project_dir"
mkdir -p "$output_dir"

host_os="$(go env GOOS)"
host_arch="$(go env GOARCH)"

all_targets=(
  darwin/arm64
  darwin/amd64
  windows/amd64
  windows/arm64
  linux/amd64
  linux/arm64
)

usage() {
  /bin/cat <<'EOF'
Usage: ./build.sh [local | os | os/arch ...]

  ./build.sh                 all desktop platforms
  ./build.sh local           this machine only
  ./build.sh windows         windows/amd64 and windows/arm64
  ./build.sh darwin/arm64    one target

Each success creates dist/mypet-<os>-<arch>/ (binary + assets)
and dist/mypet-<os>-<arch>.zip. The current platform binary is
also copied to dist/mypet.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

select_targets() {
  if [[ $# -eq 0 ]]; then
    printf '%s\n' "${all_targets[@]}"
    return
  fi
  if [[ $# -eq 1 && "$1" == "local" ]]; then
    echo "${host_os}/${host_arch}"
    return
  fi

  local arg os arch target
  for arg in "$@"; do
    case "$arg" in
      */*)
        echo "$arg"
        ;;
      darwin|windows|linux)
        os="$arg"
        for target in "${all_targets[@]}"; do
          if [[ "$target" == "$os/"* ]]; then
            echo "$target"
          fi
        done
        ;;
      *)
        echo "unknown target: $arg" >&2
        usage >&2
        exit 2
        ;;
    esac
  done
}

targets=()
while IFS= read -r line; do
  [[ -n "$line" ]] || continue
  targets+=("$line")
done < <(select_targets "$@")
if [[ ${#targets[@]} -eq 0 ]]; then
  echo "no targets selected" >&2
  exit 2
fi

copy_assets() {
  local dest="$1"
  mkdir -p "$dest"
  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete \
      --exclude '.DS_Store' \
      --exclude '*.zip' \
      "$project_dir/assets/" "$dest/"
    return
  fi
  rm -rf "$dest"
  mkdir -p "$dest"
  cp -R "$project_dir/assets/." "$dest/"
  find "$dest" -name '.DS_Store' -delete
  find "$dest" -name '*.zip' -delete
}

write_package_readme() {
  local dest="$1"
  local goos="$2"
  local bin_name="$3"
  local run_hint="./$bin_name"
  if [[ "$goos" == windows ]]; then
    run_hint="$bin_name"
  fi
  /bin/cat >"$dest/README.txt" <<EOF
MyPet

Run: ${run_hint}
Keep the assets folder next to the executable.

Drag = hold left mouse button
Hit = click or Space
Next pet = right-click, then "下一只"
Quit = Esc
Speed = edit assets/config.txt
Custom pet = put PNGs in assets/pets/<name>/

解压后运行：${run_hint}
请保持 assets 和可执行文件在同一目录。
拖动=按住左键；拍打=点击或空格；换一只=右键「下一只」；退出=Esc。
EOF
}

linux_c_compiler() {
  local arch="$1"
  case "$arch" in
    amd64)
      if command -v x86_64-linux-gnu-gcc >/dev/null 2>&1; then
        echo "x86_64-linux-gnu-gcc"
        return 0
      fi
      if command -v zig >/dev/null 2>&1; then
        echo "zig cc -target x86_64-linux-gnu"
        return 0
      fi
      ;;
    arm64)
      if command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then
        echo "aarch64-linux-gnu-gcc"
        return 0
      fi
      if command -v zig >/dev/null 2>&1; then
        echo "zig cc -target aarch64-linux-gnu"
        return 0
      fi
      ;;
  esac
  return 1
}

package_name() {
  local goos="$1"
  local goarch="$2"
  local ver="${VERSION:-}"
  ver="${ver#v}"
  if [[ -n "$ver" ]]; then
    echo "mypet-${ver}-${goos}-${goarch}"
  else
    echo "mypet-${goos}-${goarch}"
  fi
}

zip_package() {
  local name="$1"
  local zip_path="$output_dir/${name}.zip"
  rm -f "$zip_path"
  # Prefer Python so Windows CI does not need zip(1), and macOS does not
  # emit AppleDouble (._*) files.
  if command -v python3 >/dev/null 2>&1; then
    (cd "$output_dir" && python3 -m zipfile -c "${name}.zip" "$name")
    return
  fi
  if command -v python >/dev/null 2>&1; then
    (cd "$output_dir" && python -m zipfile -c "${name}.zip" "$name")
    return
  fi
  if command -v zip >/dev/null 2>&1; then
    (cd "$output_dir" && COPYFILE_DISABLE=1 zip -qry "${name}.zip" "$name")
    return
  fi
  echo "python or zip is required to package ${name}" >&2
  return 1
}

build_target() {
  local goos="$1"
  local goarch="$2"
  local name
  name="$(package_name "$goos" "$goarch")"
  local ext=""
  if [[ "$goos" == windows ]]; then
    ext=".exe"
  fi
  local pkg_dir="$output_dir/$name"
  local bin_name="mypet${ext}"
  local bin_path="$pkg_dir/$bin_name"

  rm -rf "$pkg_dir"
  mkdir -p "$pkg_dir"

  local cgo=0
  local cc=""
  case "$goos" in
    windows)
      cgo=0
      ;;
    darwin)
      cgo=1
      ;;
    linux)
      cgo=1
      if [[ "$host_os" != linux || "$host_arch" != "$goarch" ]]; then
        if ! cc="$(linux_c_compiler "$goarch")"; then
          echo "skip ${goos}/${goarch}: Linux C compiler not found (install zig or ${goarch} linux gcc)"
          return 2
        fi
      fi
      ;;
    *)
      echo "skip ${goos}/${goarch}: unsupported OS"
      return 2
      ;;
  esac

  local ldflags="-s -w"
  if [[ "$goos" == windows ]]; then
    ldflags="$ldflags -H windowsgui"
  fi

  echo "building ${goos}/${goarch} (cgo=${cgo})..."
  local build_env=(
    GOOS="$goos"
    GOARCH="$goarch"
    CGO_ENABLED="$cgo"
  )
  if [[ -n "$cc" ]]; then
    build_env+=(CC="$cc")
  fi

  if ! env "${build_env[@]}" go build -trimpath -ldflags="$ldflags" -o "$bin_path" .; then
    echo "fail ${goos}/${goarch}"
    rm -rf "$pkg_dir"
    return 1
  fi

  if [[ "$goos" == darwin ]] && command -v codesign >/dev/null 2>&1; then
    codesign --sign - --force "$bin_path" >/dev/null 2>&1 || true
  fi

  copy_assets "$pkg_dir/assets"
  write_package_readme "$pkg_dir" "$goos" "$bin_name"
  zip_package "$name"

  if [[ "$goos" == "$host_os" && "$goarch" == "$host_arch" ]]; then
    cp "$bin_path" "$output_dir/$bin_name"
  fi

  echo "ok ${name} -> ${pkg_dir}  ${output_dir}/${name}.zip"
}

built=()
skipped=()
failed=()

for target in "${targets[@]}"; do
  goos="${target%%/*}"
  goarch="${target##*/}"
  if [[ "$goos" == "$goarch" || -z "$goarch" ]]; then
    echo "invalid target: $target" >&2
    failed+=("$target")
    continue
  fi
  if build_target "$goos" "$goarch"; then
    built+=("$target")
  else
    status=$?
    if [[ "$status" -eq 2 ]]; then
      skipped+=("$target")
    else
      failed+=("$target")
    fi
  fi
done

echo
echo "built:   ${built[*]:-none}"
echo "skipped: ${skipped[*]:-none}"
echo "failed:  ${failed[*]:-none}"

if [[ ${#failed[@]} -gt 0 ]]; then
  exit 1
fi
if [[ ${#built[@]} -eq 0 ]]; then
  echo "nothing was built" >&2
  exit 1
fi
