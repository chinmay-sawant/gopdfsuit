// Mirrors test/generate_template-pdf/payload_generator.js trade pools and 80/15/5 mix,
// but emits self-contained HTML for Gotenberg Chromium conversion (no PDF/A or signing).

const SYMBOLS = [
    'RELIANCE', 'TCS', 'INFY', 'HDFCBANK', 'TATASTEEL',
    'ICICIBANK', 'SBIN', 'WIPRO', 'BHARTIARTL', 'LT',
    'NIFTY24FEB22000CE', 'NIFTY24FEB22000PE',
    'BANKNIFTY24FEB46000CE', 'BANKNIFTY24FEB46000PE',
    'AXISBANK', 'KOTAKBANK', 'MARUTI', 'TITAN', 'ADANIENT', 'BAJFINANCE',
];

const ACTIONS = ['BUY', 'SELL'];

const BASE_STYLES = `
  @page { size: A4; margin: 12mm; }
  * { box-sizing: border-box; }
  body { font-family: Helvetica, Arial, sans-serif; margin: 0; color: #1a1a1a; font-size: 9pt; }
  h1 { margin: 0; font-size: 20pt; }
  .header-bar { background: #154360; color: #fff; padding: 12px 16px; display: flex; justify-content: space-between; align-items: center; }
  .header-meta { color: #AED6F1; font-size: 11pt; text-align: right; }
  table { width: 100%; border-collapse: collapse; margin: 8px 0; }
  th, td { border: 1px solid #d5d8dc; padding: 4px 6px; }
  th { background: #D4E6F1; font-weight: 700; text-align: center; }
  .info td.label { background: #EBF5FB; font-weight: 700; width: 18%; }
  .info td.value { background: #EBF5FB; }
  .stripe { background: #F8F9F9; }
  .buy { color: #27AE60; }
  .sell { color: #E74C3C; }
  .total td { background: #A9CCE3; font-weight: 700; }
  .footer { margin-top: 16px; text-align: center; font-size: 7pt; color: #566573; }
  .right { text-align: right; }
  .center { text-align: center; }
`;

function escapeHtml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

function formatINR(amount) {
    // k6/goja has no locale-aware Number.toLocaleString - use fixed decimals like payload_generator.js.
    return `₹${Number(amount).toFixed(2)}`;
}

function generateTrades(n) {
    const trades = [];
    let hour = 9;
    let min = 15;
    let sec = 0;

    for (let i = 0; i < n; i++) {
        const sym = SYMBOLS[Math.floor(Math.random() * SYMBOLS.length)];
        const action = ACTIONS[Math.floor(Math.random() * ACTIONS.length)];
        const qty = (Math.floor(Math.random() * 50) + 1) * 10;
        let price = 100.0 + Math.random() * 3400.0;
        price = Math.round(price * 100) / 100;
        const total = qty * price;
        const timeStr = `${String(hour).padStart(2, '0')}:${String(min).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;

        sec++;
        if (sec >= 60) {
            sec = 0;
            min++;
        }
        if (min >= 60) {
            min = 0;
            hour++;
        }

        trades.push({ id: i + 1, time: timeStr, symbol: sym, action, qty, price, total });
    }
    return trades;
}

function wrapDocument(title, headerRight, bodyHtml, footerText) {
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>${escapeHtml(title)}</title>
  <style>${BASE_STYLES}</style>
</head>
<body>
  <div class="header-bar">
    <h1>${escapeHtml(title)}</h1>
    <div class="header-meta">${escapeHtml(headerRight)}</div>
  </div>
  ${bodyHtml}
  <div class="footer">${escapeHtml(footerText)}</div>
</body>
</html>`;
}

function clientInfoTable(name, code) {
    return `
  <table class="info">
    <tr>
      <td class="label">Client Name:</td>
      <td class="value">${escapeHtml(name)}</td>
      <td class="label">Client Code:</td>
      <td class="value">${escapeHtml(code)}</td>
    </tr>
  </table>`;
}

function tradeRow(cells, stripe) {
    const cls = stripe ? ' class="stripe"' : '';
    return `<tr${cls}>${cells.map((c) => `<td>${c}</td>`).join('')}</tr>`;
}

export function generateRetailHTML() {
    const trades = [
        { symbol: 'TATASTEEL', isin: 'INE081A01012', action: 'BUY', qty: 10, price: 145.5, total: 1455.0 },
        { symbol: 'INFY', isin: 'INE009A01021', action: 'SELL', qty: 5, price: 1650.0, total: 8250.0 },
    ];
    let rows = `
  <tr>
    <th>Symbol</th><th>ISIN</th><th>Action</th><th>Qty</th><th class="right">Price</th><th class="right">Total</th>
  </tr>`;
    trades.forEach((t, i) => {
        const actionClass = t.action === 'SELL' ? 'sell' : 'buy';
        rows += tradeRow([
            `<span class="center">${escapeHtml(t.symbol)}</span>`,
            `<span class="center">${escapeHtml(t.isin)}</span>`,
            `<span class="center ${actionClass}">${escapeHtml(t.action)}</span>`,
            `<span class="center">${t.qty}</span>`,
            `<span class="right">${formatINR(t.price)}</span>`,
            `<span class="right">${formatINR(t.total)}</span>`,
        ], i % 2 !== 0);
    });

    const body = `
  ${clientInfoTable('Rahul Sharma', 'RS9988')}
  <table>${rows}</table>
  <table class="total">
    <tr><td>Total Payable</td><td class="right">${formatINR(6807.5)}</td></tr>
  </table>`;

    return wrapDocument(
        'CONTRACT NOTE',
        'CN2024001 | 2024-02-12',
        body,
        'ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL'
    );
}

export function generateActiveTraderHTML() {
    const trades = generateTrades(40);
    let rows = `
  <tr>
    <th>Symbol</th><th>Action</th><th>Qty</th><th class="right">Price</th><th class="right">Total</th>
  </tr>`;
    let totalTurnover = 0;
    trades.forEach((t, i) => {
        const actionClass = t.action === 'SELL' ? 'sell' : 'buy';
        rows += tradeRow([
            `<span class="center">${escapeHtml(t.symbol)}</span>`,
            `<span class="center ${actionClass}">${escapeHtml(t.action)}</span>`,
            `<span class="center">${t.qty}</span>`,
            `<span class="right">${formatINR(t.price)}</span>`,
            `<span class="right">${formatINR(t.total)}</span>`,
        ], i % 2 !== 0);
        totalTurnover += t.total;
    });

    const body = `
  ${clientInfoTable('Priya Venkatesh', 'PV5544')}
  <table>${rows}</table>
  <table class="total">
    <tr><td>Net Payable</td><td class="right">${formatINR(totalTurnover + 170)}</td></tr>
  </table>`;

    return wrapDocument(
        'ACTIVE TRADER CONTRACT NOTE',
        '40 Trades | 2024-02-12',
        body,
        'ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL'
    );
}

export function generateHFTHTML() {
    const trades = generateTrades(2000);
    let rows = `
  <tr>
    <th>ID</th><th>Time</th><th>Symbol</th><th>Action</th><th>Qty</th><th class="right">Price</th><th class="right">Total</th>
  </tr>`;
    trades.forEach((t, i) => {
        const actionClass = t.action === 'SELL' ? 'sell' : 'buy';
        rows += tradeRow([
            `<span class="center">${t.id}</span>`,
            `<span class="center">${escapeHtml(t.time)}</span>`,
            `<span class="center">${escapeHtml(t.symbol)}</span>`,
            `<span class="center ${actionClass}">${escapeHtml(t.action)}</span>`,
            `<span class="center">${t.qty}</span>`,
            `<span class="right">${formatINR(t.price)}</span>`,
            `<span class="right">${formatINR(t.total)}</span>`,
        ], i % 2 !== 0);
    });

    const body = `
  ${clientInfoTable('Algo Capital LLP', 'HFT001')}
  <table>${rows}</table>`;

    return wrapDocument(
        'HFT CONTRACT NOTE',
        'ALGO CAPITAL LLP | 2,000 Trades',
        body,
        'ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL'
    );
}

/**
 * Same scenario names as gopdfsuit k6 for apples-to-apples env vars.
 * Signing/PDF-A flags are accepted for parity but ignored (Gotenberg renders HTML only).
 */
export function getPayloadOptions(scenario = 'tagged_ecdsa') {
    switch (scenario) {
        case 'retail_only_signed':
            return { tier: 'retail', retailOnly: true };
        case 'retail_active_signed':
            return { tier: 'retail_active', retailActiveOnly: true };
        case 'unsigned':
        case 'tagged_rsa':
        case 'tagged':
        case 'tagged_ecdsa':
        default:
            return { tier: 'weighted' };
    }
}

export function getWeightedHTMLWithTier(opts) {
    const resolved = opts || getPayloadOptions('tagged_ecdsa');
    if (resolved.retailOnly) {
        return { html: generateRetailHTML(), tier: 'retail' };
    }
    if (resolved.retailActiveOnly) {
        const r = Math.random() * 100;
        if (r < 85) {
            return { html: generateRetailHTML(), tier: 'retail' };
        }
        return { html: generateActiveTraderHTML(), tier: 'active' };
    }
    const r = Math.random() * 100;
    if (r < 80) {
        return { html: generateRetailHTML(), tier: 'retail' };
    }
    if (r < 95) {
        return { html: generateActiveTraderHTML(), tier: 'active' };
    }
    return { html: generateHFTHTML(), tier: 'hft' };
}