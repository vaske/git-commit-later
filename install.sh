#!/bin/sh
# Install git-commit-later from GitHub Releases. Does not require Go.
set -eu

REPO="vaske/git-commit-later"
BINARY="git-commit-later"

normalize_os() {
	os=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')
	case "$os" in
	linux) echo linux ;;
	darwin) echo darwin ;;
	mingw* | msys* | cygwin*) echo windows ;;
	*)
		echo "unsupported os: $1" >&2
		return 1
		;;
	esac
}

normalize_arch() {
	arch=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')
	case "$arch" in
	x86_64 | amd64) echo amd64 ;;
	aarch64 | arm64) echo arm64 ;;
	*)
		echo "unsupported arch: $1" >&2
		return 1
		;;
	esac
}

archive_name() {
	os=$1
	arch=$2
	if [ "$os" = windows ]; then
		echo "${BINARY}_${os}_${arch}.zip"
	else
		echo "${BINARY}_${os}_${arch}.tar.gz"
	fi
}

release_asset_url() {
	version=$1
	name=$2
	if [ -z "$version" ] || [ "$version" = latest ]; then
		echo "https://github.com/${REPO}/releases/latest/download/${name}"
	else
		echo "https://github.com/${REPO}/releases/download/${version}/${name}"
	fi
}

release_tag() {
	version=${VERSION:-latest}
	if [ "$version" = latest ]; then
		echo latest
		return
	fi
	case "$version" in
	v*) echo "$version" ;;
	*) echo "v${version}" ;;
	esac
}

resolve_prefix() {
	if [ -n "${PREFIX:-}" ]; then
		echo "$PREFIX"
		return
	fi
	if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
		echo /usr/local/bin
		return
	fi
	echo "${HOME}/.local/bin"
}

download() {
	url=$1
	dest=$2
	if ! command -v curl >/dev/null 2>&1; then
		echo "error: curl is required" >&2
		return 1
	fi
	curl -fsSL "$url" -o "$dest"
}

file_sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
	else
		echo "error: need sha256sum or shasum to verify the download" >&2
		return 1
	fi
}

verify_checksum() {
	archive=$1
	sums=$2
	base=$(basename "$archive")
	expected=$(awk -v f="$base" '$2 == f { print $1; exit }' "$sums")
	if [ -z "$expected" ]; then
		echo "error: no checksum for $base in checksums.txt" >&2
		return 1
	fi
	actual=$(file_sha256 "$archive")
	if [ "$expected" != "$actual" ]; then
		echo "error: checksum mismatch for $base" >&2
		return 1
	fi
}

extract_archive() {
	archive=$1
	dest=$2
	case "$archive" in
	*.zip)
		if ! command -v unzip >/dev/null 2>&1; then
			echo "error: unzip is required for Windows archives" >&2
			return 1
		fi
		unzip -o -q "$archive" -d "$dest"
		;;
	*)
		tar -xzf "$archive" -C "$dest"
		;;
	esac
}

main() {
	os=$(normalize_os "$(uname -s)")
	arch=$(normalize_arch "$(uname -m)")
	tag=$(release_tag)
	name=$(archive_name "$os" "$arch")
	prefix=$(resolve_prefix)
	tmpdir=$(mktemp -d)
	trap 'rm -rf "$tmpdir"' EXIT

	echo "Downloading ${name} (${tag})..."
	download "$(release_asset_url "$tag" "$name")" "${tmpdir}/${name}"
	download "$(release_asset_url "$tag" checksums.txt)" "${tmpdir}/checksums.txt"
	verify_checksum "${tmpdir}/${name}" "${tmpdir}/checksums.txt"
	extract_archive "${tmpdir}/${name}" "$tmpdir"

	bin="${tmpdir}/${BINARY}"
	if [ "$os" = windows ]; then
		bin="${tmpdir}/${BINARY}.exe"
	fi
	if [ ! -f "$bin" ]; then
		echo "error: archive did not contain ${BINARY}" >&2
		return 1
	fi
	chmod +x "$bin"
	mkdir -p "$prefix"
	cp "$bin" "${prefix}/$(basename "$bin")"
	echo "Installed ${prefix}/$(basename "$bin")"
	echo "Git will expose it as: git commit-later"
}

if [ "${GIT_COMMIT_LATER_LIB:-}" != "1" ]; then
	main "$@"
fi
