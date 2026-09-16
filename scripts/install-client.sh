#!/bin/bash

set -euo pipefail

BINARY_TMP=

GITHUB_REPO=eevandeya/lar
BINARY_DIR="$HOME/.local/bin"
BINARY_NAME=lar

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
RESET=$'\033[0m'

info() { printf "%s[INFO]%s %s\n" "$GREEN" "$RESET" "$*"; }
warn() { printf "%s[WARN]%s %s\n" "$YELLOW" "$RESET" "$*"; }
error() { printf "%s[ERROR]%s %s\n" "$RED" "$RESET" "$*"; }

check_os() {
	local os_name
	os_name=$(uname -s)
	info "os: $os_name"

	case "$os_name" in
	Linux)
		OS=linux
		;;
	Darwin)
		OS=darwin
		;;
	*)
		error "$os_name is currently unsupported by lar"
		exit 1
		;;
	esac
}

check_architecture() {
	local machine
	machine=$(uname -m)

	case "$machine" in
	x86_64 | amd64)
		ARCH="amd64"
		;;
	aarch64 | arm64)
		ARCH="arm64"
		;;
	*)
		error "$machine architecture is unsupported by lar"
		exit 1
		;;
	esac

	info "arch: $ARCH"
}

check_dependencies() {
	local command

	for command in curl tar mktemp rm sed mkdir; do
		if ! command -v "$command" >/dev/null 2>&1; then
			error "script requires ${command}"
			exit 1
		fi
	done
}

get_latest_version() {
	info "getting latest lar version"
	VERSION=$(curl -fsSL \
		"https://api.github.com/repos/${GITHUB_REPO}/releases/latest" |
		sed -n 's/.*"tag_name": "\([^"]*\)".*/\1/p')

	if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
		error "failed to determine latest version"
		exit 1
	fi

	info "latest version for lar is $VERSION"
}

download_binary() {
	local asset download_url

	asset="lar-${VERSION}-${OS}-${ARCH}.tar.gz"
	download_url="https://github.com/${GITHUB_REPO}/releases/latest/download/${asset}"

	BINARY_TMP=$(mktemp)

	info "downloading lar binary $VERSION for target architecture..."
	if curl -fsSL "$download_url" -o "$BINARY_TMP"; then
		info "binary download successful"
	else
		error "error downloading lar binary"
		exit 1
	fi
}

install_binary() {
	local binary_path

	info "installing lar binary..."
	binary_path="${BINARY_DIR}/${BINARY_NAME}"
	mkdir -p "$BINARY_DIR"
	tar -xzf "$BINARY_TMP" -C "$BINARY_DIR"

	if [ ! -f "$binary_path" ]; then
		error "binary was not found after extraction"
		exit 1
	fi

	chmod +x "$binary_path"
	info "lar binary installed in ${BINARY_DIR}"
}

cleanup() {
	[ -z "$BINARY_TMP" ] || rm -f "$BINARY_TMP"
}

print_header() {
	printf "lar installer\n"
	printf "=============\n\n"
	printf "This script installs the lar command-line client.\n"
	printf "The client binary will be installed to ~/.local/bin/lar.\n\n"
}

print_success_message() {
	printf "\n${GREEN}lar installation completed.${RESET}\n\n"
	printf "The lar binary was installed to:\n"
	printf "    ${BINARY_DIR}/${BINARY_NAME}\n\n"
	printf "Make sure ~/.local/bin is included in your PATH.\n"
	printf "Verify the installation with:\n"
	printf "    lar --version\n\n"
	printf "Before using lar, create the client configuration file at:\n"
	printf "    ~/.config/lar/config.yml\n\n"
}

main() {
	trap cleanup EXIT

	print_header
	check_os
	check_architecture
	check_dependencies
	get_latest_version
	download_binary
	install_binary
	print_success_message
}

main "$@"
