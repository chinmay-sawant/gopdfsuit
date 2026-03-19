'use strict';

const { PDFDocument, StandardFonts, rgb } = require('pdf-lib');
const fs = require('fs');
const path = require('path');
const { performance } = require('perf_hooks');
const { Worker, isMainThread, workerData, parentPort } = require('worker_threads');

const DATA_PATH = path.resolve(__dirname, '../data.json');
const data = JSON.parse(fs.readFileSync(DATA_PATH, 'utf8'));
const iterations = 10;
const NUM_WORKERS = Math.min(48, iterations);

async function runOnce() {
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
    // Generate bytes in-memory — no disk I/O in hot path

    return performance.now() - start;
}

if (isMainThread) {
    function spawnWorker(runIndex) {
        return new Promise((resolve, reject) => {
            const worker = new Worker(__filename, { workerData: { runIndex } });
            worker.on('message', (msg) => resolve(msg));
            worker.on('error', reject);
            worker.on('exit', (code) => {
                if (code !== 0) reject(new Error(`Worker ${runIndex} exited with code ${code}`));
            });
        });
    }

    async function main() {
        console.log('=== pdf-lib Data Benchmark ===');
        console.log(`Iterations: ${iterations} | Workers: ${NUM_WORKERS}`);

        const totalStart = performance.now();
        const results = await Promise.all(
            Array.from({ length: iterations }, (_, i) => spawnWorker(i + 1))
        );
        const totalSeconds = (performance.now() - totalStart) / 1000;

        const ordered = results.sort((a, b) => a.runIndex - b.runIndex);
        for (const { runIndex, elapsed } of ordered) {
            console.log(`Run ${runIndex}: ${elapsed.toFixed(2)} ms`);
        }

        const timings = ordered.map((r) => r.elapsed);
        const sorted = [...timings].sort((a, b) => a - b);
        const p95idx = Math.max(0, Math.ceil(sorted.length * 0.95) - 1);

        console.log('');
        console.log(`Min:        ${sorted[0].toFixed(2)} ms`);
        console.log(`Avg:        ${(sorted.reduce((s, v) => s + v, 0) / sorted.length).toFixed(2)} ms`);
        console.log(`P95:        ${sorted[p95idx].toFixed(2)} ms`);
        console.log(`Max:        ${sorted[sorted.length - 1].toFixed(2)} ms`);
        console.log(`Throughput: ${(timings.length / totalSeconds).toFixed(2)} ops/sec`);
        console.log('PDF Standard: PDF 1.7 (pdf-lib does not support PDF/A natively)');
    }

    main().catch((err) => { console.error(err); process.exit(1); });
} else {
    // Worker thread: generate one PDF and report elapsed time
    runOnce()
        .then((elapsed) => {
            parentPort.postMessage({ elapsed, runIndex: workerData.runIndex });
        })
        .catch((err) => {
            console.error(`Worker ${workerData.runIndex} error:`, err);
            process.exit(1);
        });
}
