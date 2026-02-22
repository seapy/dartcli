#!/bin/sh
set -e

REPO="seapy/dartcli"
BINARY="dartcli"
INSTALL_DIR="/usr/local/bin"

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
$FETCH "$URL" -o "$TMP/$ARCHIVE" 2>/dev/null || {
  # wget은 -o 옵션 형태가 다름
  wget -qO "$TMP/$ARCHIVE" "$URL"
}

tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
rm "$TMP/$ARCHIVE"

# /usr/local/bin 쓰기 권한 확인 → 없으면 sudo 사용
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "관리자 권한으로 $INSTALL_DIR 에 설치합니다 (sudo 비밀번호가 필요할 수 있습니다)."
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"
rm -rf "$TMP"

# ── 완료 ─────────────────────────────────────────────────────────────────────

echo ""
echo "설치 완료: $INSTALL_DIR/$BINARY"
echo ""
echo "시작하기:"
echo "  export DART_API_KEY=<발급받은_키>"
echo "  $BINARY --help"
