import {
  Braces,
  Eraser,
  FileCheck2,
  FileOutput,
  FileText,
  Image,
  Layers3,
  Minimize2,
  Scissors,
  TableProperties,
} from 'lucide-react'

export const toolGroups = [
  {
    title: 'Create',
    items: [
      { to: '/editor', label: 'Template editor', description: 'Build a document template with a visual canvas.', icon: TableProperties },
      { to: '/viewer', label: 'Template preview', description: 'Load JSON, render a PDF, then download it.', icon: FileText },
    ],
  },
  {
    title: 'Work with PDFs',
    items: [
      { to: '/merge', label: 'Merge', description: 'Join several PDFs in the order you choose.', icon: Layers3 },
      { to: '/split', label: 'Split', description: 'Extract page ranges or make smaller documents.', icon: Scissors },
      { to: '/compress', label: 'Compress', description: 'Reduce file size with a clear quality choice.', icon: Minimize2 },
      { to: '/filler', label: 'Fill forms', description: 'Apply XFDF data to an AcroForm PDF.', icon: FileCheck2 },
      { to: '/redact', label: 'Redact', description: 'Mark and remove sensitive document content.', icon: Eraser },
    ],
  },
  {
    title: 'Convert',
    items: [
      { to: '/htmltopdf', label: 'HTML to PDF', description: 'Render HTML into a PDF document.', icon: FileOutput },
      { to: '/htmltoimage', label: 'HTML to image', description: 'Render HTML into a PNG or JPG image.', icon: Image },
      { to: '/viewer', label: 'JSON templates', description: 'Start from a template file or write JSON directly.', icon: Braces },
    ],
  },
]

export const allTools = toolGroups.flatMap((group) => group.items)
