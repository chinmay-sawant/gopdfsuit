#!/usr/bin/env bash
# veraPDF validity, optional PDF/A compliance, and PDF size variance checks.
#
# Usage:
#   ./test/verify_pdfs.sh              # post-test validation (make test)
#   ./test/verify_pdfs.sh --scan-all   # scan every PDF under sampledata/
#
# Environment:
#   VERIFY_PDFS_JOBS  Max parallel veraPDF workers (default: nproc or 4)
#   VERAPDF_BIN       Path to veraPDF CLI (default: <repo>/verapdf/verapdf)
#
# Post-test manifest is built from:
#   - Every sampledata/**/generated.* baseline (excluding oldata/)
#   - temp_* outputs in the same directory as each baseline
#   - Additional reference baselines (split/*.pdf, split/maxperfile.zip, etc.)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SAMPLEDATA="${REPO_ROOT}/sampledata"
VERAPDF="${VERAPDF_BIN:-${REPO_ROOT}/verapdf/verapdf}"
PARALLEL_JOBS="${VERIFY_PDFS_JOBS:-$(
    if command -v nproc >/dev/null 2>&1; then
        nproc
    else
        echo 4
    fi
)}"

if [[ ! -x "${VERAPDF}" ]]; then
    echo "veraPDF CLI not found at ${VERAPDF}" >&2
    echo "Install with: make install-verapdf" >&2
    exit 1
fi

human_size() {
    local bytes="$1"
    if (( bytes >= 1048576 )); then
        printf "%.2f MB" "$(awk "BEGIN {print ${bytes}/1048576}")"
    elif (( bytes >= 1024 )); then
        printf "%.1f KB" "$(awk "BEGIN {print ${bytes}/1024}")"
    else
        printf "%d B" "${bytes}"
    fi
}

media_type_for_path() {
    local rel="$1"
    case "${rel##*.}" in
        pdf) echo "pdf" ;;
        png) echo "png" ;;
        zip) echo "zip" ;;
        *) echo "binary" ;;
    esac
}

# Parse the PDF with veraPDF (validation off). Success means Acrobat-openable structure.
check_valid_pdf() {
    local pdf="$1"

    if [[ ! -s "${pdf}" ]]; then
        echo "invalid|empty or missing file"
        return
    fi

    local output=""
    local exit_code=0
    output="$("${VERAPDF}" --off --extract lowLevelInfo --format text --loglevel 0 "${pdf}" 2>&1)" || exit_code=$?

    if (( exit_code == 0 )); then
        echo "valid|veraPDF parsed successfully"
        return
    fi

    local msg="${output//$'\n'/; }"
    msg="${msg//$'\r'/}"
    if [[ -z "${msg}" ]]; then
        msg="veraPDF parse failed (exit ${exit_code})"
    fi
    echo "invalid|${msg}"
}

check_valid_png() {
    local file="$1"
    if [[ ! -s "${file}" ]]; then
        echo "invalid|empty or missing file"
        return
    fi
    local header
    header=$(head -c 8 "${file}" | od -An -tx1 | tr -d ' \n')
    if [[ "${header}" == 89504e470d0a1a0a ]]; then
        echo "valid|PNG signature ok"
        return
    fi
    echo "invalid|missing PNG signature"
}

check_valid_zip() {
    local file="$1"
    if [[ ! -s "${file}" ]]; then
        echo "invalid|empty or missing file"
        return
    fi
    local header
    header=$(head -c 4 "${file}")
    if [[ "${header}" == $'PK\x03\x04' ]]; then
        echo "valid|ZIP signature ok"
        return
    fi
    echo "invalid|missing ZIP signature"
}

check_valid_file() {
    local file="$1"
    local media="$2"
    case "${media}" in
        pdf) check_valid_pdf "${file}" ;;
        png) check_valid_png "${file}" ;;
        zip) check_valid_zip "${file}" ;;
        *)
            if [[ -s "${file}" ]]; then
                echo "valid|non-empty ${media} file"
            else
                echo "invalid|empty or missing file"
            fi
            ;;
    esac
}

check_verapdf_compliance() {
    local pdf="$1"
    shift
    local flavours=("$@")

    local ok=true
    local details=()

    for flavour in "${flavours[@]}"; do
        local output
        if output="$("${VERAPDF}" -f "${flavour}" --format text --loglevel 0 "${pdf}" 2>&1)"; then
            details+=("PASS ${flavour}")
        else
            ok=false
            local msg="${output//$'\n'/; }"
            details+=("FAIL ${flavour}: ${msg}")
        fi
    done

    if [[ "${ok}" == true ]]; then
        echo "compliant|$(IFS='; '; echo "${details[*]}")"
    else
        echo "not compliant|$(IFS='; '; echo "${details[*]}")"
    fi
}

size_within_tolerance() {
    local generated="$1"
    local baseline="$2"
    local tolerance="$3"

    local gen_size base_size diff
    gen_size=$(stat -c '%s' "${generated}")
    base_size=$(stat -c '%s' "${baseline}")
    diff=$((gen_size > base_size ? gen_size - base_size : base_size - gen_size))

    if (( diff <= tolerance )); then
        echo "ok|${gen_size}|${base_size}|${diff}"
    else
        echo "fail|${gen_size}|${base_size}|${diff}"
    fi
}

wait_for_slot() {
    local -n _running_ref="$1"
    local max_jobs="$2"
    while ((_running_ref >= max_jobs)); do
        if ! wait -n 2>/dev/null; then
            wait || true
        fi
        ((_running_ref--)) || true
    done
}

manifest_tolerance() {
    local rel="$1"
    local default_tol="$2"
    case "${rel}" in
        editor/temp_editor.pdf|editor/temp_editor_python.pdf) echo "8192" ;;
        filler/temp_filler.pdf|filler/temp_filler_python.pdf) echo "700" ;;
        filler/compressed/temp_filler_compressed.pdf) echo "700" ;;
        filler/compressed/temp_filler_compressed_python.pdf) echo "500" ;;
        htmltopdf/temp_htmltopdf.pdf|htmltopdf/temp_htmltopdf_python.pdf) echo "skip" ;;
        htmltoimg/temp_htmltoimage.png|htmltoimg/temp_htmltoimage_python.png) echo "skip" ;;
        *) echo "${default_tol}" ;;
    esac
}

manifest_compliance() {
    local rel="$1"
    case "${rel}" in
        editor/temp_editor.pdf|editor/temp_editor_python.pdf) echo "4,ua2" ;;
        *) echo "" ;;
    esac
}

manifest_add_unique() {
    local -n _seen_ref="$1"
    local -n _manifest_ref="$2"
    local key="$3"
    local entry="$4"

    if [[ -n "${_seen_ref[${key}]:-}" ]]; then
        return
    fi
    _seen_ref["${key}"]=1
    _manifest_ref+=("${entry}")
}

# Manifest entry: generated_rel|baseline_rel|tolerance|flavours|media
build_post_test_manifest() {
    POST_TEST_MANIFEST=()
    declare -A MANIFEST_SEEN=()

    # 1) All generated.* baselines under sampledata (excluding oldata/)
    while IFS= read -r -d '' baseline; do
        local rel="${baseline#${SAMPLEDATA}/}"
        local dir
        dir=$(dirname "${rel}")
        local dir_path="${SAMPLEDATA}/${dir}"
        local media
        media=$(media_type_for_path "${rel}")

        # Validate the committed baseline artifact itself
        manifest_add_unique MANIFEST_SEEN POST_TEST_MANIFEST "${rel}@baseline" \
            "${rel}|${rel}|0||${media}"

        # temp_* outputs in the same folder as the baseline
        if [[ -d "${dir_path}" ]]; then
            while IFS= read -r -d '' temp; do
                local temp_rel="${temp#${SAMPLEDATA}/}"
                local tol flavours entry_media
                tol=$(manifest_tolerance "${temp_rel}" "0")
                flavours=$(manifest_compliance "${temp_rel}")
                entry_media=$(media_type_for_path "${temp_rel}")
                manifest_add_unique MANIFEST_SEEN POST_TEST_MANIFEST "${temp_rel}" \
                    "${temp_rel}|${rel}|${tol}|${flavours}|${entry_media}"
            done < <(find "${dir_path}" -maxdepth 1 -type f -name 'temp_*' -print0 2>/dev/null | sort -z)
        fi
    done < <(find "${SAMPLEDATA}" \( -path '*/oldata/*' \) -prune -o -type f -name 'generated.*' -print0 | sort -z)

    # 2) Split reference baselines (not named generated.*)
    local split_entries=(
        "split/temp_split.pdf|split/split.pdf|0||pdf"
        "split/temp_split_python.pdf|split/split.pdf|0||pdf"
        "split/temp_split_range.pdf|split/split_range.pdf|0||pdf"
        "split/temp_split_range_python.pdf|split/split_range.pdf|0||pdf"
        "split/temp_maxperfile.zip|split/maxperfile.zip|0||zip"
        "split/split.pdf|split/split.pdf|0||pdf"
        "split/split_range.pdf|split/split_range.pdf|0||pdf"
        "split/maxperfile.zip|split/maxperfile.zip|0||zip"
    )
    local entry
    for entry in "${split_entries[@]}"; do
        local key="${entry%%|*}"
        manifest_add_unique MANIFEST_SEEN POST_TEST_MANIFEST "${key}" "${entry}"
    done

    # 3) Other integration outputs without generated.* baselines
    local extra_entries=(
        "financialreport/financial_report.pdf||0||pdf"
        "financialreport/temp_financial_report_redacted.pdf||0||pdf"
        "typstsyntax/typst_math_showcase.pdf||0||pdf"
        "typstsyntax/typst_math_showcase_python.pdf||0||pdf"
        "typstsyntax/typst_sample.pdf||0||pdf"
        "typstsyntax/typst_sample_python.pdf||0||pdf"
        "split/temp_split_maxperfile_python.pdf||0||pdf"
        "gopdflib/zerodha/zerodha_hft_output.pdf||0|4,ua2|pdf"
        "gopdflib/zerodha/zerodha_retail_output.pdf||0|4,ua2|pdf"
        "gopdflib/zerodha/zerodha_active_output.pdf||0|4,ua2|pdf"
    )
    for entry in "${extra_entries[@]}"; do
        local key="${entry%%|*}"
        manifest_add_unique MANIFEST_SEEN POST_TEST_MANIFEST "${key}" "${entry}"
    done
}

verify_manifest_entry() {
    local entry="$1"
    local failures=0

    IFS='|' read -r generated_rel baseline_rel tolerance flavours_csv media <<< "${entry}"
    media="${media:-pdf}"
    local generated="${SAMPLEDATA}/${generated_rel}"

    echo ""
    echo "==> ${generated_rel}"

    if [[ ! -f "${generated}" ]]; then
        echo "SKIP ${generated_rel}: output not found (test may have been skipped)"
        return 0
    fi

    local validity_result validity_status validity_details
    validity_result=$(check_valid_file "${generated}" "${media}")
    validity_status="${validity_result%%|*}"
    validity_details="${validity_result#*|}"
    if [[ "${validity_status}" == "valid" ]]; then
        echo "PASS valid ${media}: ${validity_details}"
    else
        echo "FAIL valid ${media}: ${validity_details}"
        ((failures++)) || true
    fi

    if [[ -n "${baseline_rel}" ]]; then
        if [[ "${tolerance}" == "skip" ]]; then
            local gen_size
            gen_size=$(stat -c '%s' "${generated}")
            echo "INFO size: ${gen_size} bytes (baseline ${baseline_rel} present — size check skipped; live HTML output)"
        else
            local baseline="${SAMPLEDATA}/${baseline_rel}"
            if [[ ! -f "${baseline}" ]]; then
                echo "FAIL size: baseline ${baseline_rel} not found"
                ((failures++)) || true
            else
                local size_result size_status gen_size base_size diff
                size_result=$(size_within_tolerance "${generated}" "${baseline}" "${tolerance}")
                IFS='|' read -r size_status gen_size base_size diff <<< "${size_result}"
                if [[ "${size_status}" == "ok" ]]; then
                    echo "PASS size: ${gen_size} bytes (baseline ${base_size}, diff ${diff}, tolerance ${tolerance})"
                else
                    echo "FAIL size: ${gen_size} bytes vs baseline ${base_size} (diff ${diff} > tolerance ${tolerance})"
                    ((failures++)) || true
                fi
            fi
        fi
    else
        local gen_size
        gen_size=$(stat -c '%s' "${generated}")
        echo "INFO size: ${gen_size} bytes (no baseline — size variance check skipped)"
    fi

    if [[ -n "${flavours_csv}" && "${media}" == "pdf" ]]; then
        IFS=',' read -ra flavours <<< "${flavours_csv}"
        local compliance_result status details
        compliance_result=$(check_verapdf_compliance "${generated}" "${flavours[@]}")
        status="${compliance_result%%|*}"
        details="${compliance_result#*|}"
        if [[ "${status}" == "compliant" ]]; then
            echo "PASS compliance: ${details}"
        else
            echo "FAIL compliance: ${details}"
            ((failures++)) || true
        fi
    fi

    return "${failures}"
}

scan_one_pdf() {
    local pdf="$1"
    local rel="${pdf#${SAMPLEDATA}/}"
    local size_bytes size_human result status details status_label

    size_bytes=$(stat -c '%s' "${pdf}")
    size_human=$(human_size "${size_bytes}")

    result=$(check_valid_pdf "${pdf}")
    status="${result%%|*}"
    details="${result#*|}"

    if [[ "${status}" == "valid" ]]; then
        status_label="valid"
    else
        status_label="invalid"
    fi

    details="${details//|/\\|}"
    printf "| %s | %s (%s) | %s | %s |\n" "${rel}" "${size_human}" "${size_bytes}" "${status_label}" "${details}"
    if [[ "${status}" == "valid" ]]; then
        return 0
    fi
    return 1
}

post_test() {
    build_post_test_manifest
    echo "Post-test PDF validation (${#POST_TEST_MANIFEST[@]} entries, ${PARALLEL_JOBS} workers)..."
    echo "Baselines: sampledata/**/generated.* (excluding oldata/), plus split/financial/typst outputs"

    local tmpdir
    tmpdir=$(mktemp -d)
    trap 'rm -rf "${tmpdir}"' RETURN

    local idx=0
    local running=0

    for entry in "${POST_TEST_MANIFEST[@]}"; do
        wait_for_slot running "${PARALLEL_JOBS}"
        (
            set +e
            verify_manifest_entry "${entry}" > "${tmpdir}/${idx}.log"
            echo $? > "${tmpdir}/${idx}.rc"
        ) &
        ((running++)) || true
        ((idx++)) || true
    done

    while ((running > 0)); do
        wait -n 2>/dev/null || wait || true
        ((running--)) || true
    done

    local failures=0
    local i
    for ((i = 0; i < idx; i++)); do
        cat "${tmpdir}/${i}.log"
        failures=$((failures + $(cat "${tmpdir}/${i}.rc")))
    done

    echo ""
    if ((failures > 0)); then
        echo "Post-test PDF validation failed (${failures} issue(s))"
        exit 1
    fi
    echo "Post-test PDF validation passed"
}

scan_all() {
    echo "Scanning PDFs under ${SAMPLEDATA} with veraPDF parse check (${PARALLEL_JOBS} workers)..."
    echo ""
    printf "| PDF | Size | Valid | Details |\n"
    printf "|-----|------|-------|----------|\n"

    local tmpdir
    tmpdir=$(mktemp -d)
    trap 'rm -rf "${tmpdir}"' RETURN

    local -a pdfs=()
    while IFS= read -r -d '' pdf; do
        pdfs+=("${pdf}")
    done < <(find "${SAMPLEDATA}" -type f -iname '*.pdf' -print0 | sort -z)

    local total=${#pdfs[@]}
    local idx=0
    local running=0

    for pdf in "${pdfs[@]}"; do
        wait_for_slot running "${PARALLEL_JOBS}"
        (
            set +e
            scan_one_pdf "${pdf}" > "${tmpdir}/${idx}.row"
            echo $? > "${tmpdir}/${idx}.status"
        ) &
        ((running++)) || true
        ((idx++)) || true
    done

    while ((running > 0)); do
        wait -n 2>/dev/null || wait || true
        ((running--)) || true
    done

    local valid=0 invalid=0
    local i
    for ((i = 0; i < total; i++)); do
        cat "${tmpdir}/${i}.row"
        if [[ "$(cat "${tmpdir}/${i}.status")" == "0" ]]; then
            ((valid++)) || true
        else
            ((invalid++)) || true
        fi
    done

    echo ""
    echo "Summary: ${total} PDFs — ${valid} valid, ${invalid} invalid"
}

case "${1:-}" in
    --scan-all)
        scan_all
        ;;
    --help|-h)
        echo "Usage: $0 [--scan-all]"
        echo "Environment: VERIFY_PDFS_JOBS, VERAPDF_BIN"
        ;;
    *)
        post_test
        ;;
esac