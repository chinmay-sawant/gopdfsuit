# Agent 2 — Findings 151–300 Validation (Manual Review)

## Per-Finding Table
| Finding | Rule | Correctly Fired? | Reason |
|---------|------|------------------|--------|
| 151 | PERF-192 | Yes | make(map) without len(glyphs) size hint |
| 152 | PERF-3 | Yes | make([]byte) rebuilt inside glyph loop |
| 153 | PERF-107 | Yes | binary.Write called inside glyph loop |
| 154 | PERF-107 | Yes | binary.Write called inside glyph loop |
| 155 | PERF-107 | Yes | binary.Write called inside glyph loop |
| 156 | PERF-107 | Yes | binary.Write called inside glyph loop |
| 157 | PERF-192 | Yes | make(map) without len(CharToGlyph) hint |
| 158 | PERF-107 | Yes | binary.Write inside cmap subset loop |
| 159 | PERF-107 | Yes | binary.Write inside cmap subset loop |
| 160 | PERF-107 | Yes | binary.Write inside cmap subset loop |
| 161 | PERF-107 | Yes | binary.Write inside cmap subset loop |
| 162 | PERF-121 | No | Boolean if check, not struct-literal conversion |
| 163 | PERF-107 | Yes | binary.Write inside name-table loop |
| 164 | PERF-107 | Yes | binary.Write inside name-table loop |
| 165 | PERF-107 | Yes | binary.Write inside name-table loop |
| 166 | PERF-107 | Yes | binary.Write inside name-table loop |
| 167 | PERF-107 | Yes | binary.Write inside name-table loop |
| 168 | PERF-107 | Yes | binary.Write inside name-table loop |
| 169 | PERF-35 | No | fmt.Errorf only on file-read error path |
| 170 | PERF-192 | Yes | make(map) without table-count size hint |
| 171 | PERF-192 | Yes | make(map) without size hint |
| 172 | PERF-192 | Yes | make(map) without size hint |
| 173 | PERF-107 | Yes | binary.Read inside table-directory loop |
| 174 | PERF-107 | Yes | binary.Read inside table-directory loop |
| 175 | PERF-107 | Yes | binary.Read inside table-directory loop |
| 176 | PERF-107 | Yes | binary.Read inside hmtx parse loop |
| 177 | PERF-107 | Yes | binary.Read inside cmap parse loop |
| 178 | PERF-107 | Yes | binary.Read inside cmap parse loop |
| 179 | PERF-107 | Yes | binary.Read inside cmap parse loop |
| 180 | PERF-107 | Yes | binary.Read inside cmap subtable loop |
| 181 | PERF-107 | Yes | binary.Read inside cmap subtable loop |
| 182 | PERF-107 | Yes | binary.Read inside format-4 cmap loop |
| 183 | PERF-107 | Yes | binary.Read inside format-4 cmap loop |
| 184 | PERF-107 | Yes | binary.Read inside format-4 cmap loop |
| 185 | BP-1 | Yes | r.Seek error discarded into blank identifier |
| 186 | PERF-107 | Yes | binary.Read inside format-4 cmap loop |
| 187 | PERF-107 | Yes | binary.Read inside format-4 cmap loop |
| 188 | PERF-107 | Yes | binary.Read inside format-12 cmap loop |
| 189 | PERF-107 | Yes | binary.Read inside format-12 cmap loop |
| 190 | PERF-107 | Yes | binary.Read inside format-12 group loop |
| 191 | PERF-107 | Yes | binary.Read inside name-table loop |
| 192 | PERF-107 | Yes | binary.Read inside name-table loop |
| 193 | PERF-107 | Yes | binary.Read inside name-table loop |
| 194 | PERF-107 | Yes | binary.Read inside name-table loop |
| 195 | PERF-107 | Yes | binary.Read inside name-table loop |
| 196 | PERF-107 | Yes | binary.Read inside name-table loop |
| 197 | PERF-192 | Yes | make(map) without text-length size hint |
| 198 | PERF-192 | Yes | make(map) without len(fields) hint |
| 199 | BP-1 | Yes | r.Close error explicitly discarded in defer |
| 200 | BP-1 | Yes | r.Close error explicitly discarded in defer |
| 201 | PERF-188 | Yes | fmt.Sscanf inside parseArrayInts loop |
| 202 | PERF-32 | Yes | []byte literal conversion inside loop condition |
| 203 | PERF-32 | Yes | []byte literal conversion inside loop condition |
| 204 | PERF-1 | Yes | regexp.MustCompile inside widget loop |
| 205 | PERF-1 | Yes | regexp.MustCompile inside widget loop |
| 206 | PERF-1 | Yes | regexp.MustCompile inside widget loop |
| 207 | PERF-1 | Yes | regexp.MustCompile inside widget loop |
| 208 | PERF-1 | Yes | regexp.MustCompile inside widget loop |
| 209 | PERF-32 | Yes | []byte literal conversion inside loop condition |
| 210 | PERF-32 | Yes | []byte literal conversion inside loop condition |
| 211 | PERF-32 | Yes | []byte literal conversion inside loop condition |
| 212 | PERF-1 | Yes | regexp.MustCompile inside object-stream loop |
| 213 | PERF-109 | Yes | Map key string rebuilt each loop iteration |
| 214 | PERF-1 | Yes | regexp.MustCompile inside object-stream loop |
| 215 | PERF-35 | Yes | fmt.Sprintf boxes args inside loop |
| 216 | PERF-6 | Yes | fmt.Sprintf inside object-stream loop |
| 217 | PERF-192 | Yes | make(map) without known-size hint |
| 218 | PERF-1 | Yes | regexp.MustCompile inside stream parse loop |
| 219 | PERF-1 | Yes | regexp.MustCompile inside stream parse loop |
| 220 | PERF-188 | Yes | fmt.Sscanf used to parse matched digits |
| 221 | PERF-186 | Yes | strings.Fields on header inside hot loop |
| 222 | PERF-188 | Yes | fmt.Sscanf inside object-header parse loop |
| 223 | BP-1 | No | err2 checked with err2 == nil, not discarded |
| 224 | PERF-188 | Yes | fmt.Sscanf inside object-offset parse loop |
| 225 | PERF-6 | Yes | fmt.Sprintf inside object-map loop |
| 226 | PERF-192 | Yes | make(map) without size hint |
| 227 | PERF-192 | Yes | make(map) without size hint |
| 228 | PERF-192 | Yes | make(map) without size hint |
| 229 | PERF-192 | Yes | make(map) without size hint |
| 230 | PERF-114 | No | Map-merge loop, not manual slice copy |
| 231 | PERF-114 | No | Map-merge loop, not manual slice copy |
| 232 | BP-1 | Yes | xml.Marshal error discarded into blank identifier |
| 233 | PERF-192 | Yes | make(map) without fields-size hint |
| 234 | PERF-1 | Yes | regexp.MustCompile inside widget-match loop |
| 235 | PERF-1 | Yes | regexp.MustCompile inside widget-match loop |
| 236 | PERF-186 | Yes | strings.Fields on rect coords in loop |
| 237 | BP-1 | Yes | ParseFloat error discarded into blank identifier |
| 238 | BP-1 | Yes | ParseFloat error discarded into blank identifier |
| 239 | BP-1 | Yes | ParseFloat error discarded into blank identifier |
| 240 | BP-1 | Yes | ParseFloat error discarded into blank identifier |
| 241 | PERF-1 | Yes | regexp.MustCompile inside widget-match loop |
| 242 | BP-1 | Yes | Atoi error discarded into blank identifier |
| 243 | PERF-1 | Yes | regexp.MustCompile inside widget-match loop |
| 244 | PERF-1 | Yes | regexp.MustCompile inside widget-match loop |
| 245 | PERF-192 | Yes | make(map) without jobs-size hint |
| 246 | PERF-1 | Yes | regexp.MustCompile inside radioGroups loop |
| 247 | PERF-6 | Yes | fmt.Sprintf inside radioGroups loop |
| 248 | PERF-1 | Yes | regexp.MustCompile inside radioGroups loop |
| 249 | PERF-32 | Yes | []byte(fmt.Sprintf(...)) inside allJobs loop |
| 250 | PERF-6 | Yes | fmt.Sprintf inside allJobs loop |
| 251 | PERF-1 | Yes | regexp.MustCompile inside allJobs loop |
| 252 | PERF-1 | Yes | regexp.MustCompile inside allJobs loop |
| 253 | PERF-112 | Yes | strings.ToLower before string equality compare |
| 254 | PERF-112 | Yes | strings.ToLower before string equality compare |
| 255 | PERF-1 | Yes | regexp.MustCompile inside allJobs loop |
| 256 | PERF-1 | Yes | regexp.MustCompile inside allJobs loop |
| 257 | PERF-1 | Yes | regexp.MustCompile inside textJobs loop |
| 258 | PERF-6 | Yes | fmt.Sprintf argument inside textJobs loop |
| 259 | PERF-32 | Yes | []byte(fmt.Sprintf(...)) inside textJobs loop |
| 260 | PERF-6 | Yes | fmt.Sprintf inside textJobs loop |
| 261 | PERF-119 | Yes | Consecutive append calls on same slice |
| 262 | PERF-128 | Yes | Consecutive append calls on same slice |
| 263 | PERF-6 | Yes | fmt.Sprintf inside textJobs loop |
| 264 | PERF-6 | Yes | fmt.Sprintf inside textJobs loop |
| 265 | PERF-32 | Yes | []byte(string) conversion inside textJobs loop |
| 266 | PERF-15 | Yes | strconv.Itoa inside helveticaWidths loop |
| 267 | PERF-6 | Yes | fmt.Sprintf inside textJobs loop |
| 268 | PERF-32 | Yes | []byte(string) conversion inside textJobs loop |
| 269 | BP-1 | Yes | NewWriterLevel error discarded into blank identifier |
| 270 | PERF-32 | Yes | []byte(streamBody) conversion inside textJobs loop |
| 271 | PERF-6 | Yes | fmt.Sprintf inside textJobs loop |
| 272 | PERF-32 | Yes | []byte(string) conversion inside textJobs loop |
| 273 | PERF-192 | Yes | make(map) without objMatches size hint |
| 274 | BP-1 | Yes | Atoi error discarded inside xref loop |
| 275 | PERF-6 | Yes | fmt.Fprintf inside xref entry loop |
| 276 | PERF-192 | Yes | make(map) without size hint |
| 277 | PERF-1 | Yes | regexp.MustCompile inside field-processing loop |
| 278 | PERF-1 | Yes | regexp.MustCompile inside field-processing loop |
| 279 | PERF-1 | Yes | regexp.MustCompile inside kids-parse loop |
| 280 | BP-1 | Yes | Atoi error discarded into blank identifier |
| 281 | PERF-1 | Yes | regexp.MustCompile inside kids-parse loop |
| 282 | BP-1 | Yes | Atoi error discarded into blank identifier |
| 283 | PERF-1 | Yes | regexp.MustCompile inside AP-removal loop |
| 284 | PERF-15 | Yes | strconv.Itoa inside object-stream header loop |
| 285 | PERF-15 | Yes | strconv.Itoa inside object-stream header loop |
| 286 | PERF-48 | Yes | bytes.Equal without length or prefix precheck |
| 287 | PERF-110 | No | Snippet shows Pool decl only, not value-type New |
| 288 | PERF-192 | Yes | make(map) without xref size hint |
| 289 | PERF-192 | Yes | make(map) without image-count size hint |
| 290 | PERF-192 | Yes | make(map) without image-count size hint |
| 291 | PERF-192 | Yes | make(map) without cell-image size hint |
| 292 | PERF-192 | Yes | make(map) without cell-image size hint |
| 293 | PERF-192 | Yes | make(map) without element-image size hint |
| 294 | PERF-192 | Yes | make(map) without element-image size hint |
| 295 | PERF-32 | No | One-time encryption setup, not hot path |
| 296 | PERF-35 | No | One-time encryption setup, not hot path |
| 297 | PERF-192 | Yes | make(map) without font-count size hint |
| 298 | PERF-192 | Yes | make(map) without font-count size hint |
| 299 | PERF-192 | Yes | make(map) without font-count size hint |
| 300 | PERF-192 | Yes | make(map) without group-count size hint |

## Summary
- Total: 150
- True Positives: 142
- False Positives: 8
- FP rate: 5.3%

## Notable FP patterns observed
- **PERF-121 mis-anchoring:** Rule fired on a boolean `if` condition instead of a struct-literal vs type-conversion pattern (finding 162).
- **Hot-path rules on cold paths:** PERF-35 and PERF-32 flagged one-time encryption setup and error-only branches where formatting or conversion is not on a hot loop (findings 169, 295, 296).
- **BP-1 on checked errors:** Rule flagged `if _, err := ...; err == nil` patterns where the error is handled, not discarded (finding 223).
- **PERF-114 on map merges:** Rule flagged map-copy loops that assign into maps rather than manual slice `copy()` replacements (findings 230, 231).
- **PERF-110 incomplete match:** Finding anchored on `sync.Pool` declaration without snippet evidence that `New` returns a value type (finding 287).