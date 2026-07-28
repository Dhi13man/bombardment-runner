#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly ROOT_DIR
readonly APP_DIR="${ROOT_DIR}/app"
readonly DIST_DIR="${ROOT_DIR}/dist"
readonly EXPECTED_GO_VERSION="go1.25.12"
readonly SMOKE_PORT="${SMOKE_PORT:-18080}"
readonly SEMVER_CORE_IDENTIFIER='(0|[1-9][0-9]*)'
readonly SEMVER_PRERELEASE_IDENTIFIER='(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)'
readonly SEMVER_BUILD_IDENTIFIER='[0-9A-Za-z-]+'
readonly VERSION_PATTERN="^v${SEMVER_CORE_IDENTIFIER}\.${SEMVER_CORE_IDENTIFIER}\.${SEMVER_CORE_IDENTIFIER}(-${SEMVER_PRERELEASE_IDENTIFIER}(\.${SEMVER_PRERELEASE_IDENTIFIER})*)?(\+${SEMVER_BUILD_IDENTIFIER}(\.${SEMVER_BUILD_IDENTIFIER})*)?$"
readonly -a PLATFORMS=(
    "linux amd64"
    "linux arm64"
    "darwin amd64"
    "darwin arm64"
    "windows amd64"
    "windows arm64"
)

TEMP_DIR=""
SERVER_PID=""

usage() {
    printf 'Usage: %s <version>\n' "$(basename "$0")"
    printf 'Example: %s v1.0.1\n' "$(basename "$0")"
}

die() {
    printf 'ERROR: %s\n' "$*" >&2
    exit 1
}

require() {
    command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

is_valid_semver() {
    [[ "$1" =~ ${VERSION_PATTERN} ]]
}

cleanup() {
    if [[ -n "${SERVER_PID}" ]] && kill -0 "${SERVER_PID}" 2>/dev/null; then
        kill "${SERVER_PID}" 2>/dev/null || true
        wait "${SERVER_PID}" 2>/dev/null || true
    fi
    if [[ -n "${TEMP_DIR}" ]]; then
        rm -rf -- "${TEMP_DIR}"
    fi
}

runtime_manifest() {
    local binary_name="$1"
    printf '%s\n' "${binary_name}"
    (
        cd "${APP_DIR}"
        find \
            src/app/ui/index.html \
            src/app/ui/static/assets \
            src/app/ui/static/css \
            src/app/ui/static/fonts \
            src/app/ui/static/icons \
            src/app/ui/static/js \
            -type f -print | LC_ALL=C sort
    )
}

archive_manifest() {
    local archive="$1"
    case "${archive}" in
        *.tar.gz) tar -tzf "${archive}" ;;
        *.zip) unzip -Z1 "${archive}" ;;
        *) die "unsupported archive: ${archive}" ;;
    esac | sed 's#^\./##' | grep -v '/$' | LC_ALL=C sort
}

verify_archive() {
    local archive="$1"
    local binary_name="$2"
    local expected_manifest="${TEMP_DIR}/expected-manifest.txt"
    local actual_manifest="${TEMP_DIR}/actual-manifest.txt"

    runtime_manifest "${binary_name}" >"${expected_manifest}"
    archive_manifest "${archive}" >"${actual_manifest}"
    diff -u "${expected_manifest}" "${actual_manifest}" || die "unexpected files in ${archive}"
    if grep -Eq '(^|/)static/src/|(^|/).*\.test\.(js|ts|tsx)$' "${actual_manifest}"; then
        die "development sources found in ${archive}"
    fi
}

build_archive() {
    local goos="$1"
    local goarch="$2"
    local suffix=""
    local archive_extension="tar.gz"
    if [[ "${goos}" == "windows" ]]; then
        suffix=".exe"
        archive_extension="zip"
    fi

    local artifact_name="bombardment-${goos}-${goarch}"
    local binary_name="${artifact_name}${suffix}"
    local stage_dir="${TEMP_DIR}/${artifact_name}"
    local archive="${DIST_DIR}/${artifact_name}.${archive_extension}"
    mkdir -p "${stage_dir}/src/app/ui/static"

    (
        cd "${APP_DIR}"
        CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" go build \
            -buildvcs=false \
            -mod=readonly \
            -trimpath \
            -ldflags="-s -w -X github.dhi13man.com/bombardment-runner/src/app/cli.version=${VERSION}" \
            -o "${stage_dir}/${binary_name}" .
    )
    cp "${APP_DIR}/src/app/ui/index.html" "${stage_dir}/src/app/ui/index.html"
    cp -R \
        "${APP_DIR}/src/app/ui/static/assets" \
        "${APP_DIR}/src/app/ui/static/css" \
        "${APP_DIR}/src/app/ui/static/fonts" \
        "${APP_DIR}/src/app/ui/static/icons" \
        "${APP_DIR}/src/app/ui/static/js" \
        "${stage_dir}/src/app/ui/static/"

    if [[ "${archive_extension}" == "zip" ]]; then
        (
            cd "${stage_dir}"
            zip -q -r "${archive}" "${binary_name}" src/app/ui
        )
    else
        tar -C "${stage_dir}" -czf "${archive}" "${binary_name}" src/app/ui
    fi

    go version -m "${stage_dir}/${binary_name}" | grep -F "${EXPECTED_GO_VERSION}" >/dev/null || \
        die "${binary_name} was not built with ${EXPECTED_GO_VERSION}"
    verify_archive "${archive}" "${binary_name}"
}

wait_for_http_200() {
    local url="$1"
    local status=""
    for _ in {1..30}; do
        status="$(curl -s -o /dev/null -w '%{http_code}' "${url}" || true)"
        if [[ "${status}" == "200" ]]; then
            return
        fi
        if ! kill -0 "${SERVER_PID}" 2>/dev/null; then
            die "release smoke server stopped before ${url} returned 200"
        fi
        sleep 1
    done
    die "${url} returned ${status:-no response}, expected 200"
}

smoke_linux_archive() {
    local smoke_dir="${TEMP_DIR}/smoke"
    local binary="${smoke_dir}/bombardment-linux-amd64"
    mkdir -p "${smoke_dir}"
    tar -xzf "${DIST_DIR}/bombardment-linux-amd64.tar.gz" -C "${smoke_dir}"

    local version_output
    version_output="$("${binary}" --version 2>"${TEMP_DIR}/version.log")"
    [[ "${version_output}" == "bombardment version ${VERSION}" ]] || \
        die "--version returned ${version_output@Q}"

    (
        cd "${smoke_dir}"
        "${binary}" server --bind-addr 127.0.0.1 --port "${SMOKE_PORT}"
    ) >"${TEMP_DIR}/server.log" 2>&1 &
    SERVER_PID=$!

    wait_for_http_200 "http://127.0.0.1:${SMOKE_PORT}/"
    wait_for_http_200 "http://127.0.0.1:${SMOKE_PORT}/v1/ping"
    kill "${SERVER_PID}"
    wait "${SERVER_PID}" || true
    SERVER_PID=""
}

main() {
    if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
        usage
        return
    fi
    [[ $# -eq 1 ]] || { usage >&2; return 1; }
    readonly VERSION="$1"
    is_valid_semver "${VERSION}" || die "version must follow SemVer 2.0.0 and start with v"
    [[ "${SMOKE_PORT}" =~ ^[0-9]+$ ]] || die "SMOKE_PORT must be numeric"
    ((SMOKE_PORT >= 1 && SMOKE_PORT <= 65535)) || die "SMOKE_PORT must be between 1 and 65535"

    for command in curl diff find go grep npm sed sha256sum sort tar unzip zip; do
        require "${command}"
    done
    [[ "$(cd "${APP_DIR}" && go env GOVERSION)" == "${EXPECTED_GO_VERSION}" ]] || \
        die "${EXPECTED_GO_VERSION} is required"
    [[ "${DIST_DIR}" == "${ROOT_DIR}/dist" ]] || die "unsafe dist path"

    TEMP_DIR="$(mktemp -d)"
    trap cleanup EXIT
    rm -rf -- "${DIST_DIR}"
    mkdir -p "${DIST_DIR}"

    npm --prefix "${APP_DIR}" ci
    npm --prefix "${APP_DIR}" run build
    for platform in "${PLATFORMS[@]}"; do
        read -r goos goarch <<<"${platform}"
        build_archive "${goos}" "${goarch}"
    done
    (
        cd "${DIST_DIR}"
        sha256sum bombardment-*.tar.gz bombardment-*.zip >SHA256SUMS
    )
    smoke_linux_archive
    printf 'Verified six release archives for %s in %s\n' "${VERSION}" "${DIST_DIR}"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
