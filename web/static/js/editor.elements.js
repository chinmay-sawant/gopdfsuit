// Element creation, selection, and property handling
TemplateEditor.prototype.createElement = function(type, data = {}) {
    this.elementCounter++;
    const element = document.createElement('div');
    element.className = 'canvas-element';
    element.dataset.type = type;
    element.dataset.id = `element_${this.elementCounter}`;

    const deleteBtn = document.createElement('button');
    deleteBtn.className = 'element-delete';
    deleteBtn.innerHTML = '&times;';
    deleteBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.deleteElement(element);
    });

    const header = document.createElement('div');
    header.className = 'element-header';

    // Visible type label so users can see what component was added (Title/Table/Footer/Spacer)
    const typeLabel = document.createElement('span');
    // use existing `.element-label` CSS so label is visible and styled; keep `.element-type` for JS targeting
    typeLabel.className = 'element-label element-type';
    // small icon + friendly name
    switch (type) {
        case 'title':
            typeLabel.innerHTML = '<i class="fas fa-heading"></i> Title';
            break;
        case 'table':
            typeLabel.innerHTML = '<i class="fas fa-table"></i> Table';
            break;
        case 'footer':
            typeLabel.innerHTML = '<i class="fas fa-align-center"></i> Footer';
            break;
        case 'spacer':
            typeLabel.innerHTML = '<i class="fas fa-arrows-alt-v"></i> Spacer';
            break;
        default:
            typeLabel.textContent = type;
    }

    const content = document.createElement('div');
    content.className = 'element-content';

    // Minimal wiring for types - consumer code can populate dataset
    switch (type) {
        case 'title':
            element.dataset.fontSize = data.fontSize || data.fontSize || '18';
            element.dataset.alignment = data.alignment || 'center';
            element.dataset.bold = data.bold ? 'true' : 'false';
            element.dataset.italic = data.italic ? 'true' : 'false';
            element.dataset.underline = data.underline ? 'true' : 'false';

            const titleInput = document.createElement('input');
            titleInput.className = 'title-text';
            titleInput.value = data.text || '';
            content.appendChild(titleInput);
            // apply visual styles immediately so alignment/font-size take effect
            if (typeof this.applyTextElementStyles === 'function') this.applyTextElementStyles(element, titleInput);
            break;
        case 'table':
            element.dataset.cellFontSize = data.cellFontSize || '12';
            element.dataset.cellAlignment = data.cellAlignment || 'left';

            const table = document.createElement('table');
            table.className = 'template-table';
            content.appendChild(table);
            this.populateTable(table, data);
            break;
        case 'footer':
            element.dataset.fontSize = data.fontSize || '12';
            element.dataset.alignment = data.alignment || 'center';
            const footerInput = document.createElement('input');
            footerInput.className = 'footer-text';
            footerInput.value = data.text || '';
            content.appendChild(footerInput);
            // apply styles (alignment/font size) for footer immediately
            if (typeof this.applyTextElementStyles === 'function') this.applyTextElementStyles(element, footerInput);
            break;
        case 'spacer':
            // no extra content
            break;
    }

    header.appendChild(typeLabel);
    header.appendChild(deleteBtn);
    element.appendChild(header);
    element.appendChild(content);

    // Add input event listeners for real-time JSON updates
    element.querySelectorAll('input').forEach(input => {
        input.addEventListener('input', () => this.generateJSON());
    });

    element.addEventListener('click', (e) => {
        e.stopPropagation();
        this.selectElement(element);
    });

    this.canvas.appendChild(element);
    this.clearPlaceholder();
    this.selectElement(element);
    this.updateStatus(`Added ${type} element`);
    this.generateJSON();
};

TemplateEditor.prototype.selectElement = function(element) {
    this.deselectAllElements();
    element.classList.add('selected');
    this.selectedElement = element;
    this.showProperties(element);
    this.updateTableTools();
};

TemplateEditor.prototype.deselectAllElements = function() {
    document.querySelectorAll('.canvas-element').forEach(el => el.classList.remove('selected'));
    this.selectedElement = null;
    if (this.propertiesPanel) this.propertiesPanel.innerHTML = '<p class="no-selection">Select an element to edit its properties</p>';
    this.updateTableTools();
    // clear any multi-cell selection state
    if (this.selectedCells && this.selectedCells.size > 0) {
        this.selectedCells.forEach(td => td.classList.remove('selected-cell'));
        this.selectedCells.clear();
    }
};

TemplateEditor.prototype.showProperties = function(element) {
    const type = element.dataset.type;
    let propertiesHTML = '';

    switch (type) {
        case 'title':
            propertiesHTML = `\n                    <h4><i class=\"fas fa-heading\"></i> Title Properties</h4>\n                    <div class=\"property-group\">\n                        <div class=\"table-property-group\">\n                            <label><i class=\"fas fa-text-height\"></i> Font Size:</label>\n                            <input type=\"number\" id=\"propFontSize\" value=\"${element.dataset.fontSize}\" min=\"8\" max=\"72\">\n                        </div>\n                    </div>\n                    <div class=\"property-group\">\n                        <div class=\"table-property-group\">\n                            <label><i class=\"fas fa-align-center\"></i> Alignment:</label>\n                            <select id=\"propAlignment\">\n                                <option value=\"left\" ${element.dataset.alignment === 'left' ? 'selected' : ''}>Left</option>\n                                <option value=\"center\" ${element.dataset.alignment === 'center' ? 'selected' : ''}>Center</option>\n                                <option value=\"right\" ${element.dataset.alignment === 'right' ? 'selected' : ''}>Right</option>\n                            </select>\n                        </div>\n                    </div>\n                    <div class=\"property-group\">\n                        <label><i class=\"fas fa-bold\"></i> Style:</label>\n                        <div class=\"style-checkboxes\">\n                            <label><input type=\"checkbox\" id=\"propBold\" ${element.dataset.bold === 'true' ? 'checked' : ''}> <i class=\"fas fa-bold\"></i> Bold</label>\n                            <label><input type=\"checkbox\" id=\"propItalic\" ${element.dataset.italic === 'true' ? 'checked' : ''}> <i class=\"fas fa-italic\"></i> Italic</label>\n                            <label><input type=\"checkbox\" id=\"propUnderline\" ${element.dataset.underline === 'true' ? 'checked' : ''}> <i class=\"fas fa-underline\"></i> Underline</label>\n                        </div>\n                    </div>\n                `;
            break;

                case 'table':
                        const table = element.querySelector('.template-table');
                        const rows = table.querySelectorAll('tr');
                        const cols = rows.length > 0 ? rows[0].children.length : 0;
                        const selectedCellsCount = this.selectedCells.size;

                        let selectedCountHtml = '';
                        if (selectedCellsCount > 0) {
                                selectedCountHtml = `<div class="stat-item selected-cells-info"><i class="fas fa-check-square"></i><span>Selected: ${selectedCellsCount} cells</span></div>`;
                        }

                        let selectedCellsHtml = '';
                        if (selectedCellsCount > 0) {
                                selectedCellsHtml = `
            <div class="property-group">
                <div class="table-property-group">
                    <label><i class="fas fa-border-style"></i> Selected Cells Border:</label>
                    <div class="border-controls">
                        <div class="border-grid">
                            <div></div>
                            <button type="button" class="border-btn" data-border="top" title="Top Border"><i class="fas fa-minus"></i></button>
                            <div></div>
                            <button type="button" class="border-btn" data-border="left" title="Left Border"><i class="fas fa-minus" style="transform: rotate(90deg);"></i></button>
                            <div class="border-center"><i class="fas fa-th"></i></div>
                            <button type="button" class="border-btn" data-border="right" title="Right Border"><i class="fas fa-minus" style="transform: rotate(90deg);"></i></button>
                            <div></div>
                            <button type="button" class="border-btn" data-border="bottom" title="Bottom Border"><i class="fas fa-minus"></i></button>
                            <div></div>
                        </div>
                        <div class="border-actions">
                            <button type="button" id="clearBordersBtn" class="btn-secondary"><i class="fas fa-eraser"></i> Clear</button>
                            <button type="button" id="allBordersBtn" class="btn-primary"><i class="fas fa-border-all"></i> All</button>
                        </div>
                    </div>
                </div>
            </div>`;
                        } else {
                                selectedCellsHtml = `
            <div class="property-group">
                <div class="table-property-group">
                    <label><i class="fas fa-border-style"></i> Table Border:</label>
                    <button type="button" id="propBorderLeft" class="table-property-btn"><i class="fas fa-minus" style="transform: rotate(90deg);"></i> Left</button>
                    <button type="button" id="propBorderRight" class="table-property-btn"><i class="fas fa-minus" style="transform: rotate(90deg);"></i> Right</button>
                    <button type="button" id="propBorderTop" class="table-property-btn"><i class="fas fa-minus"></i> Top</button>
                    <button type="button" id="propBorderBottom" class="table-property-btn"><i class="fas fa-minus"></i> Bottom</button>
                    <div class="border-hint"><i class="fas fa-info-circle"></i> Select cells with Ctrl+click or Ctrl+drag for individual cell borders</div>
                </div>
            </div>`;
                        }

                        propertiesHTML = `
        <h4><i class="fas fa-table"></i> Table Properties</h4>
        <div class="property-stats">
            <div class="stat-item"><i class="fas fa-columns"></i><span>Columns: ${cols}</span></div>
            <div class="stat-item"><i class="fas fa-grip-lines"></i><span>Rows: ${rows.length}</span></div>
            ${selectedCountHtml}
        </div>

        <div class="property-group">
            <div class="table-property-group">
                <label><i class="fas fa-text-height"></i> Cell Font Size:</label>
                <input type="number" id="propCellFontSize" value="${element.dataset.cellFontSize}" min="6" max="24">
            </div>
        </div>

        <div class="property-group">
            <div class="table-property-group">
                <label><i class="fas fa-align-left"></i> Cell Alignment:</label>
                <select id="propCellAlignment">
                    <option value="left" ${element.dataset.cellAlignment === 'left' ? 'selected' : ''}>Left</option>
                    <option value="center" ${element.dataset.cellAlignment === 'center' ? 'selected' : ''}>Center</option>
                    <option value="right" ${element.dataset.cellAlignment === 'right' ? 'selected' : ''}>Right</option>
                </select>
            </div>
        </div>

        ${selectedCellsHtml}
    `;
            break;

        case 'footer':
            propertiesHTML = `\n                    <h4><i class=\"fas fa-align-center\"></i> Footer Properties</h4>\n                    <div class=\"property-group\">\n                        <div class=\"table-property-group\">\n                            <label><i class=\"fas fa-text-height\"></i> Font Size:</label>\n                            <input type=\"number\" id=\"propFontSize\" value=\"${element.dataset.fontSize}\" min=\"6\" max=\"24\">\n                        </div>\n                    </div>\n                    <div class=\"property-group\">\n                        <div class=\"table-property-group\">\n                            <label><i class=\"fas fa-align-center\"></i> Alignment:</label>\n                            <select id=\"propAlignment\">\n                                <option value=\"left\" ${element.dataset.alignment === 'left' ? 'selected' : ''}>Left</option>\n                                <option value=\"center\" ${element.dataset.alignment === 'center' ? 'selected' : ''}>Center</option>\n                                <option value=\"right\" ${element.dataset.alignment === 'right' ? 'selected' : ''}>Right</option>\n                            </select>\n                        </div>\n                    </div>\n                `;
            break;
    }

    if (this.propertiesPanel) this.propertiesPanel.innerHTML = propertiesHTML;

    if (type === 'table') {
        this.setupBorderControls();
        this.initializeTableBorderButtons();
    }

    if (type === 'table' && this.selectedCells && this.selectedCells.size > 0) {
        const firstCell = this.selectedCells.values().next().value;
        if (firstCell) {
            const input = firstCell.querySelector('input');
            const props = firstCell.dataset.props || input?.dataset?.props;
            if (props) {
                const parts = props.split(':');
                const size = parts[1] || this.selectedElement.dataset.cellFontSize || '12';
                const align = parts[3] || this.selectedElement.dataset.cellAlignment || 'left';
                const fontInput = this.propertiesPanel.querySelector('#propCellFontSize');
                const alignSelect = this.propertiesPanel.querySelector('#propCellAlignment');
                if (fontInput) fontInput.value = size;
                if (alignSelect) alignSelect.value = align;
            }
        }
    }

    this.attachPropertyListeners();
};

TemplateEditor.prototype.setupBorderControls = function() {
    const borderBtns = this.propertiesPanel.querySelectorAll('.border-btn');
    borderBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const borderType = btn.dataset.border;
            this.toggleSelectedCellsBorder(borderType);
            this.updateBorderButtonStates();
        });
    });

    const clearBtn = this.propertiesPanel.querySelector('#clearBordersBtn');
    if (clearBtn) {
        clearBtn.addEventListener('click', () => { this.clearSelectedCellsBorders(); this.updateBorderButtonStates(); });
    }

    const allBtn = this.propertiesPanel.querySelector('#allBordersBtn');
    if (allBtn) {
        allBtn.addEventListener('click', () => { this.setAllSelectedCellsBorders(); this.updateBorderButtonStates(); });
    }

    this.updateBorderButtonStates();
};

TemplateEditor.prototype.setAllSelectedCellsBorders = function() {
    if (this.selectedCells.size === 0) return;

    this.selectedCells.forEach(cell => {
        const input = cell.querySelector('input');
        const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
        const parts = existing.split(':');
        while (parts.length < 8) parts.push('0');
        parts[4] = parts[5] = parts[6] = parts[7] = '1';
        const updated = parts.join(':');
        cell.dataset.props = updated;
        if (input) input.dataset.props = updated;
        cell.style.borderTop = cell.style.borderRight = cell.style.borderBottom = cell.style.borderLeft = '2px solid var(--primary-color)';
    });

    this.generateJSON();
};

TemplateEditor.prototype.initializeTableBorderButtons = function() {
    const buttons = {
        'propBorderLeft': 4,
        'propBorderRight': 5,
        'propBorderTop': 6,
        'propBorderBottom': 7
    };

    Object.entries(buttons).forEach(([buttonId, borderIndex]) => {
        const button = document.getElementById(buttonId);
        if (button) {
            const table = this.selectedElement.querySelector('.template-table');
            const cells = table.querySelectorAll('td');
            let allActive = true;

            cells.forEach(cell => {
                const input = cell.querySelector('input');
                const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
                const parts = existing.split(':');
                while (parts.length < 8) parts.push('0');
                if (parts[borderIndex] !== '1') {
                    allActive = false;
                }
            });

            button.classList.toggle('active', allActive);

            button.addEventListener('click', () => {
                const isActive = button.classList.contains('active');
                this.toggleTableBorder(borderIndex, !isActive);
                button.classList.toggle('active', !isActive);
            });
        }
    });
};

TemplateEditor.prototype.clearSelectedCellsBorders = function() {
    if (this.selectedCells.size === 0) return;

    this.selectedCells.forEach(cell => {
        const input = cell.querySelector('input');
        const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
        const parts = existing.split(':');
        while (parts.length < 8) parts.push('0');
        parts[4] = parts[5] = parts[6] = parts[7] = '0';
        const updated = parts.join(':');
        cell.dataset.props = updated;
        if (input) input.dataset.props = updated;
        cell.style.borderTop = cell.style.borderRight = cell.style.borderBottom = cell.style.borderLeft = '';
    });

    this.generateJSON();
};

TemplateEditor.prototype.toggleTableBorder = function(borderIndex, activate) {
    const table = this.selectedElement.querySelector('.template-table');
    const cells = table.querySelectorAll('td');

    cells.forEach(cell => {
        const input = cell.querySelector('input');
        const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
        const parts = existing.split(':');
        while (parts.length < 8) parts.push('0');
        parts[borderIndex] = activate ? '1' : '0';
        const updated = parts.join(':');
        cell.dataset.props = updated;
        if (input) input.dataset.props = updated;
    });

    this.generateJSON();
};

TemplateEditor.prototype.toggleSelectedCellsBorder = function(borderType) {
    if (this.selectedCells.size === 0) return;
    const borderMap = { left: 4, right: 5, top: 6, bottom: 7 };
    const index = borderMap[borderType];

    let allHaveBorder = true;
    this.selectedCells.forEach(cell => {
        const input = cell.querySelector('input');
        const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
        const parts = existing.split(':');
        while (parts.length < 8) parts.push('0');
        if (parts[index] !== '1') allHaveBorder = false;
    });

    const newState = !allHaveBorder;

    this.selectedCells.forEach(cell => {
        const input = cell.querySelector('input');
        const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
        const parts = existing.split(':');
        while (parts.length < 8) parts.push('0');
        parts[index] = newState ? '1' : '0';
        const updated = parts.join(':');
        cell.dataset.props = updated;
        if (input) input.dataset.props = updated;
        const borderSide = borderType.charAt(0).toUpperCase() + borderType.slice(1);
        cell.style[`border${borderSide}`] = newState ? '2px solid var(--primary-color)' : '';
    });

    this.generateJSON();
};

TemplateEditor.prototype.updateBorderButtonStates = function() {
    if (!this.selectedCells || this.selectedCells.size === 0) return;
    const borderMap = { left: 4, right: 5, top: 6, bottom: 7 };
    const borderBtns = this.propertiesPanel.querySelectorAll('.border-btn');

    borderBtns.forEach(btn => {
        const borderType = btn.dataset.border;
        const index = borderMap[borderType];
        let anySelected = false;
        this.selectedCells.forEach(cell => {
            const input = cell.querySelector('input');
            const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
            const parts = existing.split(':');
            while (parts.length < 8) parts.push('0');
            if (parts[index] === '1') anySelected = true;
        });
        btn.classList.toggle('active', anySelected);
    });
};

TemplateEditor.prototype.attachPropertyListeners = function() {
    if (!this.propertiesPanel) return;
    this.propertiesPanel.querySelectorAll('input, select').forEach(input => {
        input.addEventListener('input', () => this.updateElementFromProperties());
    });
};

TemplateEditor.prototype.updateElementFromProperties = function() {
    if (!this.selectedElement) return;
    const type = this.selectedElement.dataset.type;

    switch (type) {
        case 'title':
        case 'footer':
            this.selectedElement.dataset.fontSize = document.getElementById('propFontSize')?.value;
            this.selectedElement.dataset.alignment = document.getElementById('propAlignment')?.value;
            if (type === 'title') {
                this.selectedElement.dataset.bold = document.getElementById('propBold')?.checked;
                this.selectedElement.dataset.italic = document.getElementById('propItalic')?.checked;
                this.selectedElement.dataset.underline = document.getElementById('propUnderline')?.checked;
            }
            const inputEl = this.selectedElement.querySelector('.title-text, .footer-text');
            if (inputEl) this.applyTextElementStyles(this.selectedElement, inputEl);
            break;
        case 'table':
            const newCellFontSize = document.getElementById('propCellFontSize')?.value;
            const newCellAlignment = document.getElementById('propCellAlignment')?.value;
            if (this.selectedCells && this.selectedCells.size > 0) {
                this.selectedCells.forEach(td => {
                    const input = td.querySelector('input');
                    const existing = td.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:1:1:1:1`;
                    const parts = existing.split(':');
                    while (parts.length < 8) parts.push('0');
                    if (newCellFontSize) parts[1] = newCellFontSize;
                    if (newCellAlignment) parts[3] = newCellAlignment;
                    const updated = parts.join(':');
                    td.dataset.props = updated;
                    if (input) input.dataset.props = updated;
                });
            } else {
                if (newCellFontSize) this.selectedElement.dataset.cellFontSize = newCellFontSize;
                if (newCellAlignment) this.selectedElement.dataset.cellAlignment = newCellAlignment;
                const table = this.selectedElement.querySelector('.template-table');
                table.querySelectorAll('td').forEach(td => {
                    const input = td.querySelector('input');
                    const existing = td.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:1:1:1:1`;
                    const parts = existing.split(':');
                    while (parts.length < 8) parts.push('0');
                    if (newCellFontSize) parts[1] = newCellFontSize;
                    if (newCellAlignment) parts[3] = newCellAlignment;
                    const updated = parts.join(':');
                    td.dataset.props = updated;
                    if (input) input.dataset.props = updated;
                });
            }

            const bt = document.getElementById('propBorderTop')?.checked;
            const br = document.getElementById('propBorderRight')?.checked;
            const bb = document.getElementById('propBorderBottom')?.checked;
            const bl = document.getElementById('propBorderLeft')?.checked;

            const applyBorderToCell = (td) => {
                if (!td) return;
                const input = td.querySelector('input');
                const existing = td.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
                const parts = existing.split(':');
                while (parts.length < 8) parts.push('0');
                parts[4] = bl ? '1' : '0';
                parts[5] = br ? '1' : '0';
                parts[6] = bt ? '1' : '0';
                parts[7] = bb ? '1' : '0';
                const updated = parts.join(':');
                td.dataset.props = updated;
                if (input) input.dataset.props = updated;
                td.style.borderTop = bt ? '2px solid var(--primary-color)' : '';
                td.style.borderRight = br ? '2px solid var(--primary-color)' : '';
                td.style.borderBottom = bb ? '2px solid var(--primary-color)' : '';
                td.style.borderLeft = bl ? '2px solid var(--primary-color)' : '';
            };

            const table = this.selectedElement.querySelector('.template-table');
            if (this.selectedCells.size > 0) {
                this.selectedCells.forEach(cell => applyBorderToCell(cell));
            } else {
                table.querySelectorAll('td').forEach(td => applyBorderToCell(td));
            }
            break;
    }
    this.generateJSON();
};
