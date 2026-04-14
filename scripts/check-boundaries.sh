#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"

module_path="github.com/pthethanh/nano"
cache_dir="${TMPDIR:-/tmp}/nano-go-build"
mkdir -p "$cache_dir"
tracked_roots=(
  "broker"
  "cache"
  "config"
  "grpc"
  "log"
  "metric"
  "status"
  "validator"
)

failures=0

contains_root() {
  local value="$1"
  shift
  local item
  for item in "$@"; do
    if [[ "$item" == "$value" ]]; then
      return 0
    fi
  done
  return 1
}

check_root_module_boundaries() {
  while IFS=$'\t' read -r import_path imports; do
    [[ -z "$import_path" ]] && continue

    rel="${import_path#"$module_path"/}"
    root="${rel%%/*}"
    if ! contains_root "$root" "${tracked_roots[@]}"; then
      continue
    fi

    for dep in $imports; do
      if [[ "$dep" != "$module_path/"* ]]; then
        continue
      fi

      dep_rel="${dep#"$module_path"/}"
      dep_root="${dep_rel%%/*}"
      if [[ "$dep_root" == "$root" ]]; then
        continue
      fi
      if ! contains_root "$dep_root" "${tracked_roots[@]}"; then
        continue
      fi

      printf 'boundary violation: %s imports %s\n' "$import_path" "$dep"
      failures=1
    done
  done < <(GOCACHE="$cache_dir" go list -f '{{.ImportPath}}{{"\t"}}{{join .Imports " "}}' ./...)
}

check_workspace_modules() {
  local mod_file="$1"
  local mod_dir
  mod_dir="$(dirname "$mod_file")"

  case "$mod_file" in
    "$repo_root/plugins/"*|"$repo_root/examples/"*|"$repo_root/cmd/"*)
      ;;
    *)
      return
      ;;
  esac

  if grep -Eq '^replace github\.com/pthethanh/nano( |$)' "$mod_file"; then
    printf 'workspace wiring violation: %s should not declare a local replace for github.com/pthethanh/nano; rely on go.work\n' "${mod_file#"$repo_root/"}"
    failures=1
  fi

  if ! (
    cd "$mod_dir"
    GOCACHE="$cache_dir" go list ./... >/dev/null
  ); then
    printf 'workspace wiring violation: %s does not resolve in workspace mode\n' "${mod_file#"$repo_root/"}"
    failures=1
  fi
}

check_root_module_boundaries

while IFS= read -r mod_file; do
  check_workspace_modules "$mod_file"
done < <(find "$repo_root" -name go.mod | sort)

if [[ "$failures" -ne 0 ]]; then
  exit 1
fi

echo "boundary check passed"
