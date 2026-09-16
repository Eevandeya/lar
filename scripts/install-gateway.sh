#!/bin/bash

set -euo pipefail

BINARY_TMP=
SERVICE_TMP=

GITHUB_REPO=eevandeya/lar
LAR_USER=lar
BINARY_DIR=/usr/local/bin
BINARY_NAME=lar-gateway
SYSTEMD_UNIT_NAME=lar-gateway.service
SYSTEMD_UNIT_PATH=systemd/${SYSTEMD_UNIT_NAME}
CONFIG_DIR=/etc/lar-gateway

NO_SERVICE=false

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
RESET=$'\033[0m'

info() { printf "%s[INFO]%s %s\n" "$GREEN" "$RESET" "$*"; }
warn() { printf "%s[WARN]%s %s\n" "$YELLOW" "$RESET" "$*"; }
error() { printf "%s[ERROR]%s %s\n" "$RED" "$RESET" "$*"; }

usage() {
	cat <<EOF
Usage: $0 [OPTION]
Options:
  --no-service    Install lar-gateway without installing the systemd service.
  --help          Show this help message.
EOF
}

parse_args() {
	while [ "$#" -gt 0 ]; do
		case "$1" in
		--no-service)
			NO_SERVICE=true
			;;
		--help)
			usage
			exit 0
			;;
		*)
			error "unknown option: $1"
			exit 1
			;;
		esac

		shift
	done
}

check_root() {
	if [ "$(id -u)" -ne "0" ]; then
		error "This script must be run as root. Try again with sudo."
		exit 1
	fi
}

check_os() {
	local os
	os=$(uname -s)
	info "os: $os"

	case "$os" in
	Linux)
		;;
	*)
		error "$os is currently unsupported by lar-gateway"
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
	armv7l)
		ARCH="armv7"
		;;
	armv6l)
		ARCH="armv6"
		;;
	*)
		error "$machine architecture is unsupported by lar-gateway"
		exit 1
		;;
	esac

	info "arch: $ARCH"
}

check_dependencies() {
	local command

	for command in "$@"; do
		if ! command -v "$command" >/dev/null 2>&1; then
			error "script requires ${command}"
			exit 1
		fi
	done
}

create_config_dir() {
	info "creating configuration directory"
	mkdir -p "$CONFIG_DIR"

	if [ "$NO_SERVICE" = "false" ]; then
		chown "$LAR_USER:$LAR_USER" "$CONFIG_DIR"
	fi
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
	local asset download_url
	asset="lar-gateway-${VERSION}-linux-${ARCH}.tar.gz"
	download_url="https://github.com/${GITHUB_REPO}/releases/latest/download/${asset}"

	BINARY_TMP=$(mktemp)

	info "downloading lar-gateway binary $VERSION for target architecture..."
	if curl -fsSL "$download_url" -o "$BINARY_TMP"; then
		info "binary download successful"
	else
		error "error downloading lar-gateway binary"
		exit 1
	fi
}

install_binary() {
	local binary_path

	info "installing lar-gateway binary..."
	binary_path="${BINARY_DIR}/${BINARY_NAME}"
	tar -xzf "$BINARY_TMP" -C "$BINARY_DIR"

	if [ ! -f "$binary_path" ]; then
		error "binary was not found after extraction"
		exit 1
	fi

	chmod +x "$binary_path"
	info "lar-gateway binary installed in ${BINARY_DIR}"
}

download_systemd_service() {
	local download_url

	SERVICE_TMP=$(mktemp)
	download_url="https://raw.githubusercontent.com/${GITHUB_REPO}/${VERSION}/${SYSTEMD_UNIT_PATH}"

	info "downloading lar-gateway service unit..."
	if curl -fsSL "$download_url" -o "$SERVICE_TMP"; then
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
}

print_success_message() {
	printf "\n${GREEN}lar-gateway installation completed.${RESET}\n\n"
}

main() {
	trap cleanup EXIT

	parse_args "$@"

	if [ "$NO_SERVICE" = "true" ]; then
		print_header

		printf "Only lar-gateway binary will be installed.\n\n"

		check_root
		check_os
		check_architecture
		check_dependencies curl tar mktemp rm sed

		create_config_dir
		get_latest_version
		download_binary
		install_binary

		print_success_message

		printf "\nOnly the lar-gateway binary was installed.\n"
		printf "For systemd service use this script without --no-service flag.\n\n"
		printf "Before lar-gateway start, create the configuration file at:\n"
		printf "    ${CONFIG_DIR}/config.yml\n\n"
	else
		print_header
		printf "lar-gateway will be installed as a systemd service.\n"
		printf "The service will not be started automatically.\n\n"

		check_root
		check_os
		check_architecture
		check_dependencies curl systemctl tar useradd mktemp rm sed

		create_config_dir
		create_user
		get_latest_version
		download_binary
		install_binary
		download_systemd_service
		install_systemd_service

		print_success_message

		printf "The service has not been started yet.\n"
		printf "It cannot start until its configuration file is created.\n\n"
		printf "Create the configuration file at:\n"
		printf "    ${CONFIG_DIR}/config.yml\n\n"
		printf "Then start the service with:\n"
		printf "    sudo systemctl start ${SYSTEMD_UNIT_NAME}\n\n"
	fi
}

main "$@"
