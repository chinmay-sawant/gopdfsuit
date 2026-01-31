export const examplesSection = {
    title: 'Examples',
    items: [
        {
            id: 'example-financial-report',
            title: 'Financial Report',
            description: 'Complete financial report with sections, tables, and styling.',
            content: `This example demonstrates a professional financial report with:
• Colored section headers
• Multi-column data tables
• Internal navigation links
• Bookmarks for quick navigation
• Footer with company information`,
            code: {
                json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pdfTitle": "Financial Report Q4 2025",
    "pdfaCompliant": true,
    "embedFonts": true
  },
  "title": {
    "props": "Helvetica:24:100:center:0:0:0:0",
    "text": "FINANCIAL REPORT",
    "bgcolor": "#154360",
    "textcolor": "#FFFFFF"
  },
  "elements": [
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:12:100:left:1:1:1:1", "text": "COMPANY INFORMATION", "bgcolor": "#21618C", "textcolor": "#FFFFFF"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 4,
        "columnwidths": [1.2, 2, 1.2, 2],
        "rows": [
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:0:1", "text": "Company:", "bgcolor": "#F4F6F7"},
            {"props": "Helvetica:10:000:left:0:0:0:1", "text": "TechCorp Inc.", "link": "https://techcorp.example.com"},
            {"props": "Helvetica:10:100:left:0:0:0:1", "text": "Period:"},
            {"props": "Helvetica:10:000:left:0:1:0:1", "text": "Q4 2025"}
          ]}
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:12:100:left:1:1:1:1", "text": "FINANCIAL SUMMARY", "bgcolor": "#21618C", "textcolor": "#FFFFFF", "dest": "financial-summary"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 2,
        "columnwidths": [2, 1],
        "rows": [
          {"row": [
            {"props": "Helvetica:10:000:left:1:0:0:1", "text": "Total Revenue"},
            {"props": "Helvetica:10:000:right:0:1:0:1", "text": "$2,450,000"}
          ]},
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:0:1", "text": "Gross Profit", "bgcolor": "#D4E6F1"},
            {"props": "Helvetica:10:100:right:0:1:0:1", "text": "$1,225,000", "bgcolor": "#D4E6F1"}
          ]},
          {"row": [
            {"props": "Helvetica:11:100:left:1:0:1:1", "text": "Net Income", "bgcolor": "#A9CCE3"},
            {"props": "Helvetica:11:100:right:0:1:1:1", "text": "$125,000", "bgcolor": "#A9CCE3"}
          ]}
        ]
      }
    }
  ],
  "footer": {
    "font": "Helvetica:8:000:center",
    "text": "TECHCORP INC. | FINANCIAL REPORT Q4 2025 | CONFIDENTIAL"
  },
  "bookmarks": [
    {"title": "Financial Summary", "dest": "financial-summary"}
  ]
}`
            }
        },
        {
            id: 'example-legal-contract',
            title: 'Legal Contract',
            description: 'Professional services agreement with watermark and signature blocks.',
            content: `This example demonstrates a legal contract with:
• CONFIDENTIAL watermark
• Dual-column party information
• Numbered sections with styling
• Signature blocks`,
            code: {
                json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "watermark": "CONFIDENTIAL",
    "pdfaCompliant": true,
    "embedFonts": true
  },
  "title": {
    "props": "Helvetica:18:100:center:0:0:0:0",
    "text": "PROFESSIONAL SERVICES AGREEMENT",
    "bgcolor": "#2C3E50",
    "textcolor": "#FFFFFF"
  },
  "elements": [
    {
      "type": "table",
      "table": {
        "maxcolumns": 2,
        "bgcolor": "#ECF0F1",
        "rows": [
          {"row": [
            {"props": "Helvetica:12:100:center:1:1:1:1", "text": "CLIENT", "bgcolor": "#34495E", "textcolor": "#FFFFFF"},
            {"props": "Helvetica:12:100:center:1:1:1:1", "text": "PROVIDER", "bgcolor": "#34495E", "textcolor": "#FFFFFF"}
          ]},
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:0:0", "text": "Global Tech Solutions Ltd."},
            {"props": "Helvetica:10:100:left:0:1:0:0", "text": "Apex Legal Consultants"}
          ]},
          {"row": [
            {"props": "Helvetica:10:000:left:1:0:0:1", "text": "Silicon Valley, CA 94025"},
            {"props": "Helvetica:10:000:left:0:1:0:1", "text": "New York, NY 10018"}
          ]}
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:14:100:left:0:0:0:1", "text": "1. SERVICES AND TERM", "textcolor": "#2980B9"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 2,
        "columnwidths": [1, 3],
        "rows": [
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:1:1", "text": "Project:", "bgcolor": "#EBF5FB"},
            {"props": "Helvetica:10:000:left:0:1:0:1", "text": "Enterprise Software Licensing Review"}
          ]},
          {"row": [
            {"props": "Helvetica:10:100:left:1:0:0:1", "text": "Start Date:", "bgcolor": "#EBF5FB"},
            {"props": "Helvetica:10:000:left:0:1:0:1", "text": "November 1, 2023"}
          ]}
        ]
      }
    }
  ],
  "footer": {
    "font": "Helvetica:8:000:center",
    "text": "Service Agreement - Confidential"
  }
}`
            }
        },
        {
            id: 'example-form',
            title: 'Interactive Form',
            description: 'Library form with text inputs, checkboxes, and radio buttons.',
            content: `This example demonstrates an interactive form with:
• Text input fields with default values
• Radio button groups for selections
• Checkboxes for multi-select options
• Color-coded sections`,
            code: {
                json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pageBorder": "1:1:1:1",
    "pdfaCompliant": true,
    "embedFonts": true
  },
  "title": {
    "props": "Helvetica:18:100:center:0:0:0:1",
    "text": "LIBRARY BOOK RECEIVING FORM"
  },
  "elements": [
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:11:100:left:1:1:1:1", "text": "LIBRARY INFORMATION", "bgcolor": "#E8F4FD", "textcolor": "#1565C0"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 4,
        "columnwidths": [1.2, 2, 1.2, 2],
        "rows": [
          {"row": [
            {"props": "Helvetica:9:100:left:1:1:1:1", "text": "Library Name:"},
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "Central City Public Library",
              "form_field": {"type": "text", "name": "library_name", "value": "Central City Public Library"}},
            {"props": "Helvetica:9:100:left:1:1:1:1", "text": "Branch Code:"},
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "CCPL-MAIN-001",
              "form_field": {"type": "text", "name": "branch_code", "value": "CCPL-MAIN-001"}}
          ]}
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:11:100:left:1:1:1:1", "text": "ACQUISITION TYPE", "bgcolor": "#E8F4FD", "textcolor": "#1565C0"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 6,
        "columnwidths": [1.2, 0.4, 1.2, 0.4, 1.2, 0.4],
        "rows": [
          {"row": [
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "Purchase"},
            {"props": "Helvetica:9:000:center:1:1:1:1",
              "form_field": {"type": "radio", "name": "acq_purchase", "value": "purchase", "checked": true, "group_name": "acquisition_type", "shape": "round"}},
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "Donation"},
            {"props": "Helvetica:9:000:center:1:1:1:1",
              "form_field": {"type": "radio", "name": "acq_donation", "value": "donation", "checked": false, "group_name": "acquisition_type", "shape": "round"}},
            {"props": "Helvetica:9:000:left:1:1:1:1", "text": "Exchange"},
            {"props": "Helvetica:9:000:center:1:1:1:1",
              "form_field": {"type": "radio", "name": "acq_exchange", "value": "exchange", "checked": false, "group_name": "acquisition_type", "shape": "round"}}
          ]}
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "rows": [{"row": [
          {"props": "Helvetica:11:100:left:1:1:1:1", "text": "INSPECTION CHECKLIST", "bgcolor": "#E8F4FD", "textcolor": "#1565C0"}
        ]}]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 6,
        "columnwidths": [0.3, 1.5, 0.3, 1.5, 0.3, 1.5],
        "rows": [
          {"row": [
            {"props": "Helvetica:8:000:center:1:1:1:1",
              "form_field": {"type": "checkbox", "name": "check_cover", "value": "Yes", "checked": true}},
            {"props": "Helvetica:8:000:left:1:1:1:1", "text": "Cover Intact"},
            {"props": "Helvetica:8:000:center:1:1:1:1",
              "form_field": {"type": "checkbox", "name": "check_spine", "value": "Yes", "checked": true}},
            {"props": "Helvetica:8:000:left:1:1:1:1", "text": "Spine OK"},
            {"props": "Helvetica:8:000:center:1:1:1:1",
              "form_field": {"type": "checkbox", "name": "check_pages", "value": "Yes", "checked": true}},
            {"props": "Helvetica:8:000:left:1:1:1:1", "text": "All Pages"}
          ]}
        ]
      }
    }
  ],
  "footer": {
    "font": "Helvetica:7:000:center",
    "text": "CENTRAL CITY PUBLIC LIBRARY - BOOK RECEIVING DEPARTMENT"
  }
}`
            }
        }
    ]
};
