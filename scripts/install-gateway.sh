#!/bin/bash

set -eu

BINARY_TMP=
SERVICE_TMP=

GITHUB_REPO=eevandeya/lar
LAR_USER=lar
BINARY_DIR=/usr/local/bin
BINARY_NAME=lar-gateway
SYSTEMD_UNIT_NAME=lar-gateway.service
SYSTEMD_UNIT_PATH=systemd/${SYSTEMD_UNIT_NAME}

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
RESET=$'\033[0m'

info() { printf "%s[INFO]%s %s\n" "$GREEN" "$RESET" "$*"; }
warn() { printf "%s[WARN]%s %s\n" "$YELLOW" "$RESET" "$*"; }
error() { printf "%s[ERROR]%s %s\n" "$RED" "$RESET" "$*"; }

check_root() {
	if [ "$(id -u)" -ne "0" ]; then
		error "This script must be run as root. Try again with sudo."
		exit 1
	fi
}

check_os() {
	OS=$(uname -s)
	info "os: $OS"

	case "$OS" in
	"Linux")
		;;
	*)
		error "$OS is currently unsupported by lar-gateway"
		exit 1
		;;
	esac
}

check_architecture() {
	MACHINE=$(uname -m)

	case "$MACHINE" in
	"x86_64" | "amd64")
		ARCH="amd64"
		;;
	"aarch64" | "arm64")
		ARCH="arm64"
		;;
	"armv7l")
		ARCH="armv7"
		;;
	"armv6l")
		ARCH="armv6"
		;;
	*)
		error "$MACHINE architecture is unsupported by lar-gateway"
		exit 1
		;;
	esac

	info "arch: $ARCH"
}

check_dependencies() {
	for command in curl systemctl tar useradd mktemp rm sed; do
		if ! command -v "$command" >/dev/null 2>&1; then
			error "script requires ${command}"
			exit 1
		fi
	done
}

create_user() {
	if id "$LAR_USER" >/dev/null 2>&1; then
		info "user ${LAR_USER} already exists, using it for lar-gateway"
	else
		info "creating user $LAR_USER"
		useradd --system --shell /usr/sbin/nologin "$LAR_USER"
	fi
}

get_latest_version() {
	info "getting latest lar-gateway version"
	VERSION=$(curl -fsSL \
		"https://api.github.com/repos/${GITHUB_REPO}/releases/latest" |
		sed -n 's/.*"tag_name": "\([^"]*\)".*/\1/p')

	if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
		error "failed to determine latest version"
		exit 1
	fi

	info "latest version for lar-gateway is $VERSION"
}

download_binary() {
	BINARY_TMP=$(mktemp)
	ASSET="lar-gateway-${VERSION}-linux-${ARCH}.tar.gz"
	DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${ASSET}"

	info "downloading lar-gateway binary $VERSION for target architecture..."
	if curl -fsSL "$DOWNLOAD_URL" -o "$BINARY_TMP"; then
		info "binary download successful"
	else
		error "error downloading lar-gateway binary"
		exit 1
	fi
}

install_binary() {
	info "installing lar-gateway binary..."
	BINARY_PATH="${BINARY_DIR}/${BINARY_NAME}"
	tar -xzf "$BINARY_TMP" -C "$BINARY_DIR"

	if [ ! -f "$BINARY_PATH" ]; then
		error "binary was not found after extraction"
		exit 1
	fi

	chmod +x "$BINARY_PATH"
	info "lar-gateway binary installed in ${BINARY_DIR}"
}

download_systemd_service() {
	SERVICE_TMP=$(mktemp)
	DOWNLOAD_URL=https://raw.githubusercontent.com/${GITHUB_REPO}/${VERSION}/${SYSTEMD_UNIT_PATH}

	info "downloading lar-gateway service unit..."
	if curl -fsSL "$DOWNLOAD_URL" -o "$SERVICE_TMP"; then
		info "systemd service unit download successful"
	else
		error "error downloading lar-gateway systemd service unit"
		exit 1
	fi
}

install_systemd_service() {
	mv "$SERVICE_TMP" "/etc/systemd/system/${SYSTEMD_UNIT_NAME}"
	info "reloading systemd configuration..."
	systemctl daemon-reload
	info "systemd configuration reloaded successfully"
}

cleanup() {
	[ -z "$BINARY_TMP" ] || rm -f "$BINARY_TMP"
	[ -z "$SERVICE_TMP" ] || rm -f "$SERVICE_TMP"
}

print_header() {
	printf "lar-gateway installer\n"
	printf "=====================\n\n"
	printf "This script installs lar-gateway as a systemd service.\n"
	printf "The service will not be started automatically.\n\n"
}

print_success_message() {
	printf "\n${GREEN}lar-gateway installation completed.${RESET}\n\n"
	printf "The service has not been started yet.\n"
	printf "It cannot start until its configuration file is created.\n\n"
	printf "Create the configuration file at:\n"
	printf "    /etc/lar-gateway/config.yml\n\n"
	printf "Then start the service with:\n"
	printf "    sudo systemctl start ${SYSTEMD_UNIT_NAME}\n\n"
}

main() {
	trap cleanup EXIT

	print_header
	check_root
	check_os
	check_architecture
	check_dependencies
	create_user
	get_latest_version
	download_binary
	install_binary
	download_systemd_service
	install_systemd_service
	print_success_message
}

main "$@"
