package handlers

import (
	"io"
	"reflect"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
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
	hftBodyBufPool = sync.Pool{
		New: func() any {
			b := make([]byte, 0, 2<<20)
			return &b
		},
	}
)

// WarmJSONDecode pre-compiles the PDFTemplate JSON schema at process start.
func WarmJSONDecode() {
	jsonPretouchOnce.Do(func() {
		_ = sonic.Pretouch(reflect.TypeOf(models.PDFTemplate{}))
		_ = decoder.NewStreamDecoder(io.NopCloser(nil))
		var warm models.PDFTemplate
		warm.PreallocForDecode(0, "hft")
		_ = decodeHFTPayload([]byte(`{"config":{"page":"A4","pageAlignment":1},"title":{"props":"a","text":"b"},"elements":[{"type":"table","table":{"maxcolumns":1,"rows":[{"row":[{"props":"a","text":"b"}]}]}}],"footer":{"font":"a","text":"b"}}`), &warm)
	})
}

func decodeTemplateJSON(r io.Reader, contentLength int, tier string, template *models.PDFTemplate) error {
	// HFT (~5% traffic): read once + NoCopy unmarshal when Content-Length known.
	// Only ~2–3 concurrent HFT requests at 48 VUs — bounded heap vs StreamDecoder CPU.
	if tier == "hft" {
		if contentLength > 0 && contentLength <= maxHFTEncodeBody {
			bufPtr := hftBodyBufPool.Get().(*[]byte)
			buf := *bufPtr
			if cap(buf) < contentLength {
				buf = make([]byte, contentLength)
			} else {
				buf = buf[:contentLength]
			}
			if _, err := io.ReadFull(r, buf); err != nil {
				putHftBodyBuf(bufPtr, buf)
				return err
			}
			err := decodeHFTPayload(buf, template)
			putHftBodyBuf(bufPtr, buf)
			return err
		}
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

func putHftBodyBuf(bufPtr *[]byte, buf []byte) {
	if cap(buf) > maxHFTEncodeBody {
		*bufPtr = make([]byte, 0, 2<<20)
	} else {
		*bufPtr = buf[:0]
	}
	hftBodyBufPool.Put(bufPtr)
}
