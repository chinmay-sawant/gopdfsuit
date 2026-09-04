package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Compliance manifest entries shared with test/verify_pdfs.sh.
// See test/compliance_manifest.json.
type complianceManifestEntry struct {
	Path       string   `json:"path"`
	Baseline   string   `json:"baseline"`
	Tolerance  int      `json:"tolerance"`
	SkipSize   bool     `json:"skipSize"`
	Flavours   []string `json:"flavours"`
	AvalStrict bool     `json:"avalStrict"`
	Media      string   `json:"media"`
	Suite      string   `json:"suite"`
}

type complianceManifest struct {
	Entries []complianceManifestEntry `json:"entries"`
}

// zerodhaManifestEntries returns suite=="zerodha" entries from the shared
// compliance manifest, replacing the hardcoded list that used to mirror
// test/verify_pdfs.sh zerodha_compliance_entries().
func zerodhaManifestEntries(t *testing.T, root string) []complianceManifestEntry {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "test", "compliance_manifest.json"))
	if err != nil {
		t.Fatalf("read compliance manifest: %v", err)
	}
	var manifest complianceManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parse compliance manifest: %v", err)
	}
	var out []complianceManifestEntry
	for _, entry := range manifest.Entries {
		if entry.Suite == "zerodha" {
			out = append(out, entry)
		}
	}
	if len(out) == 0 {
		t.Fatal("compliance manifest has no suite==zerodha entries")
	}
	return out
}

var zerodhaComplianceFlavourLabels = map[string]string{
	"4":   "PDF/A-4",
	"ua2": "PDF/UA-2",
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
// Zerodha benchmark outputs listed in test/compliance_manifest.json.
// Skips when veraPDF is not installed.
func TestZerodhaPDFCompliance(t *testing.T) {
	root := repoRoot(t)
	verapdf := verapdfBin(t, root)
	report := filepath.Join(root, "test", "verapdf_report.py")
	sampledata := filepath.Join(root, "sampledata")

	for _, entry := range zerodhaManifestEntries(t, root) {
		entry := entry
		pdf := samplePath(t, entry.Path)
		if _, err := os.Stat(pdf); err != nil {
			t.Fatalf("missing baseline PDF %s: %v", entry.Path, err)
		}

		for _, flavour := range entry.Flavours {
			flavour := flavour
			label := zerodhaComplianceFlavourLabels[flavour]
			if label == "" {
				label = flavour
			}
			t.Run(entry.Path+"/"+label, func(t *testing.T) {
				t.Parallel()
				cmd := exec.Command(
					"python3", report, "check",
					"--verapdf", verapdf,
					"--pdf", pdf,
					"--flavour", flavour,
					"--sampledata", sampledata+string(filepath.Separator),
					"--no-color",
				)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("veraPDF %s (%s) failed:\n%s", entry.Path, label, string(out))
				}
			})
		}
	}
}
