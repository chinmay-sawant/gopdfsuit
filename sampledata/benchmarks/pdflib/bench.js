const { PDFDocument, StandardFonts, rgb, PDFName } = require('pdf-lib');
const fs = require('fs');
const { performance } = require('perf_hooks');

const data = JSON.parse(fs.readFileSync('../data.json', 'utf8'));

async function run() {
    const start = performance.now();
    const pdfDoc = await PDFDocument.create();
    
    // Set PDF metadata for improved compliance
    pdfDoc.setTitle('User Report');
    pdfDoc.setAuthor('Benchmark System');
    pdfDoc.setSubject('User data report for benchmarking');
    pdfDoc.setKeywords(['users', 'report', 'benchmark', 'pdf-lib']);
    pdfDoc.setCreator('pdf-lib Benchmark');
    pdfDoc.setProducer('pdf-lib (https://pdf-lib.js.org)');
    pdfDoc.setCreationDate(new Date());
    pdfDoc.setModificationDate(new Date());

    const font = await pdfDoc.embedFont(StandardFonts.Helvetica);
    const fontBold = await pdfDoc.embedFont(StandardFonts.HelveticaBold);
    const fontSize = 10;

    // Column widths proportional to pdfkit format
    const colWidths = [40, 70, 120, 80, 180]; // ID, Name, Email, Role, Description
    const margin = 30;
    const tableWidth = colWidths.reduce((a, b) => a + b, 0);
    const rowHeight = 20;
    const headerBgColor = rgb(0.9, 0.9, 0.9);
    const alternateBgColor = rgb(0.97, 0.97, 0.97);
    
    let page = pdfDoc.addPage([595, 842]); // A4 size
    let { width, height } = page.getSize();
    let y = height - 50;

    // Draw title centered
    const titleText = 'User Report';
    const titleWidth = fontBold.widthOfTextAtSize(titleText, 18);
    page.drawText(titleText, { 
        x: (width - titleWidth) / 2, 
        y: y, 
        size: 18, 
        font: fontBold,
        color: rgb(0, 0, 0)
    });
    y -= 30;

    // Draw table header
    function drawTableRow(page, y, cols, isHeader, isAlternate) {
        let x = margin;
        
        // Draw background
        if (isHeader) {
            page.drawRectangle({
                x: margin,
                y: y - rowHeight + 5,
                width: tableWidth,
                height: rowHeight,
                color: headerBgColor,
            });
        } else if (isAlternate) {
            page.drawRectangle({
                x: margin,
                y: y - rowHeight + 5,
                width: tableWidth,
                height: rowHeight,
                color: alternateBgColor,
            });
        }
        
        // Draw vertical lines and text
        for (let i = 0; i < cols.length; i++) {
            const text = String(cols[i]);
            // Truncate text if too long
            const maxWidth = colWidths[i] - 6;
            let displayText = text;
            const currentFont = isHeader ? fontBold : font;
            
            while (currentFont.widthOfTextAtSize(displayText, fontSize) > maxWidth && displayText.length > 3) {
                displayText = displayText.slice(0, -4) + '...';
            }
            
            page.drawText(displayText, {
                x: x + 3,
                y: y - 12,
                size: fontSize,
                font: currentFont,
                color: rgb(0, 0, 0),
            });
            
            // Draw vertical line
            page.drawLine({
                start: { x: x, y: y + 5 },
                end: { x: x, y: y - rowHeight + 5 },
                thickness: 0.5,
                color: rgb(0.8, 0.8, 0.8),
            });
            
            x += colWidths[i];
        }
        
        // Draw right border
        page.drawLine({
            start: { x: x, y: y + 5 },
            end: { x: x, y: y - rowHeight + 5 },
            thickness: 0.5,
            color: rgb(0.8, 0.8, 0.8),
        });
        
        // Draw horizontal line
        page.drawLine({
            start: { x: margin, y: y - rowHeight + 5 },
            end: { x: margin + tableWidth, y: y - rowHeight + 5 },
            thickness: 0.5,
            color: rgb(0.8, 0.8, 0.8),
        });
    }

    // Draw top border
    page.drawLine({
        start: { x: margin, y: y + 5 },
        end: { x: margin + tableWidth, y: y + 5 },
        thickness: 0.5,
        color: rgb(0.8, 0.8, 0.8),
    });

    // Draw header row
    drawTableRow(page, y, ['ID', 'Name', 'Email', 'Role', 'Description'], true, false);
    y -= rowHeight;

    // Draw data rows
    for (let i = 0; i < data.length; i++) {
        if (y < 60) {
            page = pdfDoc.addPage([595, 842]);
            y = height - 50;
            // Draw top border on new page
            page.drawLine({
                start: { x: margin, y: y + 5 },
                end: { x: margin + tableWidth, y: y + 5 },
                thickness: 0.5,
                color: rgb(0.8, 0.8, 0.8),
            });
        }
        
        const row = data[i];
        drawTableRow(page, y, [row.id, row.name, row.email, row.role, row.desc], false, i % 2 === 1);
        y -= rowHeight;
    }

    // Draw footer
    y -= 20;
    if (y < 40) {
        page = pdfDoc.addPage([595, 842]);
        y = height - 50;
    }
    const footerText = 'Generated by pdf-lib';
    const footerWidth = font.widthOfTextAtSize(footerText, 10);
    page.drawText(footerText, { 
        x: (width - footerWidth) / 2, 
        y: y, 
        size: 10, 
        font: font,
        color: rgb(0.5, 0.5, 0.5)
    });

    // Note: pdf-lib does NOT support PDF/A or PDF/UA standards natively
    // For PDF/A-4 or PDF/UA-2 compliance, you would need a different library
    // such as HummusJS with specialized PDF/A generation capabilities
    
    const pdfBytes = await pdfDoc.save();
    fs.writeFileSync('output_pdflib.pdf', pdfBytes);

    const end = performance.now();
    console.log(`pdf-lib Time: ${(end - start).toFixed(2)} ms`);
    console.log('PDF Standard: PDF 1.7 (pdf-lib does not support PDF/A or PDF/UA natively)');
}

run();
