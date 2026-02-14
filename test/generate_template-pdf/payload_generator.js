import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// -----------------------------------------------------------------------------
// Constants & Pools
// -----------------------------------------------------------------------------

const SYMBOLS = [
    "RELIANCE", "TCS", "INFY", "HDFCBANK", "TATASTEEL",
    "ICICIBANK", "SBIN", "WIPRO", "BHARTIARTL", "LT",
    "NIFTY24FEB22000CE", "NIFTY24FEB22000PE",
    "BANKNIFTY24FEB46000CE", "BANKNIFTY24FEB46000PE",
    "AXISBANK", "KOTAKBANK", "MARUTI", "TITAN", "ADANIENT", "BAJFINANCE",
];

const ACTIONS = ["BUY", "SELL"];

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

function generateTrades(n) {
    const trades = [];
    let hour = 9;
    let min = 15;
    let sec = 0;

    for (let i = 0; i < n; i++) {
        const sym = SYMBOLS[Math.floor(Math.random() * SYMBOLS.length)];
        const action = ACTIONS[Math.floor(Math.random() * ACTIONS.length)];
        const qty = (Math.floor(Math.random() * 50) + 1) * 10; // 10..500
        let price = 100.0 + Math.random() * 3400.0;
        price = Math.round(price * 100) / 100; // round to 2 decimals
        const total = qty * price;

        // Format time
        const timeStr = `${String(hour).padStart(2, '0')}:${String(min).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
        
        // Increment time
        sec++;
        if (sec >= 60) {
            sec = 0;
            min++;
        }
        if (min >= 60) {
            min = 0;
            hour++;
        }

        trades.push({
            id: i + 1,
            time: timeStr,
            symbol: sym,
            action: action,
            qty: qty,
            price: price,
            total: total,
        });
    }
    return trades;
}

// -----------------------------------------------------------------------------
// Templates
// -----------------------------------------------------------------------------

// Template 1: Retail Investor (1 page, ~2 trades)
// Mapped to match financial_digitalsignature.json structure
function generateRetailPayload() {
    return {
        "config": {
            "pageBorder": "0:0:0:0",
            "page": "A4",
            "pageAlignment": 1,
            "pdfTitle": "Contract Note - CN2024001",
            "pdfaCompliant": true,
            "arlingtonCompatible": true,
            "embedFonts": true,
            "signature": {
                "enabled": true,
                "visible": true,
                "name": "Zerodha Compliance",
                "reason": "I am the author of this document",
                "location": "Mumbai, India",
                "contactInfo": "compliance@brokerage.com",
                // Using placeholder certs for test stability, same as original JSON
                "privateKeyPem": "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDFR55rSZyF0oGt\nJhn7kHXowBy+FhZLl7zJhMp7tCJy5rl6yh3xaf0BwNp/j0WDToTayLimpfCWtGrZ\nV5VEjzGMdtD3RvmHWZKMk5SHot80k+FtWVof3M8H5LpLf8Ye7CgfTMk6lsH7uLHI\nresZXF2Vle3KYDCcj/ZtOlamv+5SGOVGOyIXSaamerArUpHkHkirokr1sq8bSWFv\nYyxyzrLJIZ1jqqwNBBMdtQP7MZnDMekveQU5XEWFJz2n7/PmhJu/c+aw2uVAGZ4l\nsDIq99CejE2vmdcTiAZsKY0INfik/2cpwOiKL4AQuNVD0cbwH2paZjTdk5CTimqI\nviCEsQzVAgMBAAECggEAAsYKzFfAzQGnpxQl216T+c2LQE6CyiJ8M5noit9+eH6v\nIhkDXY7vt32YCAd71hvD5gH0CD74m49pZsLcRPywzD72mY0BTYDZ4zYT9lA45fFX\nHJ8PLR5N0guW/u0kXCNWPd82sqctKDY+WAolIW2MantgJRWmUun6cF1/AfspOGg9\npzEroFMILwVaN+yib5MPZxOWG2qxf9jZAJEgn0W5isgOWyL347tgBQHLbqxTm5FF\nh6bz8nqRUwBLYbmcOswSpJZEm3kQGiTyznGiC1NDZMzLHWwZj1Dc/NCR9wVLXROh\nHx4muAq/Zry8mBdED08OIkqoFIKaFCBQRiGLYX+mYQKBgQDNgnailOR8vDXU9qPA\nXhe/Azn2NkLnJqV7wk5WI4aJaf1Ff8ebeZHHTJIvcA93sOAyCxi7pRz3gEUYjKVi\nzoBZHu+3LtHMje8dSjoKEUI7dLZuXCZ9hhoFQGr8nLuxZRCVB5/NPDrjWlJEnqpJ\npQccmxGCoEKyLMEr79DmtRGheQKBgQD1v4mz0+WM7AJi76c3fNSIgi4Zyik7xmw0\naU72wuCUXYfwF4fCh/tIp5FJqUMsYYld+i1jPhXn8zUq74/RDNGuvrbdZTtkYIXs\n/nVMdAhKL2VQ5t4La4bs+ml+6yfSApaUd8tbCt3ttNyqT0kfHP9XisQ6cepnuqPK\nufv5yQBrPQKBgC7jqn/T6wIOy1WI5Lnafh6F9O6ZWNB2v+Ep50e+GU83EKOP0RJH\nPZy0etI6Bj1v7OdeIsmFlcNez+UXChEuPpiW92jbVOEQLVOIgQ+U+oCoU4uAmQOg\n2kUCeqaieCy0e4EVWT+xk1oWXJjtfrsI3UOImgks2arfjT+iGw7Yl2o5AoGAS5pL\ngNlVq48IBOv5o6ZxtDVofWKmYM9ghpdHRb8aXEqSAZkbmQtAkU+L8P9zvPmcyx6m\nS/vTvXIjDzx4IDYzY/EkTORR60WOriRybbzcuAXww3zjHtxLvCglwHgT3hYRwUdB\ndpbXQ8P6hyKxOjMvkv0L9XcKSDMxJLMnA+eEi3kCgYAJB6nr6W6KQRyotXM1gWP+\n9Ff2Zd5figCt/zD61gw5SAhMLMaR+dj/mfSrDu20jXKr4f/WYuQSRZXxSDHk/pDs\nIXdKNOFnoX+EyvIniTXzQsUWdJVdmZdXseclVfKepUCcQZReYLaesQxdDovlFWjC\nEdvw5H4P31EKT6I5hL6jrA==\n-----END PRIVATE KEY-----",
                "certificatePem": "-----BEGIN CERTIFICATE-----\nMIIDaTCCAlECFDkGiITntD1ddujydCgb/KNltr4wMA0GCSqGSIb3DQEBCwUAMHkx\nCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVTdGF0ZTENMAsGA1UEBwwEQ2l0eTESMBAG\nA1UECgwJR29QREZTdWl0MRUwEwYDVQQLDAxJbnRlcm1lZGlhdGUxIDAeBgNVBAMM\nF0dvUERGU3VpdEludGVybWVkaWF0ZUNBMB4XDTI2MDExODA4MDIwOVoXDTI3MDEx\nODA4MDIwOVowaTELMAkGA1UEBhMCVVMxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQH\nDARDaXR5MRIwEAYDVQQKDAlHb1BERlN1aXQxDTALBgNVBAsMBExlYWYxGDAWBgNV\nBAMMD0dvUERGU3VpdFNpZ25lcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC\nggEBAMVHnmtJnIXSga0mGfuQdejAHL4WFkuXvMmEynu0InLmuXrKHfFp/QHA2n+P\nRYNOhNrIuKal8Ja0atlXlUSPMYx20PdG+YdZkoyTlIei3zST4W1ZWh/czwfkukt/\nxh7sKB9MyTqWwfu4scit6xlcXZWV7cpgMJyP9m06Vqa/7lIY5UY7IhdJpqZ6sCtS\nkeQeSKuiSvWyrxtJYW9jLHLOsskhnWOqrA0EEx21A/sxmcMx6S95BTlcRYUnPafv\n8+aEm79z5rDa5UAZniWwMir30J6MTa+Z1xOIBmwpjQg1+KT/ZynA6IovgBC41UPR\nxvAfalpmNN2TkJOKaoi+IISxDNUCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAh4d+\n3PV4bnOsGQIxSupZMIq+qXf1wcB8dLQWX9ILZz9uXho0E1nhDPHZXRvy/mWG3tZD\nedO8vzMhBY5sD2O8O7K7M+khajfG4gfwhCi3H1dTdze4Wq85K2/kNPqQ/d6qmnS4\nDbxIHWrm8p/wU1p4SYfWFijad9UVutaJixCI9FtCPfRYq5+s0c4cRSyKhjfZp6ic\nhQB01AsgOk1iDgQnSvvjwsz0n1BY/+Apnto3k42PYQx+FNIDIeRvtckVoHxWfmMl\ncWsY6Seqg6V41Yuts78fTKlfjhzI7gKdujl7JMtuyLrL3JVP1rZoMXnjf8SK4QAk\nPkJ5eGE0Ht4i9WkakA==\n-----END CERTIFICATE-----",
                "certificateChain": [
                    "-----BEGIN CERTIFICATE-----\nMIIDszCCApugAwIBAgIUGPNNf0kWGKV8Tg0Kvlqt67kM2OkwDQYJKoZIhvcNAQEL\nBQAwaTELMAkGA1UEBhMCVVMxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQHDARDaXR5\nMRIwEAYDVQQKDAlHb1BERlN1aXQxDTALBgNVBAsMBFJvb3QxGDAWBgNVBAMMD0dv\nUERGU3VpdFJvb3RDQTAeFw0yNjAxMTgwODAyMDhaFw0yODExMDcwODAyMDhaMGkx\nCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVTdGF0ZTENMAsGA1UEBwwEQ2l0eTESMBAG\nA1UECgwJR29QREZTdWl0MQ0wCwYDVQQLDARSb290MRgwFgYDVQQDDA9Hb1BERlN1\naXRSb290Q0EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDt+fuF/xXq\n1eZtUjL5PbMGgFatVpE2FAB5upEwehmGRhWo+AMhAXQtCBUsSHMcuCkB+5IQpDPT\nAdZZqni0nnKeKbSL76ryn0EjQHrWVsGa6nddPz1480ZRUXjNbSpmikT5uVc5j1ec\nR3tPw1jtP9B3xjvebEokSLX7Y0nrTPwCQeLIDzpKh80bshvRJ28vmnT38ha4UMOs\nyGV0A70J9ZzUGN9lHM68zDbsbt1ckP9EZRGWRFqjN06vXJpZkLqk/T4LcU+agwK4\n41/fhpMAy3QpYpgC9BNUWAdRzLx/Xl5F8IjGR6vV1dP7O3yKznNEth0ZMSDOsC+n\nX+67D0NLqifLAgMBAAGjUzBRMB0GA1UdDgQWBBQH8KMUDRDASQW01NixSC60o9Y4\ngTAfBgNVHSMEGDAWgBQH8KMUDRDASQW01NixSC60o9Y4gTAPBgNVHRMBAf8EBTAD\nAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQA/ntzzVNBa8bgWO8VigxTsNntGIwn/HR45\n4Og600Ynx+cLQuqIcVwT/stgjg+RO1jBSRSTCtqzbM4/LTgGTbRj4yvgluO6RDdE\n0EsLIioob97jkbLGcMRNGbI4svSBSUytDjhuvmwxz2wBJYGpxZIm6pkgtMeBHrXp\n4750iSj0ORy9TDUUkUdEXfeDBqbjeQ4M1+OaJ5LP3ze09mb1UDGnNKP2nM9m76Pt\ndT/rN+KQKFN48hLnIHMZykEVIoONEzMh3KkfJKhOdTsZrgvwyoLf56qVDCeuADfN\nztHHMRGR4xXSwWkDU/+F00oYhLi63RsFeL4IdGnXb1Tx8VbaPJVm\n-----END CERTIFICATE-----"
                ]
            }
        },
        "title": {
            "props": "Helvetica:24:100:center:0:0:0:0",
            "text": "CONTRACT NOTE",
            "table": {
                "maxcolumns": 2,
                "columnwidths": [1.5, 2.5],
                "rows": [{
                    "row": [
                        { "props": "Helvetica:20:100:center:0:0:0:0", "text": "CONTRACT NOTE", "bgcolor": "#154360", "textcolor": "#FFFFFF", "height": 45 },
                        { "props": "Helvetica:11:000:right:0:0:0:0", "text": "CN2024001 | 2024-02-12", "bgcolor": "#154360", "textcolor": "#AED6F1", "height": 45 }
                    ]
                }]
            }
        },
        "elements": [
            // ... Client Info ...
            {
                "type": "table",
                "table": {
                    "maxcolumns": 4, "columnwidths": [1.2, 2, 1.2, 2],
                    "rows": [
                        { "row": [{ "props": "Helvetica:9:100:left:1:0:0:1", "text": "Client Name:", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:9:000:left:0:0:0:1", "text": "Rahul Sharma", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:9:100:left:0:0:0:1", "text": "Client Code:", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:9:000:left:0:1:0:1", "text": "RS9988", "bgcolor": "#EBF5FB" }] },
                        { "row": [{ "props": "Helvetica:9:100:left:1:0:0:1", "text": "PAN:" }, { "props": "Helvetica:9:000:left:0:0:0:1", "text": "ABCDE1234F" }, { "props": "Helvetica:9:100:left:0:0:0:1", "text": "Trade Date:" }, { "props": "Helvetica:9:000:left:0:1:0:1", "text": "2024-02-12" }] }
                    ]
                }
            },
            // ... Trades ...
            {
                "type": "table",
                "table": {
                    "maxcolumns": 6, "columnwidths": [2, 1.5, 1, 1, 1.5, 1.5],
                    "rows": [
                        { "row": [
                            { "props": "Helvetica:9:100:center:1:0:1:1", "text": "Symbol", "bgcolor": "#D4E6F1" },
                            { "props": "Helvetica:9:100:center:0:0:1:1", "text": "ISIN", "bgcolor": "#D4E6F1" },
                            { "props": "Helvetica:9:100:center:0:0:1:1", "text": "Action", "bgcolor": "#D4E6F1" },
                            { "props": "Helvetica:9:100:center:0:0:1:1", "text": "Qty", "bgcolor": "#D4E6F1" },
                            { "props": "Helvetica:9:100:right:0:0:1:1", "text": "Price", "bgcolor": "#D4E6F1" },
                            { "props": "Helvetica:9:100:right:0:1:1:1", "text": "Total", "bgcolor": "#D4E6F1" }
                        ]},
                        { "row": [
                            { "props": "Helvetica:9:000:center:1:0:0:1", "text": "TATASTEEL" },
                            { "props": "Helvetica:9:000:center:0:0:0:1", "text": "INE081A01012" },
                            { "props": "Helvetica:9:000:center:0:0:0:1", "text": "BUY", "textcolor": "#27AE60" },
                            { "props": "Helvetica:9:000:center:0:0:0:1", "text": "10" },
                            { "props": "Helvetica:9:000:right:0:0:0:1", "text": "₹145.50" },
                            { "props": "Helvetica:9:000:right:0:1:0:1", "text": "₹1,455.00" }
                        ]},
                        { "row": [
                            { "props": "Helvetica:9:000:center:1:0:0:1", "text": "INFY", "bgcolor": "#F8F9F9" },
                            { "props": "Helvetica:9:000:center:0:0:0:1", "text": "INE009A01021", "bgcolor": "#F8F9F9" },
                            { "props": "Helvetica:9:000:center:0:0:0:1", "text": "SELL", "textcolor": "#E74C3C", "bgcolor": "#F8F9F9" },
                            { "props": "Helvetica:9:000:center:0:0:0:1", "text": "5", "bgcolor": "#F8F9F9" },
                            { "props": "Helvetica:9:000:right:0:0:0:1", "text": "₹1,650.00", "bgcolor": "#F8F9F9" },
                            { "props": "Helvetica:9:000:right:0:1:0:1", "text": "₹8,250.00", "bgcolor": "#F8F9F9" }
                        ]}
                    ]
                }
            },
            {
                "type": "table",
                "table": {
                    "maxcolumns": 2, "columnwidths": [2, 1],
                    "rows": [
                        { "row": [{ "props": "Helvetica:10:100:left:1:0:1:1", "text": "Total Payable", "bgcolor": "#A9CCE3" }, { "props": "Helvetica:10:100:right:0:1:1:1", "text": "₹6,807.50", "bgcolor": "#A9CCE3" }] }
                    ]
                }
            }
        ],
        "footer": {
            "font": "Helvetica:7:000:center",
            "text": "ZERODHA BROKING LTD | CONTRACT NOTE | CONFIDENTIAL"
        }
    };
}

// Template 2: Active Trader (40 trades)
function generateActiveTraderPayload() {
    const trades = generateTrades(40);
    const tradeRows = [];
    
    // Header
    tradeRows.push({
        "row": [
            { "props": "Helvetica:8:100:center:1:0:1:1", "text": "Symbol", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:8:100:center:0:0:1:1", "text": "Action", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:8:100:center:0:0:1:1", "text": "Qty", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:8:100:right:0:0:1:1", "text": "Price", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:8:100:right:0:1:1:1", "text": "Total", "bgcolor": "#D4E6F1" }
        ]
    });

    let totalTurnover = 0;
    trades.forEach((t, i) => {
        const bg = i % 2 !== 0 ? "#F8F9F9" : "";
        const actionColor = t.action === "SELL" ? "#E74C3C" : "#27AE60";
        tradeRows.push({
            "row": [
                { "props": "Helvetica:8:000:center:1:0:0:1", "text": t.symbol, "bgcolor": bg },
                { "props": "Helvetica:8:000:center:0:0:0:1", "text": t.action, "textcolor": actionColor, "bgcolor": bg },
                { "props": "Helvetica:8:000:center:0:0:0:1", "text": String(t.qty), "bgcolor": bg },
                { "props": "Helvetica:8:000:right:0:0:0:1", "text": `₹${t.price.toFixed(2)}`, "bgcolor": bg },
                { "props": "Helvetica:8:000:right:0:1:0:1", "text": `₹${t.total.toFixed(2)}`, "bgcolor": bg }
            ]
        });
        totalTurnover += t.total;
    });

    // Make a deep copy of retail payload structure to reuse config
    const payload = generateRetailPayload();
    payload.title.text = "ACTIVE TRADER CONTRACT NOTE";
    payload.title.table.rows[0].row[0].text = "ACTIVE TRADER CONTRACT NOTE";
    payload.title.table.rows[0].row[1].text = "40 Trades | 2024-02-12";
    
    payload.elements = [
        {
            "type": "table",
            "table": {
                "maxcolumns": 4, "columnwidths": [1.2, 2, 1.2, 2],
                "rows": [
                    { "row": [{ "props": "Helvetica:9:100:left:1:0:0:1", "text": "Client Name:", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:9:000:left:0:0:0:1", "text": "Priya Venkatesh", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:9:100:left:0:0:0:1", "text": "Client Code:", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:9:000:left:0:1:0:1", "text": "PV5544", "bgcolor": "#EBF5FB" }] }
                ]
            }
        },
        {
            "type": "table",
            "table": {
                "maxcolumns": 5,
                "columnwidths": [2.5, 1, 1, 1.5, 1.5],
                "rows": tradeRows
            }
        },
        {
            "type": "table",
            "table": {
                "maxcolumns": 2, "columnwidths": [2, 1],
                "rows": [
                    { "row": [{ "props": "Helvetica:10:100:left:1:0:1:1", "text": "Net Payable", "bgcolor": "#A9CCE3" }, { "props": "Helvetica:10:100:right:0:1:1:1", "text": `₹${(totalTurnover + 170).toFixed(2)}`, "bgcolor": "#A9CCE3" }] }
                ]
            }
        }
    ];
    
    return payload;
}

// Template 3: HFT (2000 trades)
function generateHFTPayload() {
    const trades = generateTrades(2000);
    const tradeRows = [];
    
    // Header
    tradeRows.push({
        "row": [
            { "props": "Helvetica:7:100:center:1:0:1:1", "text": "ID", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:7:100:center:0:0:1:1", "text": "Time", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:7:100:center:0:0:1:1", "text": "Symbol", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:7:100:center:0:0:1:1", "text": "Action", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:7:100:center:0:0:1:1", "text": "Qty", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:7:100:right:0:0:1:1", "text": "Price", "bgcolor": "#D4E6F1" },
            { "props": "Helvetica:7:100:right:0:1:1:1", "text": "Total", "bgcolor": "#D4E6F1" }
        ]
    });

    trades.forEach((t, i) => {
        const bg = i % 2 !== 0 ? "#F8F9F9" : "";
        const actionColor = t.action === "SELL" ? "#E74C3C" : "#27AE60";
        tradeRows.push({
            "row": [
                { "props": "Helvetica:7:000:center:1:0:0:1", "text": String(t.id), "bgcolor": bg },
                { "props": "Helvetica:7:000:center:0:0:0:1", "text": t.time, "bgcolor": bg },
                { "props": "Helvetica:7:000:center:0:0:0:1", "text": t.symbol, "bgcolor": bg },
                { "props": "Helvetica:7:000:center:0:0:0:1", "text": t.action, "textcolor": actionColor, "bgcolor": bg },
                { "props": "Helvetica:7:000:center:0:0:0:1", "text": String(t.qty), "bgcolor": bg },
                { "props": "Helvetica:7:000:right:0:0:0:1", "text": `₹${t.price.toFixed(2)}`, "bgcolor": bg },
                { "props": "Helvetica:7:000:right:0:1:0:1", "text": `₹${t.total.toFixed(2)}`, "bgcolor": bg }
            ]
        });
    });

    const payload = generateRetailPayload();
    payload.title.text = "HFT CONTRACT NOTE";
    payload.title.table.rows[0].row[0].text = "HFT CONTRACT NOTE";
    payload.title.table.rows[0].row[1].text = "ALGO CAPITAL LLP | 2,000 Trades";
    
    payload.elements = [
        {
            "type": "table",
            "table": {
                "maxcolumns": 4, "columnwidths": [1.2, 2, 1.2, 2],
                "rows": [
                    { "row": [{ "props": "Helvetica:8:100:left:1:0:0:1", "text": "Client Name:", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:8:000:left:0:0:0:1", "text": "Algo Capital LLP", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:8:100:left:0:0:0:1", "text": "Client Code:", "bgcolor": "#EBF5FB" }, { "props": "Helvetica:8:000:left:0:1:0:1", "text": "HFT001", "bgcolor": "#EBF5FB" }] }
                ]
            }
        },
        {
            "type": "table",
            "table": {
                "maxcolumns": 7,
                "columnwidths": [0.6, 1, 2, 0.8, 0.6, 1.5, 1.5],
                "rows": tradeRows
            }
        }
    ];
    
    return payload;
}

// -----------------------------------------------------------------------------
// Public API
// -----------------------------------------------------------------------------

export function getWeightedPayload() {
    const r = Math.random() * 100;
    if (r < 80) {
        return generateRetailPayload(); // 80%
    } else if (r < 95) {
        return generateActiveTraderPayload(); // 15%
    } else {
        return generateHFTPayload(); // 5%
    }
}
