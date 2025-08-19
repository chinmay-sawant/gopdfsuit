#!/usr/bin/env python3
"""
Simple PDF flattener that does not use external libraries.

What it does:
- Parses objects from a single-file PDF.
- Collects widget annotations from the first page (/Annots in the page object).
- For each widget with a /V (value), creates a page content stream that draws the text at the widget rectangle
  using the font named in the widget /DA or the AcroForm default /DA.
- Removes the /Annots entry from the page so fields no longer appear as interactive widgets.
- Writes a new PDF with an added content stream object and a rebuilt xref/trailer.

Limitations / assumptions:
- Works on simple PDFs similar to the attached sample (single page, objects numbered sequentially).
- Does not decode or edit existing compressed content streams; instead it appends a new content stream.
- Attempts to preserve existing font mapping by merging the page /Font resource that referenced an indirect
  font-dictionary (e.g. "1 0 R") and adds a /Helv entry referencing the AcroForm DR font (object typically 5 0 R).
- This is a best-effort flattener; complex PDFs or forms with rich appearances may need a full PDF library.

Usage:
  python scripts/flatten_pdf.py internal/pdf/filled_sample.pdf internal/pdf/filled_sample_flat.pdf

"""
import re
import sys
import os


def read_bytes(path):
    with open(path, 'rb') as f:
        return f.read()


def write_bytes(path, data):
    with open(path, 'wb') as f:
        f.write(data)


def find_objects(pdf_bytes):
    # find all object bodies between "N 0 obj" and "endobj"
    objs = {}
    for m in re.finditer(rb"(\d+)\s+0\s+obj(.*?)endobj", pdf_bytes, re.S):
        num = int(m.group(1))
        body = m.group(2).strip()
        objs[num] = body
    return objs


def get_trailer(pdf_bytes):
    m = re.search(rb"trailer(.*?)startxref", pdf_bytes, re.S)
    if not m:
        return b""
    return m.group(1).strip()


def extract_page_info(objs, page_obj_num):
    body = objs[page_obj_num].decode('latin1')
    # get Annots list of object refs
    annots = []
    m = re.search(r"/Annots\s*\[(.*?)\]", body, re.S)
    if m:
        refs = re.findall(r"(\d+)\s+0\s+R", m.group(1))
        annots = [int(r) for r in refs]

    # find Contents ref (single)
    m2 = re.search(r"/Contents\s+(\d+)\s+0\s+R", body)
    contents_ref = int(m2.group(1)) if m2 else None

    # find Resources and whether it references a font indirect ref
    res_m = re.search(r"/Resources\s*(<<.*?>>)", body, re.S)
    resources = res_m.group(1) if res_m else None

    return {'body': body, 'annots': annots, 'contents': contents_ref, 'resources': resources}


def parse_widget(obj_body):
    s = obj_body.decode('latin1')
    # Only process /Subtype /Widget
    if '/Subtype /Widget' not in s:
        return None
    # Rect
    rect_m = re.search(r"/Rect\s*\[\s*([\d\.-]+)\s+([\d\.-]+)\s+([\d\.-]+)\s+([\d\.-]+)\s*\]", s)
    rect = None
    if rect_m:
        rect = [float(rect_m.group(i)) for i in range(1,5)]
    # Value /V (literal string). Extract robustly to handle escaped parentheses
    value = None
    vm = re.search(r"/V\s*\(", s)
    if vm:
        start = vm.end()  # index after the opening '('
        # scan for matching ')' respecting backslash-escapes
        i = start
        buf = []
        data = s
        length = len(data)
        while i < length:
            ch = data[i]
            if ch == ')':
                # end of literal string
                break
            if ch == '\\':
                # escape sequence; include both backslash and next char for decoding
                if i + 1 < length:
                    esc = data[i+1]
                    buf.append('\\' + esc)
                    i += 2
                    continue
                else:
                    # dangling backslash
                    buf.append('\\')
                    i += 1
                    continue
            else:
                buf.append(ch)
                i += 1
        raw_value = ''.join(buf)
        # decode PDF literal string escapes like \(, \), \\ and octal
        def decode_pdf_literal(sraw):
            out = []
            i = 0
            L = len(sraw)
            while i < L:
                c = sraw[i]
                if c != '\\':
                    out.append(c)
                    i += 1
                else:
                    # escape
                    i += 1
                    if i >= L:
                        break
                    e = sraw[i]
                    if e == 'n':
                        out.append('\n')
                        i += 1
                    elif e == 'r':
                        out.append('\r')
                        i += 1
                    elif e == 't':
                        out.append('\t')
                        i += 1
                    elif e == 'b':
                        out.append('\b')
                        i += 1
                    elif e == 'f':
                        out.append('\f')
                        i += 1
                    elif e in ('\\', '(', ')'):
                        out.append(e)
                        i += 1
                    elif e.isdigit():
                        # up to three octal digits
                        octal = e
                        i += 1
                        for _ in range(2):
                            if i < L and sraw[i].isdigit():
                                octal += sraw[i]
                                i += 1
                            else:
                                break
                        try:
                            out.append(chr(int(octal, 8)))
                        except Exception:
                            # fallback: raw
                            out.append(octal)
                    else:
                        # unknown escape, keep char
                        out.append(e)
                        i += 1
            return ''.join(out)

        value = decode_pdf_literal(raw_value)
    # Default Appearance /DA
    da_m = re.search(r"/DA\s*\((.*?)\)", s)
    da = da_m.group(1) if da_m else None
    # Field type /FT (may be on the widget or inherited from /Parent)
    ft_m = re.search(r"/FT\s*/(\w+)", s)
    ft = ft_m.group(1) if ft_m else None
    # Parent reference (for inheritance of /FT)
    parent_m = re.search(r"/Parent\s+(\d+)\s+0\s+R", s)
    parent = int(parent_m.group(1)) if parent_m else None
    return {'rect': rect, 'value': value, 'da': da}


def acroform_da_and_font(objs):
    # Search AcroForm object (obj containing /AcroForm or object 121 in sample)
    for num, body in objs.items():
        s = body.decode('latin1')
        if '/AcroForm' in s or '/NeedAppearances' in s:
            da_m = re.search(r"/DA\s*\((.*?)\)", s)
            da = da_m.group(1) if da_m else None
            # Try to find DR font mapping like /DR << /Font << /Helv 5 0 R >> >>
            font_ref = None
            dr_m = re.search(r"/DR\s*(<<.*?>>)", s, re.S)
            if dr_m:
                helv_m = re.search(r"/Helv\s+(\d+)\s+0\s+R", dr_m.group(1))
                if helv_m:
                    font_ref = int(helv_m.group(1))
            return {'da': da, 'helv_ref': font_ref}
    return {'da': None, 'helv_ref': None}


def escape_pdf_text(t):
    # escape backslash, left and right paren for PDF literal strings
    return t.replace('\\', '\\\\').replace('(', '\\(').replace(')', '\\)')


def build_text_draw_ops(fields, acro_da):
    ops = []
    for f in fields:
        if not f['value'] or not f['rect']:
            continue
        x0,y0,x1,y1 = f['rect']
        width = x1 - x0
        height = y1 - y0
        # derive font size from DA like "/Helv 12 Tf 0 0 0 rg"
        size = 12
        if f['da']:
            m = re.search(r"(\d+(?:\.\d+)?)\s+Tf", f['da'])
            if m:
                try:
                    size = float(m.group(1))
                except:
                    size = 12
        elif acro_da and acro_da.get('da'):
            m = re.search(r"(\d+(?:\.\d+)?)\s+Tf", acro_da['da'])
            if m:
                try:
                    size = float(m.group(1))
                except:
                    size = 12

        # basic vertical centering and small left padding
        tx = x0 + 2
        ty = y0 + max(0, (height - size) / 2.0)
        text = escape_pdf_text(f['value'])
        # Build a simple text drawing command using the /Helv font and rgb black
        cmd = "BT /Helv {size} Tf 0 0 0 rg {tx:.3f} {ty:.3f} Td ({text}) Tj ET".format(
            size=size, tx=tx, ty=ty, text=text)
        ops.append(cmd)
    if not ops:
        return b""
    # wrap with q/Q to preserve graphics state and separate from page content
    content = "q\n" + "\n".join(ops) + "\nQ\n"
    return content.encode('latin1')


def merge_font_resources(page_resources_text, font_obj_text):
    # page_resources_text like "<< /Font 1 0 R /ProcSet [...] >>"
    # font_obj_text like "<< /F1 2 0 R /F2 3 0 R >>"
    # produce combined /Font << /Helv 5 0 R /F1 2 0 R /F2 3 0 R >>
    pr = page_resources_text
    font_obj_inner = re.sub(r"^<<|>>$", "", font_obj_text.strip()).strip()
    # extract existing ProcSet so we can keep it
    proc_m = re.search(r"/ProcSet\s*(\[.*?\])", pr, re.S)
    proc = proc_m.group(1) if proc_m else None

    # build new resources block
    font_entries = font_obj_inner
    new_font = "<< /Helv 5 0 R " + (" " + font_entries if font_entries else "") + " >>"
    # assemble resources
    if proc:
        new_res = "<< /Font %s /ProcSet %s >>" % (new_font, proc)
    else:
        new_res = "<< /Font %s >>" % new_font
    return new_res


def rebuild_pdf(objs, modified_objs, trailer_raw, out_path):
    # objs: dict of original objects body bytes
    # modified_objs: dict num->body (string or bytes) to override
    # Build header
    header = b"%PDF-1.4\n%\xFF\xFF\xFF\xFF\n"
    # find max obj number
    max_obj = max(max(objs.keys()) if objs else 0, max(modified_objs.keys()) if modified_objs else 0)
    # if modified_objs contain higher numbers, include them
    # prepare objects in order from 1..max_obj
    out = bytearray()
    out += header
    offsets = {}
    # object 0 entry will be the free entry
    offsets[0] = 0
    for i in range(1, max_obj+1):
        offsets[i] = len(out)
        body = None
        if i in modified_objs:
            body = modified_objs[i]
        elif i in objs:
            body = objs[i]
        else:
            # emit empty obj to preserve numbering
            body = b"<< /Type /Null >>"
        # body may be str or bytes
        if isinstance(body, str):
            body_bytes = body.encode('latin1')
        else:
            body_bytes = body
        out += (str(i) + " 0 obj\n").encode('latin1')
        out += body_bytes
        out += b"\nendobj\n"

    # xref
    xref_offset = len(out)
    out += b"xref\n"
    out += ("0 %d\n" % (max_obj+1)).encode('latin1')
    out += b"0000000000 65535 f \n"
    for i in range(1, max_obj+1):
        out += ("%010d 00000 n \n" % offsets[i]).encode('latin1')

    # trailer: keep original entries if possible but update /Size
    tr = trailer_raw.decode('latin1')
    # replace /Size if present
    if '/Size' in tr:
        tr = re.sub(r"/Size\s+\d+", "/Size %d" % (max_obj+1), tr)
    else:
        tr = tr + "\n/Size %d" % (max_obj+1)

    out += b"trailer\n"
    out += tr.encode('latin1')
    out += b"\nstartxref\n"
    out += (str(xref_offset) + "\n").encode('latin1')
    out += b"%%EOF\n"

    write_bytes(out_path, bytes(out))


def main(argv):
    if len(argv) < 3:
        print("Usage: python flatten_pdf.py input.pdf output.pdf")
        return 1
    inp = argv[1]
    outp = argv[2]
    pdf = read_bytes(inp)
    objs = find_objects(pdf)
    trailer = get_trailer(pdf)

    # find page object: look for object with "/Type /Page" (prefer single page)
    page_num = None
    for num, body in objs.items():
        if b"/Type /Page" in body:
            page_num = num
            break
    if page_num is None:
        print("No page object found.")
        return 1

    pinfo = extract_page_info(objs, page_num)
    annots = pinfo['annots']

    fields = []
    annots_to_keep = []
    for a in annots:
        if a in objs:
            raw = objs[a]
            w = parse_widget(raw)
            if not w:
                # keep by default
                annots_to_keep.append(a)
                continue
            # determine field type: check widget or its Parent
            is_button = False
            # check widget /FT
            s = raw.decode('latin1')
            if '/FT /Btn' in s:
                is_button = True
            else:
                # look for Parent reference
                pm = re.search(r"/Parent\s+(\d+)\s+0\s+R", s)
                if pm:
                    pnum = int(pm.group(1))
                    if pnum in objs:
                        if b'/FT /Btn' in objs[pnum]:
                            is_button = True

            if is_button:
                # preserve button annotations (do not flatten)
                annots_to_keep.append(a)
            else:
                # non-button: collect for flattening if value exists
                if w.get('value'):
                    fields.append(w)

    if not fields:
        print("No non-button widget fields with values found; only buttons/checkboxes will be preserved.")

    acro = acroform_da_and_font(objs)

    content_bytes = build_text_draw_ops(fields, acro)

    if not content_bytes:
        print("No overlay content to add; writing original file copy.")
        write_bytes(outp, pdf)
        return 0

    # Create new content stream object number = max_obj+1
    next_obj = max(objs.keys()) + 1
    # build stream object with Length entry
    stream_body = "<< /Length %d >>\nstream\n" % len(content_bytes)
    stream_body = stream_body.encode('latin1') + content_bytes + b"\nendstream"

    # Modify page object: replace /Annots with only kept annotations and replace /Contents ref
    page_body = objs[page_num].decode('latin1')
    # rebuild Annots array with preserved annotations
    if annots_to_keep:
        annots_str = ' '.join([f"{n} 0 R" for n in annots_to_keep])
        page_body = re.sub(r"/Annots\s*\[.*?\]", f"/Annots [ {annots_str} ]", page_body, flags=re.S)
    else:
        # remove Annots entirely
        page_body = re.sub(r"/Annots\s*\[.*?\]\s*", "", page_body, flags=re.S)
    # replace Contents reference
    page_body = re.sub(r"/Contents\s+\d+\s+0\s+R", "/Contents [ %d 0 R %d 0 R ]" % (pinfo['contents'], next_obj), page_body)

    # Merge font resources: if page had "/Font <num> 0 R" and object exists, include its entries and add /Helv 5 0 R
    if pinfo['resources']:
        res_text = pinfo['resources']
        # find any "/Font <num> 0 R"
        m = re.search(r"/Font\s+(\d+)\s+0\s+R", res_text)
        if m:
            font_ref_num = int(m.group(1))
            if font_ref_num in objs:
                font_obj_text = objs[font_ref_num].decode('latin1')
                # font_obj_text might be '<< /F1 2 0 R ... >>'
                new_res = merge_font_resources(res_text, font_obj_text)
                # replace the whole Resources <<...>> in page_body
                page_body = re.sub(r"/Resources\s*<<.*?>>", "/Resources %s" % new_res, page_body, flags=re.S)

    # prepare modified objects dict
    modified = {}
    modified[next_obj] = stream_body
    modified[page_num] = page_body.encode('latin1')

    # Rebuild PDF
    rebuild_pdf(objs, modified, trailer, outp)
    print("Wrote flattened PDF to", outp)
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv))
