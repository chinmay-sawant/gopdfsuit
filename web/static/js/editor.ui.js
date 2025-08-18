// UI helpers: styling, theme, status, and small UI utilities
TemplateEditor.prototype.applyTextElementStyles = function(element, inputEl) {
    const fontSize = parseInt(element.dataset.fontSize || '12', 10);
    const alignment = element.dataset.alignment || 'left';
    const isBold = String(element.dataset.bold) === 'true';
    const isItalic = String(element.dataset.italic) === 'true';
    const isUnderline = String(element.dataset.underline) === 'true';

    // Base input styles (match control-group inputs)
    inputEl.style.padding = '8px 12px';
    inputEl.style.border = '2px solid var(--input-border)';
    inputEl.style.borderRadius = '6px';
    inputEl.style.background = 'var(--input-bg)';
    inputEl.style.color = 'var(--text-color)';
    inputEl.style.width = '100%';
    inputEl.style.boxSizing = 'border-box';
    inputEl.style.fontFamily = 'inherit';

    // Text-specific styles
    inputEl.style.fontSize = fontSize + 'px';
    inputEl.style.textAlign = alignment;
    inputEl.style.fontWeight = isBold ? '700' : '400';
    inputEl.style.fontStyle = isItalic ? 'italic' : 'normal';
    inputEl.style.textDecoration = isUnderline ? 'underline' : 'none';
};

// Apply styles to a table cell <td> based on its dataset.props or parent table defaults
TemplateEditor.prototype.applyCellStyles = function(td) {
    if (!td) return;
    const input = td.querySelector('input');
    if (!input) return;

    // props format: font1:<size>:<styleBits>:<alignment>:... where styleBits e.g. 101 => bold, no-italic, underline
    const props = td.dataset.props || input.dataset.props;
    let size = null, align = null, styleBits = null;
    if (props) {
        const parts = props.split(':');
        styleBits = parts[2] || '000';
        size = parts[1] || null;
        align = parts[3] || null;
    }

    const parent = td.closest('.canvas-element');
    if (!size && parent) size = parent.dataset.cellFontSize;
    if (!align && parent) align = parent.dataset.cellAlignment;

    if (size) input.style.fontSize = size + 'px';
    if (align) input.style.textAlign = align;

    // apply style bits
    if (styleBits) {
        const isBold = styleBits.charAt(0) === '1';
        const isItalic = styleBits.charAt(1) === '1';
        const isUnderline = styleBits.charAt(2) === '1';
        input.style.fontWeight = isBold ? '700' : '400';
        input.style.fontStyle = isItalic ? 'italic' : 'normal';
        input.style.textDecoration = isUnderline ? 'underline' : 'none';
    }
};

TemplateEditor.prototype.updateStatus = function(message, isError = false) {
    if (!this.statusMessage) return;
    this.statusMessage.textContent = message;
    this.statusMessage.className = isError ? 'status-error' : 'status-success';
};

TemplateEditor.prototype.clearCanvas = function() {
    if (!this.canvas) return;
    this.canvas.innerHTML = '<div class="canvas-placeholder">Drag elements here to start building</div>';
    this.deselectAllElements();
    this.elementCounter = 0;
    this.generateJSON();
};

TemplateEditor.prototype.clearPlaceholder = function() {
    const placeholder = this.canvas?.querySelector('.canvas-placeholder');
    if (placeholder) placeholder.remove();
};

TemplateEditor.prototype.updateTableTools = function() {
    const isTableSelected = this.selectedElement && this.selectedElement.dataset.type === 'table';
    document.getElementById('addRowBtn') && (document.getElementById('addRowBtn').disabled = !isTableSelected);
    document.getElementById('addColumnBtn') && (document.getElementById('addColumnBtn').disabled = !isTableSelected);
    document.getElementById('removeRowBtn') && (document.getElementById('removeRowBtn').disabled = !isTableSelected);
    document.getElementById('removeColumnBtn') && (document.getElementById('removeColumnBtn').disabled = !isTableSelected);
};

TemplateEditor.prototype.toggleTheme = function() {
    const isDark = document.body.classList.toggle('theme-dark');
    document.body.classList.toggle('theme-light', !isDark);
    const themeToggle = document.getElementById('themeToggle');
    if (themeToggle) themeToggle.textContent = isDark ? '🌞' : '🌙';
    localStorage.setItem('gopdf_editor_theme', isDark ? 'dark' : 'light');
    this.populateGradients(isDark);
    this.applyGradient(true);
};

TemplateEditor.prototype.populateGradients = function(isDark) {
    const gradients = isDark ? this.darkGradients : this.lightGradients;
    const select = document.getElementById('gradientSelect');
    if (!select) return;
    select.innerHTML = '';
    gradients.forEach(gradient => {
        const option = document.createElement('option');
        option.value = gradient.value;
        option.textContent = gradient.name;
        select.appendChild(option);
    });
};

TemplateEditor.prototype.applyGradient = function(themeChanged = false) {
    const select = document.getElementById('gradientSelect');
    if (!select) return;
    if (themeChanged) {
        const isDark = document.body.classList.contains('theme-dark');
        select.value = isDark ? this.darkGradients[0].value : this.lightGradients[0].value;
    }
    const gradient = select.value;
    document.body.style.background = gradient;
    localStorage.setItem('gopdf_editor_gradient', gradient);
};

TemplateEditor.prototype.previewPDF = async function() {
    try {
        this.updateStatus('Generating PDF preview...');
        const response = await fetch('/api/v1/generate/template-pdf', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: this.jsonOutput.value
        });
        if (!response.ok) throw new Error(`Server error: ${response.statusText}`);

        const blob = await response.blob();
        const url = URL.createObjectURL(blob);
        window.open(url, '_blank');
        this.updateStatus('PDF preview opened in a new tab');
    } catch (error) {
        this.updateStatus(`Error generating PDF: ${error.message}`, true);
    }
};
