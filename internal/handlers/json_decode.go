package handlers

import (
	"io"
	"reflect"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

const (
	maxPooledDecodeBody = 512 << 10 // 512 KiB — retail/active payloads only
	maxHFTEncodeBody    = 8 << 20   // 8 MiB — stream-decode fallback limit
)

var (
	jsonPretouchOnce sync.Once
	bodyBufPool      = sync.Pool{
		New: func() any {
			b := make([]byte, 0, 64*1024)
			return &b
		},
	}
)

// WarmJSONDecode pre-compiles the PDFTemplate JSON schema at process start.
func WarmJSONDecode() {
	jsonPretouchOnce.Do(func() {
		_ = sonic.Pretouch(reflect.TypeOf(models.PDFTemplate{}))
		_ = decoder.NewStreamDecoder(io.NopCloser(nil))
	})
}

func decodeTemplateJSON(r io.Reader, contentLength int, tier string, template *models.PDFTemplate) error {
	// HFT bodies are multi-MiB JSON; stream-decode avoids 48× large pooled buffers in flight.
	if tier == "hft" {
		return decoder.NewStreamDecoder(r).Decode(template)
	}
	if contentLength > 0 && contentLength <= maxPooledDecodeBody {
		bufPtr := bodyBufPool.Get().(*[]byte)
		buf := *bufPtr
		if cap(buf) < contentLength {
			buf = make([]byte, contentLength)
		} else {
			buf = buf[:contentLength]
		}
		if _, err := io.ReadFull(r, buf); err != nil {
			putBodyBuf(bufPtr, buf)
			return err
		}
		err := sonic.Unmarshal(buf, template)
		putBodyBuf(bufPtr, buf)
		return err
	}
	if contentLength > maxPooledDecodeBody && contentLength <= maxHFTEncodeBody {
		return decoder.NewStreamDecoder(r).Decode(template)
	}
	return decoder.NewStreamDecoder(r).Decode(template)
}

func putBodyBuf(bufPtr *[]byte, buf []byte) {
	// Never return large backing arrays to the pool — keeps heap bounded under 48 workers.
	if cap(buf) > 128<<10 {
		*bufPtr = make([]byte, 0, 64*1024)
	} else {
		*bufPtr = buf[:0]
	}
	bodyBufPool.Put(bufPtr)
}