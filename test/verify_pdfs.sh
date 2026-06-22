#!/usr/bin/env bash
# PDF validation: veraPDF (PDF/A-4, PDF/UA-2), structure-tree consistency,
# avalpdf heuristics, validity parse checks, and size variance.
#
# Usage:
#   ./test/verify_pdfs.sh                      # post-test validation (make test)
#   ./test/verify_pdfs.sh --scan-all             # scan every PDF under sampledata/
#   ./test/verify_pdfs.sh --scan-all-compliance  # scan-all + PDF/A-4 and PDF/UA-2 table
#   ./test/verify_pdfs.sh --zerodha-only         # Zerodha retail/active/HFT PDF/A-4 + PDF/UA-2
#
# Environment:
#   VERIFY_PDFS_JOBS  Max parallel veraPDF workers (default: nproc or 4)
#   VERAPDF_BIN       Path to veraPDF CLI (default: <repo>/verapdf/verapdf)
#   AVALPDF_BIN       Path to avalpdf CLI (default: <repo>/.pdf-validators/venv/bin/avalpdf)
#   VERIFY_STRUCTURE_TREE  Run structure_tree_check.py on compliance PDFs (default: 1)
#   VERIFY_AVALPDF    Run avalpdf on compliance PDFs (default: 1)
#   VERIFY_AVALPDF_STRICT  Fail on avalpdf issues (default: 0 — warnings only)
#
# Post-test manifest is built from:
#   - Every sampledata/**/generated.* baseline (excluding oldata/)
#   - temp_* outputs in the same directory as each baseline
#   - Additional reference baselines (split/*.pdf, split/maxperfile.zip, etc.)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SAMPLEDATA="${REPO_ROOT}/sampledata"
VERAPDF="${VERAPDF_BIN:-${REPO_ROOT}/verapdf/verapdf}"
VERAPDF_REPORT="${REPO_ROOT}/test/verapdf_report.py"
STRUCTURE_TREE_CHECK="${REPO_ROOT}/test/structure_tree_check.py"
AVALPDF="${AVALPDF_BIN:-${REPO_ROOT}/.pdf-validators/venv/bin/avalpdf}"
VERIFY_STRUCTURE_TREE="${VERIFY_STRUCTURE_TREE:-1}"
VERIFY_AVALPDF="${VERIFY_AVALPDF:-1}"
VERIFY_AVALPDF_STRICT="${VERIFY_AVALPDF_STRICT:-0}"
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

if [[ ! -f "${VERAPDF_REPORT}" ]]; then
    echo "veraPDF report helper not found at ${VERAPDF_REPORT}" >&2
    exit 1
fi

if [[ -t 1 && -z "${NO_COLOR:-}" ]]; then
    COLOR_RESET=$'\033[0m'
    COLOR_BOLD=$'\033[1m'
    COLOR_GREEN=$'\033[32m'
    COLOR_RED=$'\033[31m'
    COLOR_YELLOW=$'\033[33m'
    COLOR_CYAN=$'\033[36m'
else
    COLOR_RESET=""
    COLOR_BOLD=""
    COLOR_GREEN=""
    COLOR_RED=""
    COLOR_YELLOW=""
    COLOR_CYAN=""
fi

print_pass() {
    printf '%bPASS%b %s\n' "${COLOR_GREEN}" "${COLOR_RESET}" "$*"
}

print_fail() {
    printf '%bFAIL%b %s\n' "${COLOR_RED}${COLOR_BOLD}" "${COLOR_RESET}" "$*"
}

print_info() {
    printf '%bINFO%b %s\n' "${COLOR_CYAN}" "${COLOR_RESET}" "$*"
}

print_skip() {
    printf '%bSKIP%b %s\n' "${COLOR_YELLOW}" "${COLOR_RESET}" "$*"
}

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

check_structure_tree() {
    local pdf="$1"

    if [[ "${VERIFY_STRUCTURE_TREE}" != "1" ]]; then
        echo "skip|structure-tree checks disabled"
        return 0
    fi
    if [[ ! -f "${STRUCTURE_TREE_CHECK}" ]]; then
        echo "skip|structure_tree_check.py not found"
        return 0
    fi

    local output exit_code=0
    output="$(python3 "${STRUCTURE_TREE_CHECK}" "${pdf}" 2>&1)" || exit_code=$?
    if ((exit_code == 0)); then
        echo "ok|$(printf '%s\n' "${output}" | tail -1)"
    else
        echo "fail|${output}"
    fi
}

check_avalpdf() {
    local pdf="$1"

    if [[ "${VERIFY_AVALPDF}" != "1" ]]; then
        echo "skip|avalpdf checks disabled"
        return 0
    fi
    if [[ ! -x "${AVALPDF}" ]]; then
        echo "skip|avalpdf not installed (make install-pdf-validators)"
        return 0
    fi

    local tmpdir report_json exit_code=0
    tmpdir="$(mktemp -d)"
    "${AVALPDF}" "${pdf}" --report -o "${tmpdir}" --quiet 2>/dev/null || exit_code=$?
    report_json="$(find "${tmpdir}" -maxdepth 1 -name '*validation_report.json' -print -quit 2>/dev/null || true)"

    if ((exit_code != 0)); then
        rm -rf "${tmpdir}"
        echo "fail|avalpdf exited ${exit_code}"
        return
    fi
    if [[ -z "${report_json}" ]]; then
        rm -rf "${tmpdir}"
        echo "fail|avalpdf produced no validation report"
        return
    fi

    local counts
    counts="$(python3 - "${report_json}" <<'PY'
import json, sys
with open(sys.argv[1], encoding="utf-8") as handle:
    payload = json.load(handle)
results = payload.get("validation_results", payload)
issues = len(results.get("issues") or [])
warnings = len(results.get("warnings") or [])
print(f"{issues} {warnings}")
PY
)"
    rm -rf "${tmpdir}"
    local issues="${counts%% *}"
    local warnings="${counts##* }"

    if [[ "${VERIFY_AVALPDF_STRICT}" == "1" && "${issues}" -gt 0 ]]; then
        echo "fail|avalpdf ${issues} issue(s), ${warnings} warning(s)"
    elif [[ "${issues}" -gt 0 || "${warnings}" -gt 0 ]]; then
        echo "warn|avalpdf ${issues} issue(s), ${warnings} warning(s) (non-blocking)"
    else
        echo "ok|avalpdf clean"
    fi
}

check_verapdf_compliance() {
    local pdf="$1"
    local json_dir="$2"
    shift 2
    local flavours=("$@")

    local ok=true
    local details=()
    local json_files=()

    mkdir -p "${json_dir}"

    for flavour in "${flavours[@]}"; do
        local json_out="${json_dir}/$(basename "${pdf}")_${flavour}.json"
        local output
        local exit_code=0
        output="$(
            python3 "${VERAPDF_REPORT}" check \
                --verapdf "${VERAPDF}" \
                --pdf "${pdf}" \
                --flavour "${flavour}" \
                --json-out "${json_out}" \
                --sampledata "${SAMPLEDATA}/" \
                2>&1
        )" || exit_code=$?

        json_files+=("${json_out}")
        if [[ -n "${output}" ]]; then
            printf '%s\n' "${output}" >&2
        fi

        if ((exit_code == 0)); then
            details+=("PASS ${flavour}")
        else
            ok=false
            details+=("FAIL ${flavour}")
        fi
    done

    if [[ "${ok}" == true ]]; then
        echo "compliant|$(IFS='; '; echo "${details[*]}")|$(IFS=','; echo "${json_files[*]}")"
    else
        echo "not compliant|$(IFS='; '; echo "${details[*]}")|$(IFS=','; echo "${json_files[*]}")"
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
    zerodha_compliance_entries
    local extra_entries=(
        "financialreport/financial_report.pdf||0||pdf"
        "financialreport/temp_financial_report_redacted.pdf||0||pdf"
        "typstsyntax/typst_math_showcase.pdf||0||pdf"
        "typstsyntax/typst_math_showcase_python.pdf||0||pdf"
        "typstsyntax/typst_sample.pdf||0||pdf"
        "typstsyntax/typst_sample_python.pdf||0||pdf"
        "split/temp_split_maxperfile_python.pdf||0||pdf"
        "${ZERODHA_MANIFEST[@]}"
    )
    for entry in "${extra_entries[@]}"; do
        local key="${entry%%|*}"
        manifest_add_unique MANIFEST_SEEN POST_TEST_MANIFEST "${key}" "${entry}"
    done
}

verify_manifest_entry() {
    local entry="$1"
    local json_dir="$2"
    local failures=0

    IFS='|' read -r generated_rel baseline_rel tolerance flavours_csv media <<< "${entry}"
    media="${media:-pdf}"
    local generated="${SAMPLEDATA}/${generated_rel}"

    echo ""
    printf '%b==> %s%b\n' "${COLOR_BOLD}" "${generated_rel}" "${COLOR_RESET}"

    if [[ ! -f "${generated}" ]]; then
        print_skip "${generated_rel}: output not found (test may have been skipped)"
        return 0
    fi

    local validity_result validity_status validity_details
    validity_result=$(check_valid_file "${generated}" "${media}")
    validity_status="${validity_result%%|*}"
    validity_details="${validity_result#*|}"
    if [[ "${validity_status}" == "valid" ]]; then
        print_pass "valid ${media}: ${validity_details}"
    else
        print_fail "valid ${media}: ${validity_details}"
        ((failures++)) || true
    fi

    if [[ -n "${baseline_rel}" ]]; then
        if [[ "${tolerance}" == "skip" ]]; then
            local gen_size
            gen_size=$(stat -c '%s' "${generated}")
            print_info "size: ${gen_size} bytes (baseline ${baseline_rel} present — size check skipped; live HTML output)"
        else
            local baseline="${SAMPLEDATA}/${baseline_rel}"
            if [[ ! -f "${baseline}" ]]; then
                print_fail "size: baseline ${baseline_rel} not found"
                ((failures++)) || true
            else
                local size_result size_status gen_size base_size diff
                size_result=$(size_within_tolerance "${generated}" "${baseline}" "${tolerance}")
                IFS='|' read -r size_status gen_size base_size diff <<< "${size_result}"
                if [[ "${size_status}" == "ok" ]]; then
                    print_pass "size: ${gen_size} bytes (baseline ${base_size}, diff ${diff}, tolerance ${tolerance})"
                else
                    print_fail "size: ${gen_size} bytes vs baseline ${base_size} (diff ${diff} > tolerance ${tolerance})"
                    ((failures++)) || true
                fi
            fi
        fi
    else
        local gen_size
        gen_size=$(stat -c '%s' "${generated}")
        print_info "size: ${gen_size} bytes (no baseline — size variance check skipped)"
    fi

    if [[ -n "${flavours_csv}" && "${media}" == "pdf" ]]; then
        IFS=',' read -ra flavours <<< "${flavours_csv}"
        local compliance_result status details json_list
        compliance_result=$(check_verapdf_compliance "${generated}" "${json_dir}" "${flavours[@]}")
        status="${compliance_result%%|*}"
        details="${compliance_result#*|}"
        json_list="${details##*|}"
        details="${details%%|*}"
        if [[ "${status}" == "compliant" ]]; then
            print_pass "compliance: ${details}"
        else
            print_fail "compliance: ${details}"
            ((failures++)) || true
        fi
        if [[ -n "${json_list}" ]]; then
            printf 'COMPLIANCE_JSON:%s\n' "${json_list}"
        fi

        local struct_result struct_status struct_details
        struct_result=$(check_structure_tree "${generated}")
        struct_status="${struct_result%%|*}"
        struct_details="${struct_result#*|}"
        case "${struct_status}" in
            ok)
                print_pass "structure-tree: ${struct_details}"
                ;;
            fail)
                print_fail "structure-tree: ${struct_details}"
                ((failures++)) || true
                ;;
            skip)
                print_skip "structure-tree: ${struct_details}"
                ;;
        esac

        local aval_result aval_status aval_details
        aval_result=$(check_avalpdf "${generated}")
        aval_status="${aval_result%%|*}"
        aval_details="${aval_result#*|}"
        case "${aval_status}" in
            ok)
                print_pass "avalpdf: ${aval_details}"
                ;;
            warn)
                print_info "avalpdf: ${aval_details}"
                ;;
            fail)
                print_fail "avalpdf: ${aval_details}"
                ((failures++)) || true
                ;;
            skip)
                print_skip "avalpdf: ${aval_details}"
                ;;
        esac
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

zerodha_compliance_entries() {
    ZERODHA_MANIFEST=(
        "gopdflib/zerodha/zerodha_hft_output.pdf||0|4,ua2|pdf"
        "gopdflib/zerodha/zerodha_retail_output.pdf||0|4,ua2|pdf"
        "gopdflib/zerodha/zerodha_active_output.pdf||0|4,ua2|pdf"
    )
}

zerodha_only() {
    zerodha_compliance_entries
    echo "Zerodha PDF compliance validation (${#ZERODHA_MANIFEST[@]} PDFs, PDF/A-4 + PDF/UA-2, ${PARALLEL_JOBS} workers)..."

    local tmpdir
    tmpdir=$(mktemp -d)
    trap 'rm -rf "${tmpdir}"' RETURN

    local idx=0
    local running=0
    local -a compliance_json_files=()

    for entry in "${ZERODHA_MANIFEST[@]}"; do
        wait_for_slot running "${PARALLEL_JOBS}"
        (
            set +e
            verify_manifest_entry "${entry}" "${tmpdir}/compliance" > "${tmpdir}/${idx}.log" 2>&1
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
        while IFS= read -r line; do
            if [[ "${line}" == COMPLIANCE_JSON:* ]]; then
                local json_list="${line#COMPLIANCE_JSON:}"
                local json_path
                IFS=',' read -ra json_paths <<< "${json_list}"
                for json_path in "${json_paths[@]}"; do
                    if [[ -f "${json_path}" ]]; then
                        compliance_json_files+=("${json_path}")
                    fi
                done
                continue
            fi
            printf '%s\n' "${line}"
        done < "${tmpdir}/${i}.log"
        failures=$((failures + $(cat "${tmpdir}/${i}.rc")))
    done

    if ((${#compliance_json_files[@]} > 0)); then
        python3 "${VERAPDF_REPORT}" table --sampledata "${SAMPLEDATA}/" "${compliance_json_files[@]}"
    fi

    echo ""
    if ((failures > 0)); then
        print_fail "Zerodha PDF compliance validation failed (${failures} issue(s))"
        exit 1
    fi
    print_pass "Zerodha PDF compliance validation passed"
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
    local -a compliance_json_files=()

    for entry in "${POST_TEST_MANIFEST[@]}"; do
        wait_for_slot running "${PARALLEL_JOBS}"
        (
            set +e
            verify_manifest_entry "${entry}" "${tmpdir}/compliance" > "${tmpdir}/${idx}.log" 2>&1
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
        while IFS= read -r line; do
            if [[ "${line}" == COMPLIANCE_JSON:* ]]; then
                local json_list="${line#COMPLIANCE_JSON:}"
                local json_path
                IFS=',' read -ra json_paths <<< "${json_list}"
                for json_path in "${json_paths[@]}"; do
                    if [[ -f "${json_path}" ]]; then
                        compliance_json_files+=("${json_path}")
                    fi
                done
                continue
            fi
            printf '%s\n' "${line}"
        done < "${tmpdir}/${i}.log"
        failures=$((failures + $(cat "${tmpdir}/${i}.rc")))
    done

    if ((${#compliance_json_files[@]} > 0)); then
        python3 "${VERAPDF_REPORT}" table --sampledata "${SAMPLEDATA}/" "${compliance_json_files[@]}"
    fi

    echo ""
    if ((failures > 0)); then
        print_fail "Post-test PDF validation failed (${failures} issue(s))"
        exit 1
    fi
    print_pass "Post-test PDF validation passed"
}

scan_one_pdf_compliance() {
    local pdf="$1"
    local json_dir="$2"
    local rel="${pdf#${SAMPLEDATA}/}"
    local -a json_files=()
    local flavours=("4" "ua2")
    local flavour

    mkdir -p "${json_dir}"

    for flavour in "${flavours[@]}"; do
        local json_out="${json_dir}/$(basename "${pdf}")_${flavour}.json"
        python3 "${VERAPDF_REPORT}" check \
            --verapdf "${VERAPDF}" \
            --pdf "${pdf}" \
            --flavour "${flavour}" \
            --json-out "${json_out}" \
            --sampledata "${SAMPLEDATA}/" \
            >/dev/null 2>&1 || true
        json_files+=("${json_out}")
    done

    printf 'COMPLIANCE_JSON:%s\n' "$(IFS=','; echo "${json_files[*]}")"
}

scan_all() {
    local with_compliance=false
    if [[ "${1:-}" == "--compliance" ]]; then
        with_compliance=true
    fi

    echo "Scanning PDFs under ${SAMPLEDATA} with veraPDF parse check (${PARALLEL_JOBS} workers)..."
    if [[ "${with_compliance}" == true ]]; then
        echo "Also checking PDF/A-4 and PDF/UA-2 compliance for each PDF."
    fi
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
    local -a compliance_json_files=()

    for pdf in "${pdfs[@]}"; do
        wait_for_slot running "${PARALLEL_JOBS}"
        (
            set +e
            scan_one_pdf "${pdf}" > "${tmpdir}/${idx}.row"
            echo $? > "${tmpdir}/${idx}.status"
            if [[ "${with_compliance}" == true ]]; then
                scan_one_pdf_compliance "${pdf}" "${tmpdir}/compliance" > "${tmpdir}/${idx}.compliance"
            fi
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
        if [[ "${with_compliance}" == true && -f "${tmpdir}/${i}.compliance" ]]; then
            while IFS= read -r line; do
                if [[ "${line}" == COMPLIANCE_JSON:* ]]; then
                    local json_list="${line#COMPLIANCE_JSON:}"
                    local json_path
                    IFS=',' read -ra json_paths <<< "${json_list}"
                    for json_path in "${json_paths[@]}"; do
                        if [[ -f "${json_path}" ]]; then
                            compliance_json_files+=("${json_path}")
                        fi
                    done
                fi
            done < "${tmpdir}/${i}.compliance"
        fi
    done

    echo ""
    if [[ "${with_compliance}" == true && ${#compliance_json_files[@]} -gt 0 ]]; then
        python3 "${VERAPDF_REPORT}" table --sampledata "${SAMPLEDATA}/" "${compliance_json_files[@]}"
        echo ""
    fi
    echo "Summary: ${total} PDFs — ${valid} valid, ${invalid} invalid"
}

case "${1:-}" in
    --scan-all)
        scan_all
        ;;
    --scan-all-compliance)
        scan_all --compliance
        ;;
    --zerodha-only)
        zerodha_only
        ;;
    --help|-h)
        echo "Usage: $0 [--scan-all | --scan-all-compliance | --zerodha-only]"
        echo "Environment: VERIFY_PDFS_JOBS, VERAPDF_BIN, AVALPDF_BIN,"
        echo "             VERIFY_STRUCTURE_TREE, VERIFY_AVALPDF, VERIFY_AVALPDF_STRICT, NO_COLOR=1"
        ;;
    *)
        post_test
        ;;
esac