from typing import List, Optional, Union, Any
from pydantic import BaseModel, Field

# --- Sub-models for Config ---

class SignatureConfig(BaseModel):
    enabled: bool = False
    visible: bool = False
    name: Optional[str] = None
    reason: Optional[str] = None
    location: Optional[str] = None
    contactInfo: Optional[str] = None
    privateKeyPem: Optional[str] = None
    certificatePem: Optional[str] = None
    certificateChain: Optional[List[str]] = None

class SecurityConfig(BaseModel):
    enabled: bool = False
    ownerPassword: Optional[str] = None
    userPassword: Optional[str] = None

class PdfConfig(BaseModel):
    pageBorder: Optional[str] = None
    page: Optional[str] = "A4"
    pageAlignment: int = 1
    watermark: Optional[str] = None
    pdfTitle: Optional[str] = None
    pdfaCompliant: bool = False
    signature: Optional[SignatureConfig] = None
    arlingtonCompatible: bool = False
    embedFonts: bool = False
    security: Optional[SecurityConfig] = None

# --- Tables and Elements ---

class ImageData(BaseModel):
    imagename: str
    imagedata: str  # Base64 encoded or path
    width: Optional[float] = None
    height: Optional[float] = None

class CellItem(BaseModel):
    props: Optional[str] = None
    text: Optional[str] = None
    bgcolor: Optional[str] = None
    textcolor: Optional[str] = None
    height: Optional[float] = None
    width: Optional[float] = None
    link: Optional[str] = None
    dest: Optional[str] = None
    image: Optional[ImageData] = None
    # Add other potential cell properties here as needed

class RowWrapper(BaseModel):
    row: List[CellItem]

class TableConfig(BaseModel):
    type: str = "table"
    maxcolumns: int
    columnwidths: List[float]
    rows: List[RowWrapper]
    rowheights: Optional[List[float]] = None

class SpacerConfig(BaseModel):
    type: str = "spacer"
    height: float

class DividerConfig(BaseModel):
    type: str = "divider"
    thickness: float = 1.0
    color: str = "#000000"
    
class ElementWrapper(BaseModel):
    type: str
    table: Optional[TableConfig] = None
    spacer: Optional[SpacerConfig] = None
    # We can add other types here if they exist in the schema (e.g. divider)

# --- Top Level Sections ---

class PdfTitle(BaseModel):
    props: Optional[str] = None
    text: Optional[str] = None
    table: Optional[TableConfig] = None

class PdfFooter(BaseModel):
    font: Optional[str] = None
    text: Optional[str] = None
    link: Optional[str] = None

class PdfBookmark(BaseModel):
    title: str
    page: int
    dest: Optional[str] = None
    children: Optional[List['PdfBookmark']] = None

# --- Root Request Model ---

class PdfRequest(BaseModel):
    config: PdfConfig
    title: Optional[PdfTitle] = None
    elements: List[ElementWrapper]
    footer: Optional[PdfFooter] = None
    bookmarks: Optional[List[PdfBookmark]] = None
