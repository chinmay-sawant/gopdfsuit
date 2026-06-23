package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Committed Zerodha benchmark artifacts validated for PDF/A-4 and PDF/UA-2.
// Mirrors test/verify_pdfs.sh zerodha_compliance_entries().
var zerodhaCompliancePDFs = []string{
	"gopdflib/zerodha/zerodha_hft_output.pdf",
	"gopdflib/zerodha/zerodha_retail_output.pdf",
	"gopdflib/zerodha/zerodha_active_output.pdf",
}

var zerodhaComplianceFlavours = []struct {
	code  string
	label string
}{
	{"4", "PDF/A-4"},
	{"ua2", "PDF/UA-2"},
}

func verapdfBin(t *testing.T, root string) string {
	t.Helper()
	bin := os.Getenv("VERAPDF_BIN")
	if bin == "" {
		bin = filepath.Join(root, "verapdf", "verapdf")
	}
	info, err := os.Stat(bin)
	if err != nil || info.Mode()&0111 == 0 {
		t.Skipf("veraPDF not installed at %s (run: make install-verapdf)", bin)
	}
	return bin
}

// TestZerodhaPDFCompliance runs veraPDF PDF/A-4 and PDF/UA-2 checks against the
// three committed Zerodha benchmark outputs. Skips when veraPDF is not installed.
func TestZerodhaPDFCompliance(t *testing.T) {
	root := repoRoot(t)
	verapdf := verapdfBin(t, root)
	report := filepath.Join(root, "test", "verapdf_report.py")
	sampledata := filepath.Join(root, "sampledata")

	for _, rel := range zerodhaCompliancePDFs {
		rel := rel
		pdf := samplePath(t, rel)
		if _, err := os.Stat(pdf); err != nil {
			t.Fatalf("missing baseline PDF %s: %v", rel, err)
		}

		for _, flavour := range zerodhaComplianceFlavours {
			flavour := flavour
			t.Run(rel+"/"+flavour.label, func(t *testing.T) {
				t.Parallel()
				cmd := exec.Command(
					"python3", report, "check",
					"--verapdf", verapdf,
					"--pdf", pdf,
					"--flavour", flavour.code,
					"--sampledata", sampledata+string(filepath.Separator),
					"--no-color",
				)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("veraPDF %s (%s) failed:\n%s", rel, flavour.label, string(out))
				}
			})
		}
	}
}
