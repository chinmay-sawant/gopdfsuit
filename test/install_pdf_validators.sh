#!/usr/bin/env bash
# Install PDF compliance validators into project-local directories (gitignored).
#
# Installs:
#   - veraPDF CLI  -> <repo>/verapdf/          (PDF/A-4, PDF/UA-2 ISO profiles)
#   - avalpdf      -> <repo>/.pdf-validators/  (structure/accessibility heuristics)
#
# Optional (not installed here - license required):
#   - callas pdfToolbox CLI (Docker: callassoftware/pdftoolbox-cli)
#   - axes4 PAC (Windows GUI only)
#
# Usage:
#   make install-pdf-validators
#   bash test/install_pdf_validators.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VALIDATORS_DIR="${REPO_ROOT}/.pdf-validators"
VENV_DIR="${VALIDATORS_DIR}/venv"
REQUIREMENTS="${REPO_ROOT}/test/pdf_validators_requirements.txt"

echo "==> Installing veraPDF..."
bash "${REPO_ROOT}/test/install_verapdf.sh"

echo ""
echo "==> Installing supplementary validators into ${VALIDATORS_DIR}..."
mkdir -p "${VALIDATORS_DIR}"

if [[ ! -d "${VENV_DIR}" ]]; then
    python3 -m venv "${VENV_DIR}"
fi

# shellcheck disable=SC1091
source "${VENV_DIR}/bin/activate"
python -m pip install --upgrade pip
python -m pip install -r "${REQUIREMENTS}"

echo ""
echo "Installed validators:"
echo "  veraPDF:  ${REPO_ROOT}/verapdf/verapdf"
"${REPO_ROOT}/verapdf/verapdf" --version 2>&1 | head -1 || true
echo "  avalpdf:  ${VENV_DIR}/bin/avalpdf"
"${VENV_DIR}/bin/avalpdf" --version

cat > "${VALIDATORS_DIR}/README.txt" <<EOF
Project-local PDF validators (gitignored). Installed by test/install_pdf_validators.sh.

veraPDF   ${REPO_ROOT}/verapdf/verapdf
avalpdf   ${VENV_DIR}/bin/avalpdf
structure ${REPO_ROOT}/test/structure_tree_check.py

See guides/PDF_VALIDATORS.md for usage and optional commercial validators.
EOF

echo ""
echo "Done. Run: make test-verify-pdfs"