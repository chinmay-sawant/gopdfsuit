package compress

const (
	// MaxInputBytes is the largest PDF CompressPDF will accept (library, HTTP, WASM).
	MaxInputBytes = 32 << 20 // 32 MiB
	// MaxInflateBytes caps a single Flate stream after decompression (zip bomb).
	MaxInflateBytes = 48 << 20 // 48 MiB
	// MaxImagePixels is width*height before an image is decoded or rasterized.
	MaxImagePixels = 16_000_000
	// MaxImageEdge is the larger of Width/Height accepted on an image XObject.
	MaxImageEdge = 8192
	// MaxObjects is the max PDF object number and object count after parse.
	MaxObjects = 50_000
	// maxAllowedDim caps Options.MaxImageDim (HTTP override cannot request a 1e9 downsample target).
	maxAllowedDim = 4096
)
