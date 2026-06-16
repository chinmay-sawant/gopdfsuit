#!/usr/bin/env bash
# Install veraPDF CLI into the project-local verapdf/ directory (gitignored).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${REPO_ROOT}/verapdf"
VERAPDF_BIN="${INSTALL_DIR}/verapdf"

if [[ -x "${VERAPDF_BIN}" ]]; then
    echo "veraPDF already installed:"
    "${VERAPDF_BIN}" --version
    exit 0
fi

if ! command -v java >/dev/null 2>&1; then
    echo "Java is required. Install with: sudo apt-get install -y openjdk-11-jre-headless" >&2
    exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Downloading veraPDF installer..."
curl -fsSL -o "${TMP_DIR}/verapdf-installer.zip" https://software.verapdf.org/releases/verapdf-installer.zip
unzip -q "${TMP_DIR}/verapdf-installer.zip" -d "${TMP_DIR}/installer"

INSTALLER_DIR="$(find "${TMP_DIR}/installer" -maxdepth 1 -type d -name 'verapdf-greenfield-*' | head -1)"
if [[ -z "${INSTALLER_DIR}" ]]; then
    echo "Could not find veraPDF installer directory in archive" >&2
    exit 1
fi

cat > "${INSTALLER_DIR}/auto-install.xml" <<EOF
<AutomatedInstallation langpack="eng">
    <com.izforge.izpack.panels.htmlhello.HTMLHelloPanel id="welcome"/>
    <com.izforge.izpack.panels.target.TargetPanel id="install_dir">
        <installpath>${INSTALL_DIR}</installpath>
    </com.izforge.izpack.panels.target.TargetPanel>
    <com.izforge.izpack.panels.packs.PacksPanel id="sdk_pack_select">
        <pack index="0" name="veraPDF GUI" selected="true"/>
        <pack index="1" name="veraPDF Mac and *nix Scripts" selected="true"/>
        <pack index="2" name="veraPDF Validation model" selected="false"/>
        <pack index="3" name="veraPDF Documentation" selected="false"/>
        <pack index="4" name="veraPDF Sample Plugins" selected="false"/>
    </com.izforge.izpack.panels.packs.PacksPanel>
    <com.izforge.izpack.panels.install.InstallPanel id="install"/>
    <com.izforge.izpack.panels.finish.FinishPanel id="finish"/>
</AutomatedInstallation>
EOF

echo "Installing veraPDF to ${INSTALL_DIR}..."
(cd "${INSTALLER_DIR}" && ./verapdf-install ./auto-install.xml)

echo "Installed:"
"${VERAPDF_BIN}" --version