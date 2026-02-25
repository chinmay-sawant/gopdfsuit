import zlib
import struct

class AdvancedPDFGenerator:
    """
    Generates a PDF 1.6 document from scratch using:
    - Object Streams (/ObjStm) for compression (for internal objects)
    - XRef Streams (modern PDF 1.6 cross-referencing)
    - Flate Compression for content streams
    - AcroForm structure with various field types
    """
    
    def __init__(self, filename="medical_form.pdf"):
        self.filename = filename
        self.objects = {} # ID -> Content (String) (for regular objects)
        self.obj_counter = 1
        self.fields_data = [] 
        self.field_ids = []   
        self.obj_stream_members = {} # ID -> Content (for objects inside ObjStm)

    def _get_id(self):
        val = self.obj_counter
        self.obj_counter += 1
        return val

    def compress(self, data_bytes):
        return zlib.compress(data_bytes)

    def generate_pdf(self):
        # 1. IDs for main structure
        catalog_id = self._get_id() # 1
        pages_id = self._get_id()   # 2
        page_id = self._get_id()    # 3
        font_id = self._get_id()    # 4
        content_id = self._get_id() # 5
        acroform_id = self._get_id() # 6
        
        # 2. Define Fields with Labels for the Content Stream
        self.fields_data = [
            # Row 1: Patient Name
            {"name": "patient_name", "type": "Tx", "rect": [150, 700, 350, 720], "label": "Patient Name:", "label_pos": [70, 705]},
            
            # Row 2: DOB and Phone
            {"name": "dob", "type": "Tx", "rect": [150, 670, 250, 690], "label": "DOB:", "label_pos": [70, 675]},
            {"name": "phone", "type": "Tx", "rect": [350, 670, 500, 690], "label": "Phone:", "label_pos": [280, 675]},
            
            # Row 3: Email
            {"name": "email", "type": "Tx", "rect": [150, 640, 350, 660], "label": "Email:", "label_pos": [70, 645]},
            
            # Row 4: Insurance
            {"name": "insurance", "type": "Tx", "rect": [150, 610, 350, 630], "label": "Insurance:", "label_pos": [70, 615]},
            
            # Row 5: Gender (Radio Buttons)
            # Sharing name "gender" means they are mutually exclusive part of same group
            {"name": "gender", "type": "Btn", "rect": [150, 580, 165, 595], "opt": "Male", "label": "Male", "label_pos": [170, 583], "flags": 49152},
            {"name": "gender", "type": "Btn", "rect": [220, 580, 235, 595], "opt": "Female", "label": "Female", "label_pos": [240, 583], "flags": 49152},

            # Row 6: Symptoms (Checkboxes)
            {"name": "fever", "type": "Btn", "rect": [150, 550, 165, 565], "opt": "Yes", "label": "Fever", "label_pos": [170, 553], "flags": 0},
            {"name": "cough", "type": "Btn", "rect": [150, 530, 165, 545], "opt": "Yes", "label": "Cough", "label_pos": [170, 533], "flags": 0},
            {"name": "headache", "type": "Btn", "rect": [150, 510, 165, 525], "opt": "Yes", "label": "Headache", "label_pos": [170, 513], "flags": 0},
            
            # Multiline
            {"name": "doctor_notes", "type": "Tx", "rect": [100, 350, 500, 480], "label": "Doctor Notes:", "label_pos": [100, 485], "flags": 4096},
        ]
        
        # 3. Create Field Objects (Internal Objects in ObjStm)
        self.field_ids = []
        
        for f in self.fields_data:
            fid = self._get_id()
            self.field_ids.append(fid)
            
            flags = f.get('flags', 0)
            opt = f.get('opt', '')
            
            # Construct dictionary for the field
            # appearance stream /AP is usually required for checkboxes to work visually in all viewers,
            # but simple viewers might auto-generate.
            # To be safe, we rely on NeedAppearances true.
            
            content = f"""<< 
/Type /Annot 
/Subtype /Widget 
/FT /{f['type']} 
/T ({f['name']}) 
/Rect [{f['rect'][0]} {f['rect'][1]} {f['rect'][2]} {f['rect'][3]}] 
/P {page_id} 0 R 
/DA (/Helv 10 Tf 0 g) 
"""
            if flags:
                content += f"/Ff {flags} "
            
            # Checkbox/Radio toggle value (On state)
            if opt:
                # This hint tells PDF what the "On" value is for export
                # Note: Correct way is appearance dictionary keys, but let's try just /Opt or usage
                pass

            content += ">>"
            self.obj_stream_members[fid] = content

        # 4. Create Main Structure Objects (Internal Objects in ObjStm)
        
        # AcroForm Dictionary
        fields_ref_str = " ".join([f"{fid} 0 R" for fid in self.field_ids])
        self.obj_stream_members[acroform_id] = f"""<< 
/Fields [{fields_ref_str}] 
/NeedAppearances true 
/DA (/Helv 10 Tf 0 g) 
/DR << /Font << /Helv {font_id} 0 R >> >> 
>>"""

        # Catalog
        self.obj_stream_members[catalog_id] = f"""<< 
/Type /Catalog 
/Pages {pages_id} 0 R 
/AcroForm {acroform_id} 0 R 
>>"""

        # Pages Node
        self.obj_stream_members[pages_id] = f"""<< 
/Type /Pages 
/Kids [{page_id} 0 R] 
/Count 1 
>>"""

        # Font
        self.obj_stream_members[font_id] = f"""<< 
/Type /Font 
/Subtype /Type1 
/BaseFont /Helvetica 
>>"""

        # Page Object (Also putting in ObjStm for maximum compression)
        annots_ref_str = " ".join([f"{fid} 0 R" for fid in self.field_ids])
        self.obj_stream_members[page_id] = f"""<< 
/Type /Page 
/Parent {pages_id} 0 R 
/MediaBox [0 0 595 842] 
/Contents {content_id} 0 R 
/Resources << 
  /Font << /F1 {font_id} 0 R /Helv {font_id} 0 R >> 
>> 
/Annots [{annots_ref_str}] 
>>"""

        # 5. Content Stream (Regular Object - cannot be in ObjStm)
        # Construct text drawing operations
        stream_ops = [
            "BT",
            "/F1 18 Tf",
            "100 750 Td (Medical Intake Form) Tj",
            "/F1 12 Tf",
            "0 -25 Td (Please provide your details below:) Tj",
            "ET"
        ]
        
        # Add labels dynamically
        stream_ops.append("BT /F1 10 Tf")
        for f in self.fields_data:
            if "label" in f and "label_pos" in f:
                lx, ly = f["label_pos"]
                # Escape parens
                label_text = f["label"].replace("(", "\\(").replace(")", "\\)")
                stream_ops.append(f"1 0 0 1 {lx} {ly} Tm ({label_text}) Tj")
        stream_ops.append("ET")
        
        stream_content = "\n".join(stream_ops).encode('latin1')
        compressed_content = self.compress(stream_content)
        
        self.objects[content_id] = f"""<<
/Length {len(compressed_content)}
/Filter /FlateDecode
>>
stream
{compressed_content.decode('latin1')}
endstream"""

        # 6. Build the Object Stream
        obj_stm_id = self._get_id()
        self.construct_object_stream(obj_stm_id)

        # 7. Write Result
        self.write_file(obj_stm_id, content_id)

    def construct_object_stream(self, obj_stm_id):
        # Sort objects by ID
        sorted_ids = sorted(self.obj_stream_members.keys())
        
        # ObjStm Structure:
        # N pairs of integers: "oid offset oid offset ..."
        # Followed by objects content concatenated.
        # Offsets are relative to the first object's start.
        
        pairs = []
        body_content = ""
        
        for oid in sorted_ids:
            content = self.obj_stream_members[oid]
            offset = len(body_content)
            pairs.append(f"{oid} {offset}")
            body_content += content + " " # Check spacing
            
        header_str = " ".join(pairs)
        first_offset = len(header_str) + 1 # +1 for the space usually or strictly checks
        # Actually usually it's just "header body".
        
        # We need to be careful. The "First" parameter gives the byte offset of the first object in the decoded stream.
        # So "header_str" is at the start. "body_content" starts after header.
        
        # Let's align properly
        full_stream_text = header_str + " " + body_content
        first = len(header_str) + 1
        
        compressed_stm = self.compress(full_stream_text.encode('latin1'))
        
        self.objects[obj_stm_id] = f"""<<
/Type /ObjStm
/N {len(sorted_ids)}
/First {first}
/Length {len(compressed_stm)}
/Filter /FlateDecode
>>
stream
{compressed_stm.decode('latin1')}
endstream"""

    def write_file(self, obj_stm_id, content_id):
        with open(self.filename, "wb") as f:
            # Header
            f.write(b"%PDF-1.6\n%\xe2\xe3\xcf\xd3\n")
            
            offsets = {}
            
            # Write Main Objects
            # 1. ObjStm
            offsets[obj_stm_id] = f.tell()
            f.write(f"{obj_stm_id} 0 obj\n{self.objects[obj_stm_id]}\nendobj\n\n".encode('latin1'))
            
            # 2. Content Stream
            offsets[content_id] = f.tell()
            f.write(f"{content_id} 0 obj\n{self.objects[content_id]}\nendobj\n\n".encode('latin1'))
            
            # --- XRef Stream ---
            xref_oid = self._get_id()
            startxref_offset = f.tell()
            
            # Entries construction
            # Type 1: Standard (ContentStream, ObjStm, XRef)
            # Type 2: Compressed (Inside ObjStm)
            
            entries = {}
            entries[0] = (0, 0, 65535) # Free
            
            # Standard
            entries[obj_stm_id] = (1, offsets[obj_stm_id], 0)
            entries[content_id] = (1, offsets[content_id], 0)
            # XRef stream is usually standard (1) at its own offset.
            # But we can't write its offset in itself perfectly if we compress it...
            # Actually standard practice for XRef stream is to point to itself? No, XRef stream doesn't need to be in the index usually?
            # But to be safe, we add it. 
            entries[xref_oid] = (1, startxref_offset, 0)
            
            # Internal
            sorted_internal = sorted(self.obj_stream_members.keys())
            for idx, oid in enumerate(sorted_internal):
                entries[oid] = (2, obj_stm_id, idx)
            
            # Build binary
            xref_data = bytearray()
            
            # We must cover range 0 to xref_oid
            # If there are gaps, fill with Type 0
            for i in range(xref_oid + 1):
                if i in entries:
                    t, f2, f3 = entries[i]
                    xref_data.extend(struct.pack('>B I H', t, f2, f3))
                else:
                    xref_data.extend(struct.pack('>B I H', 0, 0, 0))
            
            compressed_xref = zlib.compress(xref_data)
            
            # Catalog is likely ID 1.
            xref_stream_dict = f"""<<
/Type /XRef
/Size {xref_oid + 1}
/W [1 4 2]
/Root 1 0 R
/ID [<12345> <12345>]
/Length {len(compressed_xref)}
/Filter /FlateDecode
>>"""
            
            f.write(f"{xref_oid} 0 obj\n".encode('latin1'))
            f.write(xref_stream_dict.encode('latin1'))
            f.write(b"\nstream\n")
            f.write(compressed_xref)
            f.write(b"\nendstream\nendobj\n\n")
            
            f.write(b"startxref\n")
            f.write(f"{startxref_offset}\n".encode('latin1'))
            f.write(b"%%EOF")

def generate_xfdf(filename="medical_data.xfdf"):
    content = """<?xml version="1.0" encoding="UTF-8"?>
<xfdf xmlns="http://ns.adobe.com/xfdf/" xml:space="preserve">
  <fields>
    <field name="patient_name"><value>John Smith</value></field>
    <field name="dob"><value>1980-05-20</value></field>
    <field name="phone"><value>555-0199</value></field>
    <field name="email"><value>john.s@example.com</value></field>
    <field name="insurance"><value>MediCare Plus</value></field>
    <field name="gender"><value>Male</value></field>
    <field name="fever"><value>Yes</value></field>
    <field name="cough"><value>Off</value></field>
    <field name="headache"><value>Yes</value></field>
    <field name="doctor_notes">
        <value>Patient reports high fever for 3 days.
Prescribed antibiotics.
Follow up in 1 week.</value>
    </field>
  </fields>
  <f href="medical_form.pdf"/>
</xfdf>"""
    with open(filename, "w") as f:
        f.write(content)

if __name__ == "__main__":
    print("Generating PDF...")
    pdf = AdvancedPDFGenerator("medical_form.pdf")
    pdf.generate_pdf()
    print("PDF Generated: medical_form.pdf")
    
    print("Generating XFDF...")
    generate_xfdf("medical_data.xfdf")
    print("XFDF Generated: medical_data.xfdf")
