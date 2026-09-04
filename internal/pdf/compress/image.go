package compress

import (
	"bytes"
	"image"
	"image/jpeg"
	"math"
)

func compressImage(dict, data []byte, opts Options) ([]byte, bool) {
	if hasSMask(dict) {
		return nil, false
	}

	width := dictInt(dict, widthRe)
	height := dictInt(dict, heightRe)
	if !imagePixelBudgetOK(width, height) {
		return nil, false
	}

	bits := dictInt(dict, bitsRe)
	if bits == 0 {
		bits = 8
	}
	if bits != 8 {
		return nil, false
	}

	cs := dictNameAfter(dict, colorSpaceRe)
	filter := streamFilter(dict)

	img, ok := decodeImagePixels(filter, cs, width, height, data, dict)
	if !ok {
		return nil, false
	}

	img = downscaleBicubic(img, opts.MaxImageDim)
	bounds := img.Bounds()
	newW, newH := bounds.Dx(), bounds.Dy()

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.JPEGQuality}); err != nil {
		return nil, false
	}
	jpegData := buf.Bytes()
	if len(jpegData) == 0 {
		return nil, false
	}

	outDict := bytes.Clone(dict)
	outDict = setNameInt(outDict, "Width", newW)
	outDict = setNameInt(outDict, "Height", newH)
	outDict = setFilter(outDict, filterDCT)
	outDict = removeNamedValue(outDict, "DecodeParms")
	outDict = removeNamedValue(outDict, "DP")
	if _, isGray := img.(*image.Gray); isGray {
		outDict = setColorSpace(outDict, "DeviceGray")
	} else {
		outDict = setColorSpace(outDict, "DeviceRGB")
	}

	return buildStream(outDict, jpegData), true
}

func decodeImagePixels(filter, colorSpace string, width, height int, data, dict []byte) (image.Image, bool) {
	switch filter {
	case filterDCT:
		cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
		if err != nil || !imagePixelBudgetOK(cfg.Width, cfg.Height) {
			return nil, false
		}
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, false
		}
		return img, true
	case filterFlate:
		if hasDecodeParms(dict) {
			return nil, false
		}
		raw, err := decompressFlate(data)
		if err != nil {
			return nil, false
		}
		return rawToImage(raw, colorSpace, width, height)
	case "":
		return rawToImage(data, colorSpace, width, height)
	default:
		return nil, false
	}
}

func imagePixelBudgetOK(width, height int) bool {
	if width <= 0 || height <= 0 || width > MaxImageEdge || height > MaxImageEdge {
		return false
	}
	return int64(width)*int64(height) <= MaxImagePixels
}

func rawToImage(raw []byte, colorSpace string, width, height int) (image.Image, bool) {
	if !imagePixelBudgetOK(width, height) {
		return nil, false
	}
	switch colorSpace {
	case "DeviceGray":
		if len(raw) < width*height {
			return nil, false
		}
		img := image.NewGray(image.Rect(0, 0, width, height))
		copy(img.Pix, raw[:width*height])
		return img, true
	case "DeviceRGB", "":
		if len(raw) < width*height*3 {
			return nil, false
		}
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		src := raw
		di := 0
		for i := 0; i+2 < len(src) && di+3 < len(img.Pix); i += 3 {
			img.Pix[di] = src[i]
			img.Pix[di+1] = src[i+1]
			img.Pix[di+2] = src[i+2]
			img.Pix[di+3] = 255
			di += 4
		}
		return img, true
	default:
		return nil, false
	}
}

func downscaleBicubic(src image.Image, maxDim int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if maxDim <= 0 || (w <= maxDim && h <= maxDim) {
		return src
	}
	scale := float64(maxDim) / float64(w)
	if h > w {
		scale = float64(maxDim) / float64(h)
	}
	nw := int(float64(w)*scale + 0.5)
	nh := int(float64(h)*scale + 0.5)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	return resizeBicubic(src, nw, nh)
}

func resizeBicubic(src image.Image, nw, nh int) image.Image {
	b := src.Bounds()
	sw, sh := float64(b.Dx()), float64(b.Dy())
	_, isGray := src.(*image.Gray)
	gray := (*image.Gray)(nil)
	rgba := (*image.RGBA)(nil)
	if isGray {
		gray = image.NewGray(image.Rect(0, 0, nw, nh))
	} else {
		rgba = image.NewRGBA(image.Rect(0, 0, nw, nh))
	}
	for y := 0; y < nh; y++ {
		sy := (float64(y)+0.5)*sh/float64(nh) - 0.5 + float64(b.Min.Y)
		for x := 0; x < nw; x++ {
			sx := (float64(x)+0.5)*sw/float64(nw) - 0.5 + float64(b.Min.X)
			r, g, bl, a := sampleBicubic(src, sx, sy)
			if isGray {
				yv := (19595*uint32(r) + 38470*uint32(g) + 7471*uint32(bl) + 1<<15) >> 16
				gray.Pix[y*gray.Stride+x] = uint8(yv)
			} else {
				i := y*rgba.Stride + x*4
				rgba.Pix[i] = r
				rgba.Pix[i+1] = g
				rgba.Pix[i+2] = bl
				rgba.Pix[i+3] = a
			}
		}
	}
	if isGray {
		return gray
	}
	return rgba
}

func cubicWeight(t float64) float64 {
	const a = -0.5
	if t < 0 {
		t = -t
	}
	t2 := t * t
	t3 := t2 * t
	if t < 1 {
		return (a+2)*t3 - (a+3)*t2 + 1
	}
	if t < 2 {
		return a*t3 - 5*a*t2 + 8*a*t - 4*a
	}
	return 0
}

func sampleBicubic(src image.Image, x, y float64) (r, g, b, a uint8) {
	bnds := src.Bounds()
	ix := math.Floor(x)
	iy := math.Floor(y)
	var sr, sg, sb, sa float64
	for j := -1; j <= 2; j++ {
		wy := cubicWeight(y - (iy + float64(j)))
		py := clampInt(int(iy)+j, bnds.Min.Y, bnds.Max.Y-1)
		for i := -1; i <= 2; i++ {
			w := wy * cubicWeight(x-(ix+float64(i)))
			px := clampInt(int(ix)+i, bnds.Min.X, bnds.Max.X-1)
			cr, cg, cb, ca := src.At(px, py).RGBA()
			sr += float64(cr>>8) * w
			sg += float64(cg>>8) * w
			sb += float64(cb>>8) * w
			sa += float64(ca>>8) * w
		}
	}
	return clamp8(sr), clamp8(sg), clamp8(sb), clamp8(sa)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clamp8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}

func setColorSpace(dict []byte, name string) []byte {
	repl := []byte("/ColorSpace /" + name)
	if colorSpaceRe.Match(dict) {
		return colorSpaceRe.ReplaceAll(dict, repl)
	}
	return insertBeforeDictEnd(dict, repl)
}
