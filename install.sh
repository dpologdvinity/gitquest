#!/bin/sh
set -eu

die() {
    printf 'gitquest installer: %s\n' "$*" >&2
    exit 1
}

curl_https() {
    curl --proto '=https' --proto-redir '=https' --tlsv1.2 \
        --fail --silent --show-error --location --retry 3 "$@"
}

download() {
    curl_https --output "$2" "$1" || die "download failed: $1"
}

cleanup() {
    if [ -n "$staged" ]; then
        rm -f "$staged"
    fi
    rm -rf "$work_dir"
}

main() {
    case "$(uname -s)" in
        Linux) system=linux ;;
        Darwin) system=darwin ;;
        *) die 'unsupported operating system; use Linux or macOS' ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64) arch=amd64 ;;
        aarch64|arm64) arch=arm64 ;;
        *) die 'unsupported architecture; use amd64 or arm64' ;;
    esac

    for tool in curl tar awk mktemp install git; do
        command -v "$tool" >/dev/null 2>&1 || die "required command not found: $tool"
    done
    git version | awk '{ split($3, v, "."); exit !(v[1] > 2 || (v[1] == 2 && v[2] >= 32)) }' \
        || die 'Git 2.32 or newer is required'

    base=https://github.com/dpologdvinity/gitquest/releases
    version=${GITQUEST_VERSION:-}
    if [ -z "$version" ]; then
        release_url=$(curl_https --output /dev/null --write-out '%{url_effective}' "$base/latest") \
            || die 'could not find the latest release'
        version=${release_url##*/}
    fi
    case "$version" in
        v*) ;;
        *) version=v$version ;;
    esac
    case "$version" in
        *[!0-9A-Za-z.+-]*) die 'invalid version' ;;
        v[0-9]*) ;;
        *) die 'invalid version' ;;
    esac

    install_dir=${GITQUEST_INSTALL_DIR:-${HOME:?HOME must be set}/.local/bin}
    case "$install_dir" in
        /*) ;;
        *) die 'GITQUEST_INSTALL_DIR must be an absolute path' ;;
    esac
    [ ! -d "$install_dir/gitquest" ] || die 'installation target is a directory'

    work_dir=$(mktemp -d "${TMPDIR:-/tmp}/gitquest-install.XXXXXX")
    staged=
    trap cleanup 0
    trap 'exit 1' 1 2 15

    asset=gitquest_${version#v}_${system}_${arch}.tar.gz
    download "$base/download/$version/$asset" "$work_dir/$asset"
    download "$base/download/$version/checksums.txt" "$work_dir/checksums.txt"
    expected=$(awk -v name="$asset" '
        $2 == name && length($1) == 64 && $1 !~ /[^0-9a-f]/ { print $1; count++ }
        END { if (count != 1) exit 1 }
    ' "$work_dir/checksums.txt") || die "missing checksum for $asset"
    if command -v sha256sum >/dev/null 2>&1; then
        actual=$(sha256sum "$work_dir/$asset")
    elif command -v shasum >/dev/null 2>&1; then
        actual=$(shasum -a 256 "$work_dir/$asset")
    else
        die 'sha256sum or shasum is required'
    fi
    [ "${actual%% *}" = "$expected" ] || die "checksum mismatch for $asset"

    tar -xzOf "$work_dir/$asset" gitquest > "$work_dir/gitquest" \
        || die 'archive extraction failed'
    [ -s "$work_dir/gitquest" ] || die 'archive contains no binary'

    mkdir -p "$install_dir"
    staged=$(mktemp "$install_dir/.gitquest.XXXXXX")
    install -m 0755 "$work_dir/gitquest" "$staged"
    mv -f "$staged" "$install_dir/gitquest"
    staged=

    printf 'Installed gitquest %s to %s/gitquest\n' "$version" "$install_dir"
    case ":$PATH:" in
        *":$install_dir:"*) ;;
        *) printf 'Add %s to your PATH to run gitquest.\n' "$install_dir" ;;
    esac
}

main "$@"
