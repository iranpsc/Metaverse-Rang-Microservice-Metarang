#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"

FORCE_OVERWRITE=false
DRY_RUN=false

usage() {
  cat <<'EOF'
Usage: ./scripts/generate-configs.sh [options]

Generate service config files from config.env.sample entries referenced in docker-compose.yml.

Options:
  -f, --force      Overwrite existing config.env files
  -n, --dry-run    Print actions without creating files
  -h, --help       Show this help message
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -f|--force)
      FORCE_OVERWRITE=true
      shift
      ;;
    -n|--dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ ! -f "${COMPOSE_FILE}" ]]; then
  echo "docker-compose.yml not found at ${COMPOSE_FILE}" >&2
  exit 1
fi

env_files=()
while IFS= read -r line; do
  [[ -n "${line}" ]] && env_files+=("${line}")
done < <(
  awk '
    $1 == "-" && $2 ~ /^\.\/.*\/config\.env$/ {
      path = $2
      sub(/^\.\//, "", path)
      print path
    }
  ' "${COMPOSE_FILE}" | sort -u
)

if [[ ${#env_files[@]} -eq 0 ]]; then
  echo "No env_file config.env entries found in docker-compose.yml."
  exit 0
fi

created=0
overwritten=0
skipped=0
missing_samples=0

for rel_env_path in "${env_files[@]}"; do
  target_path="${ROOT_DIR}/${rel_env_path}"
  sample_path="${target_path}.sample"
  target_exists=false
  [[ -f "${target_path}" ]] && target_exists=true

  if [[ ! -f "${sample_path}" ]]; then
    echo "Missing sample: ${sample_path}"
    ((missing_samples+=1))
    continue
  fi

  if [[ "${target_exists}" == "true" && "${FORCE_OVERWRITE}" != "true" ]]; then
    echo "Skip existing: ${target_path}"
    ((skipped+=1))
    continue
  fi

  if [[ "${DRY_RUN}" == "true" ]]; then
    if [[ "${target_exists}" == "true" ]]; then
      echo "Would overwrite: ${target_path}"
    else
      echo "Would create: ${target_path}"
    fi
    continue
  fi

  cp "${sample_path}" "${target_path}"
  if [[ "${target_exists}" == "true" ]]; then
    echo "Overwritten: ${target_path}"
    ((overwritten+=1))
  else
    echo "Created: ${target_path}"
    ((created+=1))
  fi
done

if [[ "${DRY_RUN}" == "true" ]]; then
  echo "Dry run complete."
  exit 0
fi

echo ""
echo "Config generation complete."
echo "Created: ${created}"
echo "Overwritten: ${overwritten}"
echo "Skipped existing: ${skipped}"
echo "Missing samples: ${missing_samples}"

if [[ ${missing_samples} -gt 0 ]]; then
  exit 1
fi
