#!/usr/bin/env bash
set -Eeuo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

image_name="${IMAGE_NAME:-ghcr.io/gummy-tank-attic/sub2api-fxvia}"
git_sha="$(git rev-parse --short=12 HEAD)"
version="${VERSION:-$(git describe --tags --always --dirty)-fxvia.${git_sha}}"

docker build \
  --file Dockerfile \
  --build-arg "VERSION=${version}" \
  --build-arg "COMMIT=${git_sha}" \
  --tag "${image_name}:${git_sha}" \
  --tag "${image_name}:candidate" \
  .

if [[ "${PUSH:-false}" == "true" ]]; then
  docker push "${image_name}:${git_sha}"
  docker push "${image_name}:candidate"
fi
