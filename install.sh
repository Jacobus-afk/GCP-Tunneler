#!/usr/bin/env bash

INSTALL_DIR=""
TEMP_DIR=/tmp/gcp-tunneler
trap "rm -rf $TEMP_DIR" EXIT
CONFIG_DIR=$HOME/.config/gcp-tunneler

SUPPORTED_TARGETS="linux-amd64"

error() {
  printf "🛑 \e[31m$1\e[0m\n"
  exit 1
}

info() {
  printf "\e[32m$1\e[0m\n"
}

warn() {
  printf "⚠️ \e[33m$1\e[0m\n"
}

help() {
  echo "GCP Tunneler - Install script"
  echo
  echo "Synopsis: install.sh [-h] [-d <dir>] [-t <dir>]"
  echo "options:"
  echo "-h     Print this help."
  echo "-d     Specify the installation directory. Defaults to $HOME/bin or $HOME/.local/bin"
  # echo "-c     Specify the config directory. Defaults to $HOME/.config/gcp-tunneler"
  echo
  echo $INSTALL_DIR
}

while getopts ":hd:t:v:" option; do
  case $option in
    h) # display Help
      help
      exit;;
    d) # Enter a name
      INSTALL_DIR=${OPTARG};;
    *) # Invalid option
      warn "invalid option"
      help
      exit;;
  esac
done

validate_dependency() {
  if ! command -v $1 >/dev/null; then
    error "$1 is required by GCP Tunneler. Please install\n"
  fi
}

validate_dependencies() {
  validate_dependency tmux
  validate_dependency fzf
  validate_dependency jq
  validate_dependency curl
}

set_install_directory() {
  if [ -n "$INSTALL_DIR" ]; then
    # handle ~
    INSTALL_DIR="${INSTALL_DIR/#\~/$HOME}"
    return 0
  fi

  # check if $HOME/bin exists and is writable
  if [ -d "$HOME/bin" ] && [ -w "$HOME/bin" ]; then
    INSTALL_DIR="$HOME/bin"
    return 0
  fi

  # check if $HOME/.local/bin exists and is writable
  if ([ -d "$HOME/.local/bin" ] && [ -w "$HOME/.local/bin" ]) || mkdir -p "$HOME/.local/bin"; then
    INSTALL_DIR="$HOME/.local/bin"
    return 0
  fi

  error "Cannot determine installation directory. Please specify a directory and try again"
}

validate_install_directory() {
  #check if installation dir exists
  if [ ! -d "$INSTALL_DIR" ]; then
    error "Directory ${INSTALL_DIR} does not exist, set a different directory and try again."
  fi

  # Check if regular user has write permission
  if [ ! -w "$INSTALL_DIR" ]; then
    error "Cannot write to ${INSTALL_DIR}. Please check write permissions or set a different directory and try again"
  fi

  # check if the directory is in the PATH
  good=$(
    IFS=:
    for path in $PATH; do
    if [ "${path%/}" = "${INSTALL_DIR}" ]; then
      printf 1
      break
    fi
    done
  )

  if [ "${good}" != "1" ]; then
    warn "Installation directory ${INSTALL_DIR} is not in your \$PATH, add it using \nexport PATH=\$PATH:${INSTALL_DIR}"
  fi
}

validate_config_directory() {

    # Validate if the config directory exists
    if ! mkdir -p "$CONFIG_DIR" > /dev/null 2>&1; then
        error "Cannot write to ${CONFIG_DIR}. Please check write permissions or set a different directory and try again"
    fi

    #check user write permission
    if [ ! -w "$CONFIG_DIR" ]; then
        error "Cannot write to ${CONFIG_DIR}. Please check write permissions or set a different directory and try again"
    fi
}

validate_temp_directory() {

    # Validate if the tmp directory exists
    if ! mkdir -p "$TEMP_DIR" > /dev/null 2>&1; then
        error "Cannot write to ${TEMP_DIR}. Please check write permissions or set a different directory and try again"
    fi

    #check user write permission
    if [ ! -w "$TEMP_DIR" ]; then
        error "Cannot write to ${TEMP_DIR}. Please check write permissions or set a different directory and try again"
    fi
}


install_scripts() {

  validate_config_directory

  info "🚧 Installing GCP Tunneler scripts in ${CONFIG_DIR}\n"

  cp -a /tmp/gcp-tunneler/scripts $CONFIG_DIR
  cp -a /tmp/gcp-tunneler/config.toml.example

  if [ $? -ne 0 ]; then
    error "Unable to copy scripts to ${CONFIG_DIR}"
  fi
}

detect_arch() {
  arch="$(uname -m | tr '[:upper:]' '[:lower:]')"

  case "${arch}" in
    x86_64) arch="amd64" ;;
    armv*) arch="arm" ;;
    arm64) arch="arm64" ;;
    aarch64) arch="arm64" ;;
    i686) arch="386" ;;
  esac

  if [ "${arch}" = "arm64" ] && [ "$(getconf LONG_BIT)" -eq 32 ]; then
    arch=arm
  fi

  printf '%s' "${arch}"
}


detect_platform() {
  platform="$(uname -s | awk '{print tolower($0)}')"

  case "${platform}" in
    linux) platform="linux" ;;
    darwin) platform="darwin" ;;
  esac

  printf '%s' "${platform}"
}

install() {
  ARCH=$(detect_arch)
  PLATFORM=$(detect_platform)
  TARGET="${PLATFORM}-${ARCH}"

  good=$(
    IFS=" "
    for t in $SUPPORTED_TARGETS; do
    if [ "${t}" = "${TARGET}" ]; then
      printf 1
      break
    fi
    done
  )

  if [ "${good}" != "1" ]; then
    error "${ARCH} builds for ${PLATFORM} are not available for GCP Tunneler"
  fi

  info "\nℹ️  Installing GCP Tunneler for ${TARGET} in ${INSTALL_DIR}"

  validate_temp_directory

  TEMP_FILE=${TEMP_DIR}/gcp-tunneler.tar.gz

  BINARY=${INSTALL_DIR}/gcp-tunneler

  URL=https://github.com/Jacobus-afk/GCP-Tunneler/releases/latest/download/GCP-Tunneler_Linux_x86_64.tar.gz

  info "⬇️  Downloading GCP Tunneler from ${URL}"

  HTTP_RESPONSE=$(curl -s -f -L $URL -o $TEMP_FILE -w "%{http_code}")

  if [ $HTTP_RESPONSE != "200" ] || [ ! -f $TEMP_FILE ]; then
    error "Unable to download executable at ${URL}\nPlease validate your curl, connection and/or proxy settings"
  fi

  tar -xzvf $TEMP_FILE -C $TEMP_DIR

  chmod +x ${TEMP_DIR}/gcp-tunneler

  cp ${TEMP_DIR}/gcp-tunneler ${BINARY}

  install_scripts

  info "🚇 Installation complete."
}

validate_dependencies
set_install_directory
validate_install_directory
install
