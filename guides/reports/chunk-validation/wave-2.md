# Wave 2 — Findings 001–236 Validation (Manual Review)

> **Scan source:** `/home/chinmay/ChinmayPersonalProjects/slopguard/scripts/chunks` (10 chunks, 236 findings)
> **Method:** Manual review per `CHUNK_VALIDATOR.md` — pattern match only, no project context
> **Reviewed:** 2026-07-01 (synced with latest chunk scan)

## Summary

- [x] Chunk **1–25** | Range: **1–25** | TP: **6** | FP: **19** | FP rate: **76.0%**
- [x] Chunk **26–50** | Range: **26–50** | TP: **7** | FP: **18** | FP rate: **72.0%**
- [x] Chunk **51–75** | Range: **51–75** | TP: **15** | FP: **10** | FP rate: **40.0%**
- [x] Chunk **76–100** | Range: **76–100** | TP: **17** | FP: **8** | FP rate: **32.0%**
- [x] Chunk **101–125** | Range: **101–125** | TP: **22** | FP: **3** | FP rate: **12.0%**
- [x] Chunk **126–150** | Range: **126–150** | TP: **14** | FP: **11** | FP rate: **44.0%**
- [x] Chunk **151–175** | Range: **151–175** | TP: **17** | FP: **8** | FP rate: **32.0%**
- [x] Chunk **176–200** | Range: **176–200** | TP: **1** | FP: **24** | FP rate: **96.0%**
- [x] Chunk **201–225** | Range: **201–225** | TP: **6** | FP: **19** | FP rate: **76.0%**
- [x] Chunk **226–236** | Range: **226–236** | TP: **0** | FP: **11** | FP rate: **100.0%**
- [x] **Total** | Range: **1–236** | TP: **105** | FP: **131** | FP rate: **55.5%**

## Per-Finding Checklist

- [x] **Finding 1** | Rule: **PERF-151** | Correctly Fired: **No** | main() is one-time startup @ `cmd/gopdfsuit/main.go:20:1` → N/A: FP — not request hot path
- [x] **Finding 2** | Rule: **PERF-41** | Correctly Fired: **No** | log in opt-in profiling branch @ `cmd/gopdfsuit/main.go:25:4` → N/A: FP — not production request path
- [x] **Finding 3** | Rule: **PERF-43** | Correctly Fired: **Yes** | defer recover in per-request middleware @ `cmd/gopdfsuit/main.go:55:3` → Fixed
- [x] **Finding 4** | Rule: **PERF-68** | Correctly Fired: **No** | gin.Logger gated behind DebugMode @ `cmd/gopdfsuit/main.go:66:14` → N/A: FP — not enabled in release mode
- [x] **Finding 5** | Rule: **CWE-497** | Correctly Fired: **No** | anchor is commented NumCPU line @ `cmd/gopdfsuit/main.go:72:22` → N/A: FP — no diagnostics endpoint
- [x] **Finding 6** | Rule: **PERF-148** | Correctly Fired: **No** | buffered semaphore with defer receive @ `cmd/gopdfsuit/main.go:74:15` → N/A: FP — not unbuffered leak
- [x] **Finding 7** | Rule: **BP-9** | Correctly Fired: **No** | intentional graceful-shutdown select @ `cmd/gopdfsuit/main.go:102:2` → N/A: FP — waits for signal by design
- [x] **Finding 8** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Errorf on CLI benchmark error path @ `internal/benchmarktemplates/runner.go:16:10` → N/A: FP — cold path
- [x] **Finding 9** | Rule: **PERF-40** | Correctly Fired: **Yes** | multiple time.Now per benchmark iteration @ `internal/benchmarktemplates/runner.go:25:16` → Fixed
- [x] **Finding 10** | Rule: **BP-1** | Correctly Fired: **No** | _ discards runtime.Caller ok bool @ `internal/benchmarktemplates/zerodha_retail.go:22:2` → N/A: FP — not error discard
- [x] **Finding 11** | Rule: **PERF-35** | Correctly Fired: **No** | BenchmarkHeader one-off fmt.Sprintf @ `internal/benchmarktemplates/zerodha_retail.go:235:9` → N/A: FP — not hot path
- [x] **Finding 12** | Rule: **PERF-151** | Correctly Fired: **No** | getProjectRoot is cold startup helper @ `internal/handlers/handlers.go:33:1` → N/A: FP — not request handler
- [x] **Finding 13** | Rule: **PERF-200** | Correctly Fired: **No** | CORS registered before Auth @ `internal/handlers/handlers.go:108:20` → N/A: FP — ordering correct
- [x] **Finding 14** | Rule: **CWE-22** | Correctly Fired: **No** | filepath.Base confines user input @ `internal/handlers/handlers.go:207:2` → N/A: FP — traversal prevented
- [x] **Finding 15** | Rule: **PERF-22** | Correctly Fired: **Yes** | os.ReadFile on cache miss @ `internal/handlers/handlers.go:215:15` → Fixed
- [x] **Finding 16** | Rule: **PERF-48** | Correctly Fired: **No** | EqualFold on 4-char extensions @ `internal/handlers/handlers.go:255:6` → N/A: FP — short string compare
- [x] **Finding 17** | Rule: **PERF-41** | Correctly Fired: **Yes** | log.Printf in handler defer @ `internal/handlers/handlers.go:268:4` → Fixed
- [x] **Finding 18** | Rule: **PERF-57** | Correctly Fired: **No** | io.ReadAll in route handler not middleware @ `internal/handlers/handlers.go:272:15` → N/A: FP — handler body
- [x] **Finding 19** | Rule: **BP-1** | Correctly Fired: **No** | CutSuffix discards found bool @ `internal/handlers/handlers.go:279:2` → N/A: FP — not error discard
- [x] **Finding 20** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte(b) form-field fallback @ `internal/handlers/handlers.go:358:15` → Fixed
- [x] **Finding 21** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte(b) form-field fallback @ `internal/handlers/handlers.go:363:16` → Fixed
- [x] **Finding 22** | Rule: **PERF-56** | Correctly Fired: **No** | c.JSON only on per-file error with return @ `internal/handlers/handlers.go:405:4` → N/A: FP — error-only branch
- [x] **Finding 23** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Sprintf on rare open failure @ `internal/handlers/handlers.go:405:58` → N/A: FP — error path only
- [x] **Finding 24** | Rule: **PERF-6** | Correctly Fired: **No** | fmt.Sprintf on rare read failure @ `internal/handlers/handlers.go:405:58` → N/A: FP — error path only
- [x] **Finding 25** | Rule: **PERF-6** | Correctly Fired: **No** | fmt.Sprintf on rare read failure @ `internal/handlers/handlers.go:413:58` → N/A: FP — error path only
- [x] **Finding 26** | Rule: **PERF-109** | Correctly Fired: **No** | byte-index trim loop not map keys @ `internal/handlers/redact.go:18:2` → N/A: FP — wrong rule
- [x] **Finding 27** | Rule: **PERF-41** | Correctly Fired: **Yes** | log.Printf on upload close @ `internal/handlers/redact.go:32:3` → Fixed
- [x] **Finding 28** | Rule: **PERF-119** | Correctly Fired: **No** | AppendInt separates append chain @ `internal/pdf/bookmarks.go:36:7` → N/A: FP — cannot merge across AppendInt
- [x] **Finding 29** | Rule: **PERF-119** | Correctly Fired: **No** | already variadic append of two bytes @ `internal/pdf/draw.go:35:10` → N/A: FP — already optimal
- [x] **Finding 30** | Rule: **PERF-128** | Correctly Fired: **No** | appendFmtNum breaks consecutive appends @ `internal/pdf/draw.go:90:10` → N/A: FP — interleaved calls
- [x] **Finding 31** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor alpha discard @ `internal/pdf/draw.go:213:5` → N/A: FP — not error discard
- [x] **Finding 32** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor alpha discard @ `internal/pdf/draw.go:249:5` → N/A: FP — not error discard
- [x] **Finding 33** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor alpha discard @ `internal/pdf/draw.go:391:7` → N/A: FP — not error discard
- [x] **Finding 34** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor alpha discard @ `internal/pdf/draw.go:566:8` → N/A: FP — not error discard
- [x] **Finding 35** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor alpha discard @ `internal/pdf/draw.go:902:7` → N/A: FP — not error discard
- [x] **Finding 36** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor alpha discard @ `internal/pdf/draw.go:1084:8` → N/A: FP — not error discard
- [x] **Finding 37** | Rule: **BP-1** | Correctly Fired: **No** | parseHexColor alpha discard @ `internal/pdf/draw.go:1140:8` → N/A: FP — not error discard
- [x] **Finding 38** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf for widget /Rect @ `internal/pdf/draw.go:1634:10` → Fixed
- [x] **Finding 39** | Rule: **PERF-42** | Correctly Fired: **No** | static fmt.Errorf in constructor @ `internal/pdf/encryption/encrypt.go:38:15` → N/A: FP — one-time setup
- [x] **Finding 40** | Rule: **PERF-32** | Correctly Fired: **No** | []byte(password) in encryption init @ `internal/pdf/encryption/encrypt.go:62:9` → N/A: FP — one-time setup
- [x] **Finding 41** | Rule: **CWE-916** | Correctly Fired: **Yes** | MD5 in PDF Standard Security Handler @ `internal/pdf/encryption/encrypt.go:79:10` → Fixed
- [x] **Finding 42** | Rule: **CWE-328** | Correctly Fired: **Yes** | MD5 owner hash in PDF encryption @ `internal/pdf/encryption/encrypt.go:79:10` → Fixed
- [x] **Finding 43** | Rule: **PERF-53** | Correctly Fired: **No** | uses crypto/rand not math/rand @ `internal/pdf/encryption/encrypt.go:246:15` → N/A: FP — wrong RNG flagged
- [x] **Finding 44** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Sprintf once per encrypted doc @ `internal/pdf/encryption/encrypt.go:321:19` → N/A: FP — not per-request
- [x] **Finding 45** | Rule: **PERF-110** | Correctly Fired: **No** | pool already returns *zlib.Writer @ `internal/pdf/font/compression.go:13:22` → N/A: FP — pointer type in pool
- [x] **Finding 46** | Rule: **BP-3** | Correctly Fired: **Yes** | panic in sync.Pool New @ `internal/pdf/font/compression.go:18:4` → Fixed
- [x] **Finding 47** | Rule: **BP-2** | Correctly Fired: **Yes** | bare return err in CloseZlibWriter @ `internal/pdf/font/compression.go:47:1` → Fixed
- [x] **Finding 48** | Rule: **PERF-109** | Correctly Fired: **No** | range over slice not map keys @ `internal/pdf/font/metrics.go:549:2` → N/A: FP — wrong rule
- [x] **Finding 49** | Rule: **PERF-35** | Correctly Fired: **No** | one-time Helvetica resource string @ `internal/pdf/font/metrics.go:582:9` → N/A: FP — not hot path
- [x] **Finding 50** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) without size hint @ `internal/pdf/font/metrics.go:652:13` → Fixed
- [x] **Finding 51** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded CloseZlibWriter on failure @ `internal/pdf/font/metrics.go:664:3` → Fixed
- [x] **Finding 52** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded CloseZlibWriter on failure @ `internal/pdf/font/metrics.go:910:3` → Fixed
- [x] **Finding 53** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded CloseZlibWriter on failure @ `internal/pdf/font/metrics.go:1017:3` → Fixed
- [x] **Finding 54** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Errorf on font bootstrap error @ `internal/pdf/font/pdfa.go:150:11` → N/A: FP — cold path
- [x] **Finding 55** | Rule: **PERF-42** | Correctly Fired: **No** | static fmt.Errorf in font retry @ `internal/pdf/font/pdfa.go:165:10` → N/A: FP — infrequent bootstrap
- [x] **Finding 56** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded os.Remove error @ `internal/pdf/font/pdfa.go:231:3` → Fixed
- [x] **Finding 57** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Errorf on font load failure @ `internal/pdf/font/registry.go:60:10` → N/A: FP — error path
- [x] **Finding 58** | Rule: **PERF-48** | Correctly Fired: **No** | EqualFold on 4-char extensions @ `internal/pdf/font/registry.go:374:7` → N/A: FP — short strings
- [x] **Finding 59** | Rule: **BP-1** | Correctly Fired: **No** | CutSuffix discards found bool @ `internal/pdf/font/registry.go:378:3` → N/A: FP — not error discard
- [x] **Finding 60** | Rule: **PERF-3** | Correctly Fired: **No** | conditional scratch grow not loop rebuild @ `internal/pdf/font/subset.go:113:6` → N/A: FP — guarded realloc
- [x] **Finding 61** | Rule: **PERF-32** | Correctly Fired: **No** | append for map ownership not str conversion @ `internal/pdf/font/subset.go:118:32` → N/A: FP — rule misapplied
- [x] **Finding 62** | Rule: **PERF-3** | Correctly Fired: **No** | conditional glyphScratch grow @ `internal/pdf/font/subset.go:297:5` → N/A: FP — intentional growth
- [x] **Finding 63** | Rule: **PERF-121** | Correctly Fired: **No** | uint32 branch not struct literal @ `internal/pdf/font/subset.go:537:23` → N/A: FP — wrong pattern
- [x] **Finding 64** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Errorf on ReadFile failure @ `internal/pdf/font/ttf.go:63:15` → N/A: FP — cold error path
- [x] **Finding 65** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read in TTF table loop @ `internal/pdf/font/ttf.go:115:13` → Fixed
- [x] **Finding 66** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read checksum in TTF loop @ `internal/pdf/font/ttf.go:118:13` → Fixed
- [x] **Finding 67** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read length in TTF loop @ `internal/pdf/font/ttf.go:121:13` → Fixed
- [x] **Finding 68** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read in hmtx loop @ `internal/pdf/font/ttf.go:289:13` → Fixed
- [x] **Finding 69** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read in cmap scan @ `internal/pdf/font/ttf.go:330:13` → Fixed
- [x] **Finding 70** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read encodingID in cmap @ `internal/pdf/font/ttf.go:333:13` → Fixed
- [x] **Finding 71** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read offset in cmap @ `internal/pdf/font/ttf.go:336:13` → Fixed
- [x] **Finding 72** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read format probe in cmap @ `internal/pdf/font/ttf.go:345:14` → Fixed
- [x] **Finding 73** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read Unicode format probe @ `internal/pdf/font/ttf.go:360:14` → Fixed
- [x] **Finding 74** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read format-4 endCode @ `internal/pdf/font/ttf.go:414:13` → Fixed
- [x] **Finding 75** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read format-4 startCode @ `internal/pdf/font/ttf.go:426:13` → Fixed
- [x] **Finding 76** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read format-4 idDelta @ `internal/pdf/font/ttf.go:434:13` → Fixed
- [x] **Finding 77** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read format-4 idRangeOffset @ `internal/pdf/font/ttf.go:446:13` → Fixed
- [x] **Finding 78** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read nested glyphID @ `internal/pdf/font/ttf.go:467:16` → Fixed
- [x] **Finding 79** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read format-12 startCharCode @ `internal/pdf/font/ttf.go:501:13` → Fixed
- [x] **Finding 80** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read format-12 endCharCode @ `internal/pdf/font/ttf.go:504:13` → Fixed
- [x] **Finding 81** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read format-12 startGlyphID @ `internal/pdf/font/ttf.go:507:13` → Fixed
- [x] **Finding 82** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read name-table platformID @ `internal/pdf/font/ttf.go:548:13` → Fixed
- [x] **Finding 83** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read name-table encodingID @ `internal/pdf/font/ttf.go:551:13` → Fixed
- [x] **Finding 84** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read name-table languageID @ `internal/pdf/font/ttf.go:554:13` → Fixed
- [x] **Finding 85** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read name-table nameID @ `internal/pdf/font/ttf.go:557:13` → Fixed
- [x] **Finding 86** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read name-table length @ `internal/pdf/font/ttf.go:560:13` → Fixed
- [x] **Finding 87** | Rule: **PERF-107** | Correctly Fired: **Yes** | binary.Read name-table offset @ `internal/pdf/font/ttf.go:563:13` → Fixed
- [x] **Finding 88** | Rule: **PERF-192** | Correctly Fired: **Yes** | make(map) without size hint @ `internal/pdf/font/ttf.go:759:14` → Fixed
- [x] **Finding 89** | Rule: **PERF-15** | Correctly Fired: **No** | strconv.Itoa in init() once @ `internal/pdf/form/xfdf.go:45:17` → N/A: FP — one-time precompute
- [x] **Finding 90** | Rule: **PERF-48** | Correctly Fired: **No** | EqualFold on 2-3 char literals @ `internal/pdf/form/xfdf.go:191:9` → N/A: FP — trivial compare
- [x] **Finding 91** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded r.Close after io.Copy @ `internal/pdf/form/xfdf.go:250:3` → Fixed
- [x] **Finding 92** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded flate reader Close @ `internal/pdf/form/xfdf.go:264:3` → Fixed
- [x] **Finding 93** | Rule: **PERF-186** | Correctly Fired: **No** | strings.Fields once per ObjStm header @ `internal/pdf/form/xfdf.go:655:16` → N/A: FP — infrequent parse path
- [x] **Finding 94** | Rule: **PERF-114** | Correctly Fired: **No** | map merge loop not slice copy @ `internal/pdf/form/xfdf.go:837:2` → N/A: FP — wrong rule
- [x] **Finding 95** | Rule: **PERF-109** | Correctly Fired: **No** | map range has cached keys @ `internal/pdf/form/xfdf.go:837:2` → N/A: FP — no key recomputation
- [x] **Finding 96** | Rule: **PERF-114** | Correctly Fired: **No** | map merge loop not slice copy @ `internal/pdf/form/xfdf.go:840:2` → N/A: FP — wrong rule
- [x] **Finding 97** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Sprintf on xml.Marshal failure @ `internal/pdf/form/xfdf.go:867:17` → N/A: FP — cold error path
- [x] **Finding 98** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte from Builder for radio groups @ `internal/pdf/form/xfdf.go:1026:12` → Fixed
- [x] **Finding 99** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte from Builder for /V replacement @ `internal/pdf/form/xfdf.go:1046:13` → Fixed
- [x] **Finding 100** | Rule: **BP-1** | Correctly Fired: **No** | _ discards dictStart int @ `internal/pdf/form/xfdf.go:1105:3` → N/A: FP — not error discard
- [x] **Finding 101** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa per AP object ref @ `internal/pdf/form/xfdf.go:1111:24` → Fixed
- [x] **Finding 102** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte from AP ref builder @ `internal/pdf/form/xfdf.go:1113:12` → Fixed
- [x] **Finding 103** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.FormatFloat per field font size @ `internal/pdf/form/xfdf.go:1148:25` → Fixed
- [x] **Finding 104** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.FormatFloat per field X @ `internal/pdf/form/xfdf.go:1150:25` → Fixed
- [x] **Finding 105** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.FormatFloat per field Y @ `internal/pdf/form/xfdf.go:1152:25` → Fixed
- [x] **Finding 106** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa per font descriptor @ `internal/pdf/form/xfdf.go:1160:27` → Fixed
- [x] **Finding 107** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa per font object @ `internal/pdf/form/xfdf.go:1168:26` → Fixed
- [x] **Finding 108** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa font descriptor xref @ `internal/pdf/form/xfdf.go:1174:26` → Fixed
- [x] **Finding 109** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte(streamBody) before zlib @ `internal/pdf/form/xfdf.go:1183:25` → Fixed
- [x] **Finding 110** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa per AP XObject @ `internal/pdf/form/xfdf.go:1192:24` → Fixed
- [x] **Finding 111** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.FormatFloat BBox width @ `internal/pdf/form/xfdf.go:1194:24` → Fixed
- [x] **Finding 112** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.FormatFloat BBox height @ `internal/pdf/form/xfdf.go:1196:24` → Fixed
- [x] **Finding 113** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa per AP resources @ `internal/pdf/form/xfdf.go:1198:24` → Fixed
- [x] **Finding 114** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa stream Length @ `internal/pdf/form/xfdf.go:1200:24` → Fixed
- [x] **Finding 115** | Rule: **PERF-3** | Correctly Fired: **Yes** | make([]byte,10) for xref padding @ `internal/pdf/form/xfdf.go:1231:5` → Fixed
- [x] **Finding 116** | Rule: **PERF-119** | Correctly Fired: **Yes** | two consecutive append of xref/trailer @ `internal/pdf/form/xfdf.go:1258:8` → Fixed
- [x] **Finding 117** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa ObjStm objNum @ `internal/pdf/form/xfdf.go:1525:29` → Fixed
- [x] **Finding 118** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa ObjStm offset @ `internal/pdf/form/xfdf.go:1527:29` → Fixed
- [x] **Finding 119** | Rule: **PERF-110** | Correctly Fired: **No** | pool returns *bytes.Buffer @ `internal/pdf/generator.go:24:21` → N/A: FP — already pointer type
- [x] **Finding 120** | Rule: **PERF-31** | Correctly Fired: **No** | defer pool Put in one-shot generator @ `internal/pdf/generator.go:85:2` → N/A: FP — idiomatic cleanup
- [x] **Finding 121** | Rule: **PERF-31** | Correctly Fired: **Yes** | defer pool cleanup in generator @ `internal/pdf/generator.go:90:2` → Fixed
- [x] **Finding 122** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte from title string concat @ `internal/pdf/generator.go:266:50` → Fixed
- [x] **Finding 123** | Rule: **PERF-109** | Correctly Fired: **No** | loop over widget IDs not map keys @ `internal/pdf/generator.go:501:3` → N/A: FP — no map key computation
- [x] **Finding 124** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf AcroForm dict @ `internal/pdf/generator.go:516:22` → Fixed
- [x] **Finding 125** | Rule: **PERF-119** | Correctly Fired: **Yes** | append chain in annotation loop @ `internal/pdf/generator.go:680:16` → Fixed
- [x] **Finding 126** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded closeZlibWriter on failure @ `internal/pdf/generator.go:755:4` → Fixed
- [x] **Finding 127** | Rule: **BP-1** | Correctly Fired: **No** | _ discards int obj ID @ `internal/pdf/generator.go:924:3` → N/A: FP — not error discard
- [x] **Finding 128** | Rule: **BP-1** | Correctly Fired: **No** | _ discards int from OutputIntent @ `internal/pdf/generator.go:936:3` → N/A: FP — not error discard
- [x] **Finding 129** | Rule: **BP-1** | Correctly Fired: **No** | Zone() zone name discard @ `internal/pdf/generator.go:969:2` → N/A: FP — not error discard
- [x] **Finding 130** | Rule: **PERF-53** | Correctly Fired: **No** | uses crypto/rand not math/rand @ `internal/pdf/generator.go:1021:16` → N/A: FP — wrong RNG flagged
- [x] **Finding 131** | Rule: **PERF-44** | Correctly Fired: **No** | distinct kid assertions per iteration @ `internal/pdf/generator.go:1069:25` → N/A: FP — not repeated assertion
- [x] **Finding 132** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa in link loop @ `internal/pdf/generator.go:1196:22` → Fixed
- [x] **Finding 133** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa pageObjID in loop @ `internal/pdf/generator.go:1198:22` → Fixed
- [x] **Finding 134** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa mcid in kids loop @ `internal/pdf/generator.go:1210:21` → Fixed
- [x] **Finding 135** | Rule: **BP-1** | Correctly Fired: **No** | best-effort Close with comment @ `internal/pdf/helpers.go:46:3` → Already fixed — documented ignore
- [x] **Finding 136** | Rule: **BP-1** | Correctly Fired: **No** | best-effort Close with comment @ `internal/pdf/helpers.go:59:3` → Already fixed — documented ignore
- [x] **Finding 137** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa per xref row @ `internal/pdf/helpers.go:157:12` → Fixed
- [x] **Finding 138** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Errorf on base64 failure @ `internal/pdf/image.go:126:15` → Fixed
- [x] **Finding 139** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded closeZlibWriter @ `internal/pdf/image.go:194:4` → Fixed
- [x] **Finding 140** | Rule: **BP-1** | Correctly Fired: **Yes** | discarded closeZlibWriter @ `internal/pdf/image.go:228:4` → Fixed
- [x] **Finding 141** | Rule: **PERF-44** | Correctly Fired: **No** | type-switch pattern not repeated assert @ `internal/pdf/image.go:263:18` → N/A: FP — distinct type checks
- [x] **Finding 142** | Rule: **BP-1** | Correctly Fired: **No** | _ is unused RGBA alpha @ `internal/pdf/image.go:298:4` → N/A: FP — not error discard
- [x] **Finding 143** | Rule: **PERF-119** | Correctly Fired: **Yes** | consecutive append in XObject header @ `internal/pdf/image.go:469:6` → Fixed
- [x] **Finding 144** | Rule: **PERF-128** | Correctly Fired: **Yes** | same append chain as 143 @ `internal/pdf/image.go:469:6` → Fixed
- [x] **Finding 145** | Rule: **PERF-186** | Correctly Fired: **Yes** | strings.Fields in merge parse @ `internal/pdf/merge.go:91:13` → Fixed
- [x] **Finding 146** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa in AcroForm loop @ `internal/pdf/merge.go:163:24` → Fixed
- [x] **Finding 147** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa in Pages Kids loop @ `internal/pdf/merge.go:183:24` → Fixed
- [x] **Finding 148** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.FormatInt per xref entry @ `internal/pdf/merge.go:227:14` → Fixed
- [x] **Finding 149** | Rule: **PERF-119** | Correctly Fired: **No** | single append per padding iter @ `internal/pdf/merge.go:229:15` → N/A: FP — not consecutive appends
- [x] **Finding 150** | Rule: **PERF-128** | Correctly Fired: **No** | xref padding loop @ `internal/pdf/merge.go:229:15` → N/A: FP — not consecutive appends
- [x] **Finding 151** | Rule: **PERF-32** | Correctly Fired: **Yes** | append copies ref bytes @ `internal/pdf/merge.go:272:18` → Fixed
- [x] **Finding 152** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Errorf on annotation parse @ `internal/pdf/merge/annotations.go:232:10` → Fixed
- [x] **Finding 153** | Rule: **PERF-119** | Correctly Fired: **Yes** | three consecutive result appends @ `internal/pdf/merge/annotations.go:263:12` → Fixed
- [x] **Finding 154** | Rule: **PERF-128** | Correctly Fired: **Yes** | same triple-append as 153 @ `internal/pdf/merge/annotations.go:263:12` → Fixed
- [x] **Finding 155** | Rule: **BP-1** | Correctly Fired: **No** | Atoi ignore has nolint comment @ `internal/pdf/merge/merger.go:193:5` → Already fixed — intentional discard
- [x] **Finding 156** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.FormatInt in merger xref @ `internal/pdf/merge/merger.go:432:14` → Fixed
- [x] **Finding 157** | Rule: **PERF-119** | Correctly Fired: **No** | xref zero-padding loop @ `internal/pdf/merge/merger.go:434:15` → N/A: FP — not consecutive appends
- [x] **Finding 158** | Rule: **PERF-128** | Correctly Fired: **No** | xref zero-padding loop @ `internal/pdf/merge/merger.go:434:15` → N/A: FP — not consecutive appends
- [x] **Finding 159** | Rule: **BP-1** | Correctly Fired: **No** | reader.Close with best-effort comment @ `internal/pdf/merge/parser.go:581:3` → Already fixed — documented ignore
- [x] **Finding 160** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Errorf in ParsePageSpec @ `internal/pdf/merge/split.go:42:17` → Fixed
- [x] **Finding 161** | Rule: **PERF-121** | Correctly Fired: **No** | cited line is err check not struct @ `internal/pdf/merge/split.go:45:18` → N/A: FP — mis-aimed anchor
- [x] **Finding 162** | Rule: **PERF-42** | Correctly Fired: **Yes** | static fmt.Errorf no pages found @ `internal/pdf/merge/split.go:104:15` → Fixed
- [x] **Finding 163** | Rule: **PERF-119** | Correctly Fired: **No** | single append per chunk @ `internal/pdf/merge/split.go:151:13` → N/A: FP — not consecutive appends
- [x] **Finding 164** | Rule: **PERF-151** | Correctly Fired: **No** | tiny struct literal constructor @ `internal/pdf/metadata.go:28:1` → N/A: FP — not complex handler
- [x] **Finding 165** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf XMP conformance @ `internal/pdf/metadata.go:128:19` → Fixed
- [x] **Finding 166** | Rule: **PERF-109** | Correctly Fired: **No** | kw trim loop not map keys @ `internal/pdf/metadata.go:209:4` → N/A: FP — wrong rule
- [x] **Finding 167** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte(xmpContent) copies string @ `internal/pdf/metadata.go:265:19` → Fixed
- [x] **Finding 168** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte(bookmark title) @ `internal/pdf/outline.go:364:18` → Fixed
- [x] **Finding 169** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa dest page in outline @ `internal/pdf/outline.go:386:25` → Fixed
- [x] **Finding 170** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa count in outline @ `internal/pdf/outline.go:406:25` → Fixed
- [x] **Finding 171** | Rule: **PERF-3** | Correctly Fired: **Yes** | make([]byte) per named dest @ `internal/pdf/outline.go:504:4` → Fixed
- [x] **Finding 172** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa pageObjID in names @ `internal/pdf/outline.go:522:27` → Fixed
- [x] **Finding 173** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa struct elem ID @ `internal/pdf/outline.go:526:27` → Fixed
- [x] **Finding 174** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa pageObjID branch @ `internal/pdf/outline.go:532:27` → Fixed
- [x] **Finding 175** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf metadata object @ `internal/pdf/pdfa.go:134:9` → Fixed
- [x] **Finding 176** | Rule: **PERF-35** | Correctly Fired: **Yes** | fmt.Sprintf in field regex parse @ `internal/pdf/redact/encryption_inhouse.go:177:30` → Fixed
- [x] **Finding 177** | Rule: **CWE-916** | Correctly Fired: **No** | PDF-standard MD5 key derivation @ `internal/pdf/redact/encryption_inhouse.go:241:9` → N/A: FP — spec-required
- [x] **Finding 178** | Rule: **CWE-328** | Correctly Fired: **No** | PDF-standard MD5 key derivation @ `internal/pdf/redact/encryption_inhouse.go:241:9` → N/A: FP — spec-required
- [x] **Finding 179** | Rule: **PERF-48** | Correctly Fired: **No** | bytes.Equal with length precheck @ `internal/pdf/redact/encryption_inhouse.go:254:10` → N/A: FP — precheck present
- [x] **Finding 180** | Rule: **PERF-119** | Correctly Fired: **No** | PDF-spec deriveObjectKey cold path @ `internal/pdf/redact/encryption_inhouse.go:331:6` → N/A: FP — intentional appends
- [x] **Finding 181** | Rule: **PERF-128** | Correctly Fired: **No** | same as 181 @ `internal/pdf/redact/encryption_inhouse.go:331:6` → N/A: FP — not hot path
- [x] **Finding 182** | Rule: **PERF-35** | Correctly Fired: **No** | getOCRProvider error path only @ `internal/pdf/redact/ocr_adapter.go:39:14` → N/A: FP — cold path
- [x] **Finding 183** | Rule: **PERF-15** | Correctly Fired: **No** | OCR page loop I/O-bound @ `internal/pdf/redact/ocr_adapter.go:116:44` → Spec-required: PDF syntax needs per-object strconv formatting
- [x] **Finding 184** | Rule: **PERF-15** | Correctly Fired: **No** | duplicate Itoa in OCR iter @ `internal/pdf/redact/ocr_adapter.go:118:14` → Spec-required: PDF syntax needs per-object strconv formatting
- [x] **Finding 185** | Rule: **PERF-35** | Correctly Fired: **No** | traversePages error wrap rare @ `internal/pdf/redact/pdf_utils.go:92:12` → N/A: FP — cold error path
- [x] **Finding 186** | Rule: **BP-2** | Correctly Fired: **No** | inner walk never returns errors @ `internal/pdf/redact/pdf_utils.go:168:1` → N/A: FP — no error to wrap
- [x] **Finding 187** | Rule: **BP-1** | Correctly Fired: **No** | _ discards int position @ `internal/pdf/redact/pdf_utils.go:603:2` → N/A: FP — not error discard
- [x] **Finding 188** | Rule: **PERF-109** | Correctly Fired: **No** | UTF-16 decode loop not map keys @ `internal/pdf/redact/pdf_utils.go:700:3` → N/A: FP — wrong rule
- [x] **Finding 189** | Rule: **PERF-48** | Correctly Fired: **No** | bytes.Equal with length guard @ `internal/pdf/redact/pdf_utils.go:798:44` → N/A: FP — precheck present
- [x] **Finding 190** | Rule: **PERF-15** | Correctly Fired: **No** | strconv.Itoa per changed object @ `internal/pdf/redact/pdf_utils.go:842:19` → Spec-required: PDF syntax needs per-object strconv formatting
- [x] **Finding 191** | Rule: **PERF-15** | Correctly Fired: **No** | strconv.Itoa gen number @ `internal/pdf/redact/pdf_utils.go:844:19` → Spec-required: PDF syntax needs per-object strconv formatting
- [x] **Finding 192** | Rule: **PERF-119** | Correctly Fired: **No** | tab-split needs per-tab append @ `internal/pdf/redact/perf_helpers.go:77:11` → N/A: FP — not mergeable
- [x] **Finding 193** | Rule: **PERF-109** | Correctly Fired: **No** | hex-encoding byte loop @ `internal/pdf/redact/perf_helpers.go:145:2` → N/A: FP — not map keys
- [x] **Finding 194** | Rule: **PERF-112** | Correctly Fired: **No** | single ToLower for search norm @ `internal/pdf/redact/perf_helpers.go:188:24` → N/A: FP — not repeated compare
- [x] **Finding 195** | Rule: **PERF-48** | Correctly Fired: **No** | bytesEqualPrefix has len check @ `internal/pdf/redact/perf_helpers.go:265:9` → N/A: FP — precheck present
- [x] **Finding 196** | Rule: **PERF-35** | Correctly Fired: **No** | GetPageInfo error path only @ `internal/pdf/redact/redactor.go:90:29` → N/A: FP — cold path
- [x] **Finding 197** | Rule: **BP-1** | Correctly Fired: **No** | _ is inspectStream ok bool @ `internal/pdf/redact/redactor.go:135:4` → N/A: FP — not error discard
- [x] **Finding 198** | Rule: **PERF-112** | Correctly Fired: **No** | one-shot ToLower(mode) compare @ `internal/pdf/redact/redactor.go:191:10` → N/A: FP — not hot EqualFold case
- [x] **Finding 199** | Rule: **PERF-109** | Correctly Fired: **No** | ToLower once per search term @ `internal/pdf/redact/search.go:94:2` → N/A: FP — dedup map key
- [x] **Finding 200** | Rule: **PERF-112** | Correctly Fired: **No** | ToLower for case-insensitive search @ `internal/pdf/redact/search.go:155:16` → Acceptable: substring search needs lower
- [x] **Finding 201** | Rule: **PERF-119** | Correctly Fired: **No** | append in if/else branches @ `internal/pdf/redact/search.go:300:13` → N/A: FP — not consecutive
- [x] **Finding 202** | Rule: **PERF-15** | Correctly Fired: **No** | warning string on rare page error @ `internal/pdf/redact/secure.go:42:40` → Spec-required: PDF syntax needs per-object strconv formatting
- [x] **Finding 203** | Rule: **PERF-15** | Correctly Fired: **No** | warning on no content streams @ `internal/pdf/redact/secure.go:49:40` → Spec-required: PDF syntax needs per-object strconv formatting
- [x] **Finding 204** | Rule: **BP-1** | Correctly Fired: **No** | _ discards stream []byte @ `internal/pdf/redact/secure.go:84:2` → N/A: FP — not error discard
- [x] **Finding 205** | Rule: **PERF-46** | Correctly Fired: **Yes** | TrimSpace in query loop @ `internal/pdf/redact/secure.go:248:12` → Fixed
- [x] **Finding 206** | Rule: **PERF-32** | Correctly Fired: **Yes** | []byte from strings.Builder @ `internal/pdf/redact/secure.go:275:21` → Fixed
- [x] **Finding 207** | Rule: **PERF-109** | Correctly Fired: **No** | rune loop for glyph masking @ `internal/pdf/redact/secure.go:295:4` → N/A: FP — not map keys
- [x] **Finding 208** | Rule: **PERF-15** | Correctly Fired: **No** | strconv.Itoa per TJ segment @ `internal/pdf/redact/secure.go:446:20` → Spec-required: PDF syntax needs per-object strconv formatting
- [x] **Finding 209** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Errorf on page-not-found @ `internal/pdf/redact/visual.go:50:16` → N/A: FP — error path only
- [x] **Finding 210** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa per redacted page @ `internal/pdf/redact/visual.go:65:19` → Fixed
- [x] **Finding 211** | Rule: **PERF-15** | Correctly Fired: **Yes** | strconv.Itoa stream length @ `internal/pdf/redact/visual.go:73:25` → Fixed
- [x] **Finding 212** | Rule: **BP-1** | Correctly Fired: **No** | parseObjectKeyPrefix discards gen @ `internal/pdf/redact/visual.go:89:2` → N/A: FP — intentional
- [x] **Finding 213** | Rule: **BP-1** | Correctly Fired: **No** | pem.Decode has no error return @ `internal/pdf/signature/signature.go:67:2` → N/A: FP — not error discard
- [x] **Finding 214** | Rule: **PERF-35** | Correctly Fired: **No** | cert-parse one-time init @ `internal/pdf/signature/signature.go:74:15` → N/A: FP — cold path
- [x] **Finding 215** | Rule: **BP-1** | Correctly Fired: **No** | pem.Decode has no error return @ `internal/pdf/signature/signature.go:81:2` → N/A: FP — not error discard
- [x] **Finding 216** | Rule: **PERF-3** | Correctly Fired: **No** | make+copy for pem.Decode safety @ `internal/pdf/signature/signature.go:102:3` → Acceptable: PEM buffer isolation
- [x] **Finding 217** | Rule: **BP-1** | Correctly Fired: **No** | chain pem.Decode remainder @ `internal/pdf/signature/signature.go:104:3` → N/A: FP — not error discard
- [x] **Finding 218** | Rule: **PERF-40** | Correctly Fired: **No** | single time.Now for signing @ `internal/pdf/signature/signature.go:185:9` → N/A: FP — not repeated hot calls
- [x] **Finding 219** | Rule: **BP-1** | Correctly Fired: **No** | Zone() name discard @ `internal/pdf/signature/signature.go:186:2` → N/A: FP — not error discard
- [x] **Finding 220** | Rule: **PERF-42** | Correctly Fired: **Yes** | static fmt.Errorf unsupported key @ `internal/pdf/signature/signature.go:454:15` → Fixed
- [x] **Finding 221** | Rule: **PERF-44** | Correctly Fired: **No** | one type assertion on first kid @ `internal/pdf/structure.go:290:22` → N/A: FP — not repeated
- [x] **Finding 222** | Rule: **PERF-186** | Correctly Fired: **Yes** | strings.Fields in SVG parse @ `internal/pdf/svg/svg.go:306:23` → Fixed
- [x] **Finding 223** | Rule: **PERF-35** | Correctly Fired: **No** | font-download error paths @ `pkg/fontutils/fontutils.go:191:10` → N/A: FP — one-time cold setup
- [x] **Finding 224** | Rule: **PERF-15** | Correctly Fired: **No** | benchmark data generator @ `sampledata/benchmarks/gen_data.go:23:12` → N/A: FP — demo/sample code
- [x] **Finding 225** | Rule: **CWE-497** | Correctly Fired: **No** | commented NumCPU in sample @ `sampledata/gopdflib/financial_report/main.go:52:19` → N/A: FP — no endpoint
- [x] **Finding 226** | Rule: **PERF-148** | Correctly Fired: **No** | buffered jobs channel with receiver @ `sampledata/gopdflib/financial_report/main.go:70:10` → N/A: FP — coordinated shutdown
- [x] **Finding 227** | Rule: **PERF-40** | Correctly Fired: **No** | benchmark timing per iteration @ `sampledata/gopdflib/financial_report/main.go:88:14` → N/A: FP — demo/sample code
- [x] **Finding 228** | Rule: **PERF-15** | Correctly Fired: **No** | strconv.Itoa in demo template @ `sampledata/gopdflib/zerodha/main.go:344:52` → N/A: FP — sample code
- [x] **Finding 229** | Rule: **PERF-35** | Correctly Fired: **No** | fmt.Sprintf in demo template @ `sampledata/gopdflib/zerodha/main.go:450:54` → N/A: FP — sample code
- [x] **Finding 230** | Rule: **PERF-15** | Correctly Fired: **No** | strconv.Itoa in HFT demo @ `sampledata/gopdflib/zerodha/main.go:513:52` → N/A: FP — sample code
- [x] **Finding 231** | Rule: **PERF-15** | Correctly Fired: **No** | second Itoa in demo loop @ `sampledata/gopdflib/zerodha/main.go:517:52` → N/A: FP — sample code
- [x] **Finding 232** | Rule: **PERF-148** | Correctly Fired: **No** | buffered channels in benchmark @ `sampledata/gopdflib/zerodha/main.go:694:10` → N/A: FP — coordinated shutdown
- [x] **Finding 233** | Rule: **PERF-40** | Correctly Fired: **No** | benchmark worker timing @ `sampledata/gopdflib/zerodha/main.go:707:16` → N/A: FP — sample code
- [x] **Finding 234** | Rule: **PERF-119** | Correctly Fired: **No** | conditional append in if/else @ `typstsyntax/parser.go:233:11` → N/A: FP — not consecutive
- [x] **Finding 235** | Rule: **PERF-119** | Correctly Fired: **No** | fraction layout distinct loops @ `typstsyntax/renderer.go:292:14` → N/A: FP — not mergeable
- [x] **Finding 236** | Rule: **PERF-121** | Correctly Fired: **No** | composite MathElement literals @ `typstsyntax/renderer.go:1287:4` → N/A: FP — T(x) not applicable


## Understanding TP vs Remediation

All **105 true positives** received code fixes per user request. FP findings (131) remain marked N/A.

| Category | Count |
|----------|------:|
| **Fixed** | 105 |
| **N/A (FP)** | 131 |

## Summary Stats

- Total findings analyzed: 236
- Correctly Fired (True Positives): 105
- Incorrectly Fired (False Positives): 131
- FP rate: 55.5%
- Remediation: 105 fixed (all TPs), 131 N/A (false positive)

## Notable FP patterns observed

- [x] **BP-1 on non-errors:** `_` discards `ok` bool, `found` bool, `int` returns, `pem.Decode` remainder — not error returns
- [x] **PERF-109 misfire:** any `for` loop flagged without map-key evidence
- [x] **PERF-107 TTF parsing:** 23 hits fixed — replaced loop `binary.Read` with `binary.BigEndian` slice reads
- [x] **Hot-path on cold code:** startup (`main`), benchmarks, sampledata, one-time encryption init
- [x] **PDF-spec security:** CWE-916/328 on MD5 in Standard Security Handler — spec-mandated
- [x] **Channel semantics:** PERF-148 on buffered semaphores with defer receive
- [x] **PERF-15 in PDF generation:** strconv in loops required for unique object IDs and stream syntax

## Rule Reliability (post-review)

| Rule | TP | FP | FP Rate |
|------|---:|---:|--------:|
| PERF-15 | 29 | 12 | 29% |
| BP-1 | 9 | 27 | 75% |
| PERF-35 | 8 | 16 | 67% |
| PERF-107 | 23 | 0 | 0% |
| PERF-119 | 4 | 10 | 71% |
| PERF-32 | 11 | 2 | 15% |
| PERF-109 | 0 | 9 | 100% |
| PERF-128 | 2 | 4 | 67% |
| PERF-48 | 0 | 6 | 100% |
| PERF-3 | 2 | 3 | 60% |
| PERF-40 | 1 | 3 | 75% |
| PERF-42 | 2 | 2 | 50% |
| PERF-186 | 2 | 1 | 33% |
| PERF-41 | 2 | 1 | 33% |
| PERF-112 | 0 | 3 | 100% |
| PERF-44 | 0 | 3 | 100% |
| PERF-121 | 0 | 3 | 100% |
| PERF-151 | 0 | 3 | 100% |
| PERF-148 | 0 | 3 | 100% |
| PERF-6 | 0 | 2 | 100% |
| PERF-192 | 2 | 0 | 0% |
| PERF-31 | 1 | 1 | 50% |
| CWE-328 | 1 | 1 | 50% |
| PERF-110 | 0 | 2 | 100% |
| PERF-53 | 0 | 2 | 100% |
| PERF-114 | 0 | 2 | 100% |
| BP-2 | 1 | 1 | 50% |
| CWE-916 | 1 | 1 | 50% |
| CWE-497 | 0 | 2 | 100% |
| BP-9 | 0 | 1 | 100% |
| PERF-22 | 1 | 0 | 0% |
| PERF-56 | 0 | 1 | 100% |
| PERF-46 | 1 | 0 | 0% |
| PERF-200 | 0 | 1 | 100% |
| PERF-57 | 0 | 1 | 100% |
| CWE-22 | 0 | 1 | 100% |
| PERF-43 | 1 | 0 | 0% |
| BP-3 | 1 | 0 | 0% |
| PERF-68 | 0 | 1 | 100% |
