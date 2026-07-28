#!/usr/bin/env bash
set -euo pipefail

TEST_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly TEST_SCRIPT_DIR
# shellcheck source=scripts/package-release.sh
source "${TEST_SCRIPT_DIR}/package-release.sh"

main() {
    # Arrange
    local -a valid_versions=(
        "v0.0.0"
        "v1.2.3"
        "v1.0.0-alpha"
        "v1.0.0-alpha.1"
        "v1.0.0-0.3.7"
        "v1.0.0-x.7.z.92"
        "v1.0.0+20130313144700"
        "v1.0.0+001"
        "v1.0.0-beta+exp.sha.5114f85"
        "v1.0.0-01a"
    )
    local -a invalid_versions=(
        "1.0.0"
        "v1.0"
        "v01.0.0"
        "v1.01.0"
        "v1.0.01"
        "v1.0.0-"
        "v1.0.0-01"
        "v1.0.0-alpha.01"
        "v1.0.0-alpha..1"
        "v1.0.0+"
        "v1.0.0+build..1"
        "v1.0.0-alpha+build+again"
        "v1.0.0-alpha_1"
    )
    local -a failures=()
    local version

    # Act
    for version in "${valid_versions[@]}"; do
        if ! is_valid_semver "${version}"; then
            failures+=("rejected valid version: ${version}")
        fi
    done
    for version in "${invalid_versions[@]}"; do
        if is_valid_semver "${version}"; then
            failures+=("accepted invalid version: ${version}")
        fi
    done

    local temp_dir
    temp_dir="$(mktemp -d)"
    local sentinel="${temp_dir}/injected"
    local -a injection_versions=(
        "v1.2.3;touch ${sentinel}"
        "v1.2.3\$(touch ${sentinel})"
    )
    for version in "${injection_versions[@]}"; do
        if bash "${TEST_SCRIPT_DIR}/package-release.sh" "${version}" >/dev/null 2>&1; then
            failures+=("accepted shell injection: ${version}")
        fi
    done

    # Assert
    if [[ -e "${sentinel}" ]]; then
        failures+=("executed shell content from a version argument")
    fi
    rm -f -- "${sentinel}"
    rmdir -- "${temp_dir}"
    if (("${#failures[@]}" > 0)); then
        printf 'FAIL: %s\n' "${failures[@]}" >&2
        return 1
    fi
    printf 'Verified %d valid and %d invalid SemVer cases plus shell injection rejection.\n' \
        "${#valid_versions[@]}" "${#invalid_versions[@]}"
}

main "$@"
