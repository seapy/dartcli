#!/bin/sh
set -e

REPO="seapy/dartcli"
BINARY="dartcli"

# ── OS / Arch 감지 ────────────────────────────────────────────────────────────

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)
    echo "지원하지 않는 운영체제입니다: $OS"
    echo "Windows 사용자는 https://github.com/$REPO/releases 에서 직접 다운로드하세요."
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64 | amd64) ARCH="amd64" ;;
  arm64 | aarch64) ARCH="arm64" ;;
  *)
    echo "지원하지 않는 아키텍처입니다: $ARCH"
    exit 1
    ;;
esac

# ── 설치 디렉토리 결정 (sudo 없이 사용자 디렉토리 우선) ──────────────────────

if [ -n "$DARTCLI_INSTALL_DIR" ]; then
  INSTALL_DIR="$DARTCLI_INSTALL_DIR"
elif [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="$HOME/.local/bin"
fi

mkdir -p "$INSTALL_DIR"

# ── 최신 버전 조회 ────────────────────────────────────────────────────────────

if command -v curl >/dev/null 2>&1; then
  FETCH="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
  FETCH="wget -qO-"
else
  echo "curl 또는 wget이 필요합니다."
  exit 1
fi

echo "최신 버전을 확인하는 중..."
VERSION=$(
  $FETCH "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": *"\(.*\)".*/\1/'
)

if [ -z "$VERSION" ]; then
  echo "버전 정보를 가져오지 못했습니다. 잠시 후 다시 시도해 주세요."
  exit 1
fi

echo "설치 버전: $VERSION  ($OS/$ARCH)"

# ── 다운로드 & 설치 ───────────────────────────────────────────────────────────

ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
TMP=$(mktemp -d)

echo "다운로드 중: $URL"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
else
  wget -qO "$TMP/$ARCHIVE" "$URL"
fi

tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"
rm -rf "$TMP"

# ── 완료 ─────────────────────────────────────────────────────────────────────

echo ""
echo "설치 완료: $INSTALL_DIR/$BINARY"

# PATH에 없으면 안내
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo ""
    echo "※ $INSTALL_DIR 가 PATH에 없습니다. 아래 줄을 셸 설정 파일에 추가하세요."
    echo ""
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo ""
    echo "  (bash: ~/.bashrc 또는 ~/.bash_profile)"
    echo "  (zsh:  ~/.zshrc)"
    ;;
esac

echo ""
echo "시작하기:"
echo "  export DART_API_KEY=<발급받은_키>"
echo "  $BINARY --help"
