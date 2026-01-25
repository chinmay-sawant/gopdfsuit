
import { Type, Table, FileText, Minus, CheckSquare, Circle, Image as ImageIcon, Link } from 'lucide-react'

// Default fonts - Standard PDF Type 1 fonts
export const DEFAULT_FONTS = [
    { id: 'Helvetica', name: 'Helvetica', displayName: 'Helvetica' },
    { id: 'Helvetica-Bold', name: 'Helvetica-Bold', displayName: 'Helvetica Bold' },
    { id: 'Helvetica-Oblique', name: 'Helvetica-Oblique', displayName: 'Helvetica Italic' },
    { id: 'Helvetica-BoldOblique', name: 'Helvetica-BoldOblique', displayName: 'Helvetica Bold Italic' },
    { id: 'Times-Roman', name: 'Times-Roman', displayName: 'Times Roman' },
    { id: 'Times-Bold', name: 'Times-Bold', displayName: 'Times Bold' },
    { id: 'Times-Italic', name: 'Times-Italic', displayName: 'Times Italic' },
    { id: 'Times-BoldItalic', name: 'Times-BoldItalic', displayName: 'Times Bold Italic' },
    { id: 'Courier', name: 'Courier', displayName: 'Courier' },
    { id: 'Courier-Bold', name: 'Courier-Bold', displayName: 'Courier Bold' },
    { id: 'Courier-Oblique', name: 'Courier-Oblique', displayName: 'Courier Italic' },
    { id: 'Courier-BoldOblique', name: 'Courier-BoldOblique', displayName: 'Courier Bold Italic' },
    { id: 'Symbol', name: 'Symbol', displayName: 'Symbol' },
    { id: 'ZapfDingbats', name: 'ZapfDingbats', displayName: 'Zapf Dingbats' }
]

export const PAGE_SIZES = {
    A4: { width: 595, height: 842, name: 'A4' },
    LETTER: { width: 612, height: 792, name: 'Letter' },
    LEGAL: { width: 612, height: 1008, name: 'Legal' },
    A3: { width: 842, height: 1191, name: 'A3' },
    A5: { width: 420, height: 595, name: 'A5' }
}

export const COMPONENT_TYPES = {
    title: { icon: Type, label: 'Title', defaultText: 'Document Title' },
    table: { icon: Table, label: 'Table', rows: 3, cols: 3 },
    footer: { icon: FileText, label: 'Footer', defaultText: 'Page footer text' },
    spacer: { icon: Minus, label: 'Spacer', height: 20 },
    checkbox: { icon: CheckSquare, label: 'Checkbox' },
    radio: { icon: Circle, label: 'Radio Button' },
    text_input: { icon: Type, label: 'Text Input' },
    image: { icon: ImageIcon, label: 'Image' },
    hyperlink: { icon: Link, label: 'Hyperlink' }
}
