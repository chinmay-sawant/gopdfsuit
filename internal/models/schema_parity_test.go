package models_test

// Models-to-gopdflib field-parity test (Phase 2.3). The owned-type split
// stays: internal models and public gopdflib types are distinct Go types.
// This test pins the wire contract between them for the handler-routed
// inputs (SplitSpec, CompressOptions, HTML requests, redaction options) so
// drift fails CI. JSON tag comparison ignores ",omitempty": only the wire
// name must match.

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/compress"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/merge"
	"github.com/chinmay-sawant/gopdfsuit/v7/pkg/gopdflib"
)

// jsonNames returns the sorted base JSON names of a struct type's fields,
// skipping untagged fields (internal-only state such as precomputed hints).
func jsonNames(t *testing.T, v any) []string {
	t.Helper()
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	var names []string
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// fieldNames returns the sorted lower-cased Go field names of a struct type,
// for internal engine types that carry no JSON tags (form-parsed inputs).
func fieldNames(t *testing.T, v any) []string {
	t.Helper()
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	var names []string
	for i := 0; i < typ.NumField(); i++ {
		if !typ.Field(i).IsExported() {
			continue
		}
		names = append(names, strings.ToLower(typ.Field(i).Name))
	}
	sort.Strings(names)
	return names
}

func assertJSONParity(t *testing.T, label string, internal, public any) {
	t.Helper()
	a, b := jsonNames(t, internal), jsonNames(t, public)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("%s JSON parity drift:\n internal=%v\n public=%v", label, a, b)
	}
}

func TestHandlerInputJSONParity(t *testing.T) {
	assertJSONParity(t, "HTMLToPDFRequest", models.HTMLToPDFRequest{}, gopdflib.HTMLToPDFRequest{})
	assertJSONParity(t, "HTMLToImageRequest", models.HTMLToImageRequest{}, gopdflib.HTMLToImageRequest{})
	assertJSONParity(t, "ApplyRedactionOptions", models.ApplyRedactionOptions{}, gopdflib.ApplyRedactionOptions{})
	assertJSONParity(t, "RedactionRect", models.RedactionRect{}, gopdflib.RedactionRect{})
	assertJSONParity(t, "RedactionTextQuery", models.RedactionTextQuery{}, gopdflib.RedactionTextQuery{})
	assertJSONParity(t, "OCRSettings", models.OCRSettings{}, gopdflib.OCRSettings{})
	assertJSONParity(t, "PageInfo", models.PageInfo{}, gopdflib.PageInfo{})
	assertJSONParity(t, "PageDetail", models.PageDetail{}, gopdflib.PageDetail{})
	assertJSONParity(t, "TextPosition", models.TextPosition{}, gopdflib.TextPosition{})
	assertJSONParity(t, "PageCapability", models.PageCapability{}, gopdflib.PageCapability{})
	assertJSONParity(t, "RedactionApplyReport", models.RedactionApplyReport{}, gopdflib.RedactionApplyReport{})
}

// TestEngineOptionFieldParity pins the untagged engine option structs
// against the public JSON surface: every public wire name must match an
// engine field name case-insensitively, so renames on either side fail
// loudly instead of silently dropping options in the adapter.
func TestEngineOptionFieldParity(t *testing.T) {
	for _, tc := range []struct {
		label    string
		engine   any
		wireTags []string
	}{
		{"SplitSpec", merge.SplitSpec{}, jsonNames(t, gopdflib.SplitSpec{})},
		{"CompressOptions", compress.Options{}, jsonNames(t, gopdflib.CompressOptions{})},
	} {
		t.Run(tc.label, func(t *testing.T) {
			fields := fieldNames(t, tc.engine)
			have := make(map[string]bool, len(fields))
			for _, f := range fields {
				have[f] = true
			}
			for _, tag := range tc.wireTags {
				if !have[strings.ToLower(tag)] {
					t.Fatalf("%s: wire name %q has no engine field (fields=%v)", tc.label, tag, fields)
				}
			}
			if len(fields) != len(tc.wireTags) {
				t.Fatalf("%s: field count drift: engine=%v wire=%v", tc.label, fields, tc.wireTags)
			}
		})
	}
}
