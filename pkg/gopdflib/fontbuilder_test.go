package gopdflib

import (
	"reflect"
	"testing"
)

func TestFontBuilderRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		chain string
		want  string
	}{
		{
			name:  "defaults render Helvetica 12 regular left no borders",
			chain: Font("Helvetica").Props(),
			want:  MakeProps("Helvetica", 0, false, false, false, "left", [4]int{0, 0, 0, 0}),
		},
		{
			name:  "empty name falls back to Helvetica",
			chain: Font("").Props(),
			want:  MakeProps("", 12, false, false, false, "left", [4]int{0, 0, 0, 0}),
		},
		{
			name:  "size and bold center bordered",
			chain: Font("Helvetica").Size(12).Bold().Center().Bordered().Props(),
			want:  MakeProps("Helvetica", 12, true, false, false, "center", [4]int{1, 1, 1, 1}),
		},
		{
			name:  "italic underline right custom borders",
			chain: Font("Times").Size(10).Italic().Underline().Right().Borders(1, 0, 1, 0).Props(),
			want:  MakeProps("Times", 10, false, true, true, "right", [4]int{1, 0, 1, 0}),
		},
		{
			name:  "all styles left borderless",
			chain: Font("Courier").Size(14).Bold().Italic().Underline().Left().Borderless().Props(),
			want:  MakeProps("Courier", 14, true, true, true, "left", [4]int{0, 0, 0, 0}),
		},
		{
			name:  "non-positive size renders as 12",
			chain: Font("Helvetica").Size(-3).Props(),
			want:  MakeProps("Helvetica", 12, false, false, false, "left", [4]int{0, 0, 0, 0}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.chain != tc.want {
				t.Fatalf("chain = %q, want MakeProps = %q", tc.chain, tc.want)
			}
		})
	}
}

func TestFontBuilderCellTerminal(t *testing.T) {
	got := Font("Helvetica").Size(12).Bold().Center().Bordered().Cell("hi")
	want := NewCell("hi", MakeProps("Helvetica", 12, true, false, false, "center", [4]int{1, 1, 1, 1}))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Cell() = %+v, want %+v", got, want)
	}
}

func TestFluentCellBuilder(t *testing.T) {
	fontProps := MakeProps("Helvetica", 12, true, false, false, "center", [4]int{1, 1, 1, 1})

	t.Run("WithFont matches NewCell with MakeProps", func(t *testing.T) {
		got := Text("hi").WithFont(Font("Helvetica").Size(12).Bold().Center().Bordered()).Build()
		want := NewCell("hi", fontProps)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Build() = %+v, want %+v", got, want)
		}
	})

	t.Run("Cell alias matches Build", func(t *testing.T) {
		b := Text("hi").WithFont(Font("Helvetica").Size(12).Bold().Center().Bordered())
		if !reflect.DeepEqual(b.Cell(), b.Build()) {
			t.Fatalf("Cell() = %+v, Build() = %+v", b.Cell(), b.Build())
		}
	})

	t.Run("defaults are Helvetica 12 regular left no borders", func(t *testing.T) {
		got := Text("hi").Build()
		want := NewCell("hi", MakeProps("Helvetica", 0, false, false, false, "left", [4]int{0, 0, 0, 0}))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Build() = %+v, want %+v", got, want)
		}
	})

	t.Run("explicit Props wins over WithFont", func(t *testing.T) {
		got := Text("hi").WithFont(Font("Helvetica").Size(12).Bold()).Props(fontProps).Build()
		if got.Props != fontProps {
			t.Fatalf("Props = %q, want %q", got.Props, fontProps)
		}
	})

	t.Run("Bg Fg Math", func(t *testing.T) {
		got := Text("x^2").Bg("#F5F5F5").Fg("#B00020").Math().Build()
		if got.BgColor != "#F5F5F5" || got.TextColor != "#B00020" {
			t.Fatalf("colors = (%q, %q), want (#F5F5F5, #B00020)", got.BgColor, got.TextColor)
		}
		if got.MathEnabled == nil || !*got.MathEnabled {
			t.Fatalf("MathEnabled = %+v, want pointer to true", got.MathEnabled)
		}
	})
}

func TestFontBuilderNilSafety(t *testing.T) {
	var fb *FontBuilder
	if got := fb.Size(12).Bold().Italic().Underline().Center().Borders(1, 1, 1, 1).Props(); got != "" {
		t.Fatalf("nil chain Props() = %q, want \"\"", got)
	}
	if got := fb.Borderless().Props(); got != "" {
		t.Fatalf("nil Borderless Props() = %q, want \"\"", got)
	}
	if got := fb.Left().Right().Bordered().Props(); got != "" {
		t.Fatalf("nil align Props() = %q, want \"\"", got)
	}
	if got := fb.Cell("hi"); got.Text != "hi" || got.Props != "" {
		t.Fatalf("nil Cell() = %+v, want text only", got)
	}

	var cb *CellBuilder
	if got := cb.WithFont(Font("Helvetica")).Props("p").Bg("bg").Fg("fg").Math().Build(); !reflect.DeepEqual(got, Cell{}) {
		t.Fatalf("nil Build() = %+v, want zero Cell", got)
	}
	if got := cb.Cell(); !reflect.DeepEqual(got, Cell{}) {
		t.Fatalf("nil Cell() = %+v, want zero Cell", got)
	}

	t.Run("nil WithFont arg no-ops", func(t *testing.T) {
		got := Text("hi").WithFont(nil).Build()
		want := Text("hi").Build()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("WithFont(nil) = %+v, want %+v", got, want)
		}
	})
}
