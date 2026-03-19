'use strict';

const { jsPDF } = require('jspdf');
const fs = require('fs');
const path = require('path');
const { performance } = require('perf_hooks');
const { Worker, isMainThread, workerData, parentPort } = require('worker_threads');

const DATA_PATH = path.resolve(__dirname, '../data.json');
const data = JSON.parse(fs.readFileSync(DATA_PATH, 'utf8'));
const iterations = 10;
const NUM_WORKERS = Math.min(48, iterations);

let autoTable = null;
try {
    autoTable = require('jspdf-autotable');
} catch (e) {
    // optional dependency
}

function buildBody() {
    return data.map((record) => [
        String(record.id),
        record.name,
        record.email,
        record.role,
        record.desc,
    ]);
}

function drawTextFallback(doc) {
    let y = 30;
    const pageHeight = doc.internal.pageSize.height;
    doc.setFontSize(10);
    for (const r of data) {
        if (y > pageHeight - 20) {
            doc.addPage();
            y = 20;
        }
        doc.text(`${r.id} | ${r.name} | ${r.email} | ${r.role}`, 14, y);
        y += 10;
    }
}

function runOnce() {
    const start = performance.now();
    const doc = new jsPDF({ putOnlyUsedFonts: true });
    doc.text('User Report', 14, 20);

    const body = buildBody();
    if (doc.autoTable) {
        doc.autoTable({ startY: 25, head: [['ID', 'Name', 'Email', 'Role', 'Description']], body });
    } else if (autoTable) {
        try {
            const fn = typeof autoTable === 'function' ? autoTable
                : (autoTable.default && typeof autoTable.default === 'function') ? autoTable.default
                : null;
            if (fn) fn(doc, { startY: 25, head: [['ID', 'Name', 'Email', 'Role', 'Description']], body });
            else drawTextFallback(doc);
        } catch (e) {
            drawTextFallback(doc);
        }
    } else {
        drawTextFallback(doc);
    }

    doc.output('arraybuffer'); // Generate bytes in-memory — no disk I/O in hot path
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
        console.log('=== jsPDF Data Benchmark ===');
        console.log(`Iterations: ${iterations} | Workers: ${NUM_WORKERS}`);

        const totalStart = performance.now();
        const results = await Promise.all(
            Array.from({ length: iterations }, (_, i) => spawnWorker(i + 1))
        );
        const totalSeconds = (performance.now() - totalStart) / 1000;

        // Print runs in order
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
    }

    main().catch((err) => { console.error(err); process.exit(1); });
} else {
    // Worker thread: generate one PDF and report elapsed time
    const elapsed = runOnce();
    parentPort.postMessage({ elapsed, runIndex: workerData.runIndex });
}
