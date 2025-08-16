// Model and serialization: generate JSON, load/save templates, document settings
TemplateEditor.prototype.generateJSON = function() {
    // Reset template structure
    this.template.title = null;
    this.template.table = [];
    this.template.footer = null;

    const elements = this.canvas?.querySelectorAll('.canvas-element') || [];

    elements.forEach(element => {
        const type = element.dataset.type;

        switch (type) {
            case 'title':
                const style = `${element.dataset.bold === 'true' ? '1' : '0'}${element.dataset.italic === 'true' ? '1' : '0'}${element.dataset.underline === 'true' ? '1' : '0'}`;
                this.template.title = {
                    props: `font1:${element.dataset.fontSize}:${style}:${element.dataset.alignment}:0:0:1:0`,
                    text: element.querySelector('.title-text').value
                };
                break;
            case 'table':
                const tableEl = element.querySelector('.template-table');
                const rows = Array.from(tableEl.querySelectorAll('tr'));
                const maxColumns = rows.length > 0 ? Math.max(...rows.map(row => row.children.length)) : 0;

                const tableData = {
                    maxcolumns: maxColumns,
                    rows: rows.map(row => ({
                        row: Array.from(row.children).map(cell => {
                            const input = cell.querySelector('input');
                            const defaultProps = `font1:${element.dataset.cellFontSize}:000:${element.dataset.cellAlignment}:1:1:1:1`;
                            const props = cell.dataset.props || input?.dataset?.props || defaultProps;
                            if (input && input.type === 'checkbox') {
                                return { props: props, chequebox: input.checked };
                            } else {
                                return { props: props, text: input ? input.value : '' };
                            }
                        })
                    }))
                };
                this.template.table.push(tableData);
                break;
            case 'footer':
                this.template.footer = {
                    font: `font1:${element.dataset.fontSize}:001:${element.dataset.alignment}`,
                    text: element.querySelector('.footer-text').value
                };
                break;
            case 'spacer':
                this.template.table.push({
                    maxcolumns: 1,
                    rows: [{ row: [{ props: "font1:12:000:center:0:0:0:0", text: "" }] }]
                });
                break;
        }
    });

    if (this.jsonOutput) this.jsonOutput.value = JSON.stringify(this.template, null, 2);
    document.getElementById('previewPdfBtn') && (document.getElementById('previewPdfBtn').disabled = (elements.length === 0));
};

TemplateEditor.prototype.updateDocumentSettings = function() {
    this.template.config.page = document.getElementById('pageSize').value;
    this.template.config.pageAlignment = parseInt(document.getElementById('pageOrientation').value);
    this.template.config.pageBorder = document.getElementById('pageBorder').value;
    this.template.config.watermark = document.getElementById('watermark').value;
    this.generateJSON();
    this.updateStatus('Document settings updated');
};

TemplateEditor.prototype.loadTemplate = async function() {
    const filename = document.getElementById('templateInput').value.trim();
    if (!filename) return this.updateStatus('Please enter a filename', true);

    try {
        this.updateStatus('Loading template...');
        const response = await fetch(`/api/v1/template-data?file=${encodeURIComponent(filename)}`);
        if (!response.ok) throw new Error(`Template not found or server error (${response.status})`);

        const templateData = await response.json();
        this.loadTemplateData(templateData);
        this.updateStatus(`Template "${filename}" loaded successfully`);
    } catch (error) {
        this.updateStatus(`Error loading template: ${error.message}`, true);
    }
};

TemplateEditor.prototype.loadTemplateData = function(templateData) {
    this.clearCanvas();
    this.template.config = templateData.config || this.template.config;

    document.getElementById('pageSize').value = this.template.config.page || 'A4';
    document.getElementById('pageOrientation').value = this.template.config.pageAlignment || 1;
    document.getElementById('pageBorder').value = this.template.config.pageBorder || '1:1:1:1';
    document.getElementById('watermark').value = this.template.config.watermark || '';

    if (templateData.title) {
        const [ , fontSize, style, alignment] = templateData.title.props.split(':');
        this.createElement('title', {
            text: templateData.title.text,
            fontSize, alignment,
            bold: style[0] === '1',
            italic: style[1] === '1',
            underline: style[2] === '1'
        });
    }

    templateData.table?.forEach(tableData => {
        const firstCellProps = tableData.rows?.[0]?.row?.[0]?.props;
        let cellFontSize = '12', cellAlignment = 'left';
        if (firstCellProps) {
            [, cellFontSize, , cellAlignment] = firstCellProps.split(':');
        }
        this.createElement('table', { ...tableData, cellFontSize, cellAlignment });
    });

    if (templateData.footer) {
        const [, fontSize, , alignment] = templateData.footer.font.split(':');
        this.createElement('footer', {
            text: templateData.footer.text,
            fontSize, alignment
        });
    }

    this.generateJSON();
    if (this.canvas && this.canvas.children.length > 1) this.selectElement(this.canvas.children[1]);
};

TemplateEditor.prototype.showPasteModal = function() {
    document.getElementById('pasteModal').style.display = 'block';
    document.getElementById('pasteJsonArea').focus();
};

TemplateEditor.prototype.hidePasteModal = function() {
    document.getElementById('pasteModal').style.display = 'none';
    document.getElementById('pasteJsonArea').value = '';
};

TemplateEditor.prototype.loadFromPaste = function() {
    const jsonText = document.getElementById('pasteJsonArea').value.trim();
    try {
        const templateData = JSON.parse(jsonText);
        this.loadTemplateData(templateData);
        this.hidePasteModal();
        this.updateStatus('Template loaded from JSON');
    } catch (error) {
        this.updateStatus('Invalid JSON format', true);
    }
};

TemplateEditor.prototype.copyJSON = async function() {
    try {
        await navigator.clipboard.writeText(this.jsonOutput.value);
        this.updateStatus('JSON copied to clipboard');
    } catch (error) {
        this.updateStatus('Failed to copy JSON', true);
    }
};

TemplateEditor.prototype.formatJSON = function() {
    try {
        const parsed = JSON.parse(this.jsonOutput.value);
        this.jsonOutput.value = JSON.stringify(parsed, null, 2);
        this.updateStatus('JSON formatted');
    } catch (error) {
        this.updateStatus('Invalid JSON, cannot format', true);
    }
};

TemplateEditor.prototype.escapeHtml = function(text) {
    if (typeof text !== 'string') return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
};
