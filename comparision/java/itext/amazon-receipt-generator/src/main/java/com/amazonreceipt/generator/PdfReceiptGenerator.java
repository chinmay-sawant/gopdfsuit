package com.amazonreceipt.generator;

import com.amazonreceipt.model.*;
import com.amazonreceipt.util.ColorPalette;
import com.amazonreceipt.util.FontManager;
import com.itextpdf.kernel.colors.Color;
import com.itextpdf.kernel.geom.PageSize;
import com.itextpdf.kernel.pdf.PdfDocument;
import com.itextpdf.kernel.pdf.PdfWriter;
import com.itextpdf.kernel.pdf.canvas.draw.SolidLine;
import com.itextpdf.layout.Document;
import com.itextpdf.layout.borders.Border;
import com.itextpdf.layout.borders.SolidBorder;
import com.itextpdf.layout.element.*;
import com.itextpdf.layout.properties.HorizontalAlignment;
import com.itextpdf.layout.properties.TextAlignment;
import com.itextpdf.layout.properties.UnitValue;
import com.itextpdf.layout.properties.VerticalAlignment;

import java.io.IOException;

/**
 * PDF Receipt Generator using iText 8.
 * Creates beautiful, professionally styled receipts.
 */
public class PdfReceiptGenerator {

    private final FontManager fontManager;
    private static final float MARGIN = 36f; // 0.5 inch margins
    private static final float TABLE_SPACING = 15f;
    private static final float BORDER_WIDTH = 0.5f;

    public PdfReceiptGenerator() throws IOException {
        this.fontManager = new FontManager();
    }

    /**
     * Generate PDF from receipt data.
     */
    public void generate(ReceiptData data, String outputPath) throws IOException {
        PdfWriter writer = new PdfWriter(outputPath);
        PdfDocument pdfDoc = new PdfDocument(writer);
        
        // Set page size
        PageSize pageSize = getPageSize(data.getConfig());
        pdfDoc.setDefaultPageSize(pageSize);
        
        Document document = new Document(pdfDoc, pageSize);
        document.setMargins(MARGIN, MARGIN, MARGIN, MARGIN);

        // Add title section
        addTitle(document, data.getTitle());

        // Add separator line
        addSeparatorLine(document);

        // Add all tables
        if (data.getTables() != null) {
            for (int i = 0; i < data.getTables().size(); i++) {
                TableSection tableSection = data.getTables().get(i);
                addTable(document, tableSection, i);
                
                // Add spacing between tables (except last)
                if (i < data.getTables().size() - 1) {
                    document.add(new Paragraph().setMarginBottom(TABLE_SPACING));
                }
            }
        }

        // Add footer
        addFooter(document, data.getFooter());

        document.close();
    }

    /**
     * Get page size from config.
     */
    private PageSize getPageSize(ReceiptConfig config) {
        if (config == null || config.getPage() == null) {
            return PageSize.A4;
        }
        
        String page = config.getPage().toUpperCase();
        switch (page) {
            case "LETTER":
                return PageSize.LETTER;
            case "LEGAL":
                return PageSize.LEGAL;
            case "A3":
                return PageSize.A3;
            case "A5":
                return PageSize.A5;
            default:
                return PageSize.A4;
        }
    }

    /**
     * Add title section with Amazon-style branding.
     */
    private void addTitle(Document document, TitleSection title) {
        if (title == null) return;

        // Create a container for the title with background
        Table titleContainer = new Table(UnitValue.createPercentArray(1)).useAllAvailableWidth();
        
        Cell titleCell = new Cell()
            .add(new Paragraph(title.getText())
                .setFont(fontManager.getBoldFont())
                .setFontSize(28f)
                .setFontColor(ColorPalette.AMAZON_ORANGE))
            .setBorder(Border.NO_BORDER)
            .setPaddingBottom(10f);
        
        titleContainer.addCell(titleCell);
        document.add(titleContainer);
    }

    /**
     * Add a decorative separator line.
     */
    private void addSeparatorLine(Document document) {
        SolidLine line = new SolidLine(1f);
        line.setColor(ColorPalette.AMAZON_ORANGE);
        LineSeparator separator = new LineSeparator(line);
        separator.setMarginBottom(15f);
        document.add(separator);
    }

    /**
     * Add a table section with professional styling.
     */
    private void addTable(Document document, TableSection tableSection, int tableIndex) {
        if (tableSection == null || tableSection.getRows() == null) return;

        float[] columnWidths = tableSection.getColumnWidthsAsFloatArray();
        Table table = new Table(UnitValue.createPercentArray(columnWidths)).useAllAvailableWidth();
        table.setMarginBottom(5f);

        boolean isFirstRow = true;
        for (TableRow tableRow : tableSection.getRows()) {
            if (tableRow.getRow() == null) continue;

            for (int colIndex = 0; colIndex < tableRow.getRow().size(); colIndex++) {
                CellData cellData = tableRow.getRow().get(colIndex);
                CellData.CellProperties props = cellData.getParsedProperties();
                
                Cell cell = createStyledCell(cellData, props, isFirstRow, tableIndex);
                table.addCell(cell);
            }
            isFirstRow = false;
        }

        document.add(table);
    }

    /**
     * Create a beautifully styled cell.
     */
    private Cell createStyledCell(CellData cellData, CellData.CellProperties props, 
                                   boolean isHeaderRow, int tableIndex) {
        Cell cell = new Cell();
        
        // Create paragraph with text
        Paragraph paragraph = new Paragraph(cellData.getText() != null ? cellData.getText() : "");
        
        // Set font based on properties
        paragraph.setFont(fontManager.getFont(props.isBold(), props.isItalic()));
        paragraph.setFontSize(props.getFontSize());
        
        // Add underline if needed
        if (props.isUnderline()) {
            paragraph.setUnderline();
        }
        
        // Set text alignment
        TextAlignment alignment = getTextAlignment(props.getAlignment());
        paragraph.setTextAlignment(alignment);
        
        // Apply colors based on row type
        if (isHeaderRow) {
            // Header row styling
            cell.setBackgroundColor(ColorPalette.SECTION_HEADER_BG);
            paragraph.setFontColor(ColorPalette.PRIMARY_DARK);
        } else {
            // Regular row styling
            paragraph.setFontColor(ColorPalette.TEXT_PRIMARY);
            
            // Alternate row background for better readability (subtle)
            if (tableIndex % 2 == 0) {
                cell.setBackgroundColor(ColorPalette.BG_WHITE);
            }
        }
        
        // Apply borders based on properties
        applyBorders(cell, props);
        
        // Padding for better spacing
        cell.setPadding(6f);
        cell.setVerticalAlignment(VerticalAlignment.MIDDLE);
        
        cell.add(paragraph);
        return cell;
    }

    /**
     * Apply borders to cell based on properties.
     */
    private void applyBorders(Cell cell, CellData.CellProperties props) {
        Border border = new SolidBorder(ColorPalette.BORDER_LIGHT, BORDER_WIDTH);
        Border noBorder = Border.NO_BORDER;
        
        cell.setBorderTop(props.hasBorderTop() ? border : noBorder);
        cell.setBorderRight(props.hasBorderRight() ? border : noBorder);
        cell.setBorderBottom(props.hasBorderBottom() ? border : noBorder);
        cell.setBorderLeft(props.hasBorderLeft() ? border : noBorder);
    }

    /**
     * Get text alignment from string.
     */
    private TextAlignment getTextAlignment(String alignment) {
        if (alignment == null) return TextAlignment.LEFT;
        
        String alignLower = alignment.toLowerCase();
        switch (alignLower) {
            case "center":
                return TextAlignment.CENTER;
            case "right":
                return TextAlignment.RIGHT;
            case "justified":
            case "justify":
                return TextAlignment.JUSTIFIED;
            default:
                return TextAlignment.LEFT;
        }
    }

    /**
     * Add footer section.
     */
    private void addFooter(Document document, FooterSection footer) {
        if (footer == null) return;

        // Add some space before footer
        document.add(new Paragraph().setMarginTop(20f));

        // Create footer with styling
        SolidLine line = new SolidLine(0.5f);
        line.setColor(ColorPalette.BORDER_LIGHT);
        LineSeparator separator = new LineSeparator(line);
        document.add(separator);

        Paragraph footerParagraph = new Paragraph(footer.getText())
            .setFont(fontManager.getRegularFont())
            .setFontSize(10f)
            .setFontColor(ColorPalette.TEXT_SECONDARY)
            .setTextAlignment(TextAlignment.CENTER)
            .setMarginTop(10f);

        document.add(footerParagraph);

        // Add generated timestamp
        Paragraph timestamp = new Paragraph("Generated: " + java.time.LocalDateTime.now()
            .format(java.time.format.DateTimeFormatter.ofPattern("MMMM dd, yyyy 'at' HH:mm:ss")))
            .setFont(fontManager.getItalicFont())
            .setFontSize(8f)
            .setFontColor(ColorPalette.TEXT_MUTED)
            .setTextAlignment(TextAlignment.CENTER)
            .setMarginTop(5f);

        document.add(timestamp);
    }
}
