// Table-specific behaviors: populate, selection, and row/column ops
TemplateEditor.prototype.populateTable = function(tableElement, tableData) {
    tableElement.innerHTML = '';
    const rows = tableData.rows || [{ row: [{ text: "" }, { text: "" }] }];

    rows.forEach(rowData => {
        const tr = document.createElement('tr');
        rowData.row?.forEach(cellData => {
            const td = document.createElement('td');
            const input = document.createElement('input');
            if (cellData.props) td.dataset.props = cellData.props;
            if (cellData.chequebox !== undefined) {
                input.type = 'checkbox';
                input.checked = cellData.chequebox;
                input.className = 'cell-checkbox';
                if (cellData.props) input.dataset.props = cellData.props;
            } else {
                input.type = 'text';
                input.value = this.escapeHtml(cellData.text || '');
                input.className = 'cell-input';
                if (cellData.props) input.dataset.props = cellData.props;
                // apply alignment and font size from parent element if available
                const parentElement = tableElement.closest('.canvas-element');
                if (parentElement) {
                    const fontSize = parentElement.dataset.cellFontSize || '12';
                    const align = parentElement.dataset.cellAlignment || 'left';
                    input.style.fontSize = fontSize + 'px';
                    input.style.textAlign = align;
                }
            }
            input.addEventListener('input', () => this.generateJSON());
            td.appendChild(input);

            // apply styles (font size, alignment, bold/italic/underline) for this cell
            if (typeof this.applyCellStyles === 'function') this.applyCellStyles(td);

            td.addEventListener('mousedown', (e) => {
                if (e.ctrlKey || e.metaKey) {
                    e.preventDefault();
                    e.stopPropagation();
                    this.isSelectingCells = true;
                    this.toggleCellSelection(td);
                    if (typeof this.applyCellStyles === 'function') this.applyCellStyles(td);
                } else {
                    const parentElement = td.closest('.canvas-element');
                    if (parentElement) this.selectElement(parentElement);
                    this.clearCellSelection();
                    this.selectedCells.add(td);
                    td.classList.add('selected-cell');
                    if (typeof this.applyCellStyles === 'function') this.applyCellStyles(td);
                    setTimeout(() => this.showProperties(this.selectedElement), 0);
                }
            });

            td.addEventListener('mouseenter', (e) => {
                if (this.isSelectingCells && (e.ctrlKey || e.metaKey)) { this.addCellToSelection(td); if (typeof this.applyCellStyles === 'function') this.applyCellStyles(td); }
            });

            td.addEventListener('click', (e) => {
                e.stopPropagation();
                if (this.isSelectingCells) setTimeout(() => this.showProperties(this.selectedElement), 0);
            });

            tr.appendChild(td);
        });
        tableElement.appendChild(tr);
    });
};

TemplateEditor.prototype.toggleCellSelection = function(td) {
    if (this.selectedCells.has(td)) {
        this.selectedCells.delete(td);
        td.classList.remove('selected-cell');
    } else {
        this.selectedCells.add(td);
        td.classList.add('selected-cell');
    }
    setTimeout(() => this.showProperties(this.selectedElement), 0);
};

TemplateEditor.prototype.addCellToSelection = function(td) {
    if (!this.selectedCells.has(td)) {
        this.selectedCells.add(td);
        td.classList.add('selected-cell');
        setTimeout(() => this.showProperties(this.selectedElement), 0);
    }
};

TemplateEditor.prototype.clearCellSelection = function() {
    if (!this.selectedCells) return;
    this.selectedCells.forEach(td => td.classList.remove('selected-cell'));
    this.selectedCells.clear();
    this.isSelectingCells = false;
    if (this.selectedElement?.dataset.type === 'table') setTimeout(() => this.showProperties(this.selectedElement), 0);
};

TemplateEditor.prototype.deleteElement = function(element) {
    if (element) {
        element.remove();
        if (this.selectedElement === element) this.deselectAllElements();
        this.generateJSON();
        this.updateStatus('Element deleted');
        if (this.canvas.querySelectorAll('.canvas-element').length === 0) this.clearCanvas();
    }
};

TemplateEditor.prototype.addTableRow = function() {
    if (!this.selectedElement || this.selectedElement.dataset.type !== 'table') return;
    const table = this.selectedElement.querySelector('.template-table');
    const columnCount = table.querySelector('tr')?.children.length || 2;
    const newRow = table.insertRow();
    for (let i = 0; i < columnCount; i++) {
        const cell = newRow.insertCell();
        const input = document.createElement('input');
        input.type = 'text';
        input.className = 'cell-input';
        input.addEventListener('input', () => this.generateJSON());

            // apply parent table cell styles
            const parentElement = this.selectedElement;
            if (parentElement) {
                const fontSize = parentElement.dataset.cellFontSize || '12';
                const align = parentElement.dataset.cellAlignment || 'left';
                input.style.fontSize = fontSize + 'px';
                input.style.textAlign = align;
            }

        cell.addEventListener('mousedown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                e.preventDefault();
                e.stopPropagation();
                this.isSelectingCells = true;
                this.toggleCellSelection(cell);
            } else {
                this.clearCellSelection();
            }
        });

        cell.addEventListener('mouseenter', (e) => {
            if (this.isSelectingCells && (e.ctrlKey || e.metaKey)) this.addCellToSelection(cell);
        });

        cell.addEventListener('click', (e) => {
            e.stopPropagation();
            if (!this.isSelectingCells) setTimeout(() => this.showProperties(this.selectedElement), 0);
        });

        cell.appendChild(input);
    }
    this.showProperties(this.selectedElement);
    this.generateJSON();
    this.updateStatus('Row added');
};

TemplateEditor.prototype.addTableColumn = function() {
    if (!this.selectedElement || this.selectedElement.dataset.type !== 'table') return;
    const table = this.selectedElement.querySelector('.template-table');
    table.querySelectorAll('tr').forEach(row => {
        const cell = row.insertCell();
        const input = document.createElement('input');
        input.type = 'text';
        input.className = 'cell-input';
        input.addEventListener('input', () => this.generateJSON());

        // apply parent table cell styles
        const parentElement = this.selectedElement;
        if (parentElement) {
            const fontSize = parentElement.dataset.cellFontSize || '12';
            const align = parentElement.dataset.cellAlignment || 'left';
            input.style.fontSize = fontSize + 'px';
            input.style.textAlign = align;
        }

        cell.addEventListener('mousedown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                e.preventDefault();
                e.stopPropagation();
                this.isSelectingCells = true;
                this.toggleCellSelection(cell);
            } else {
                this.clearCellSelection();
            }
        });

        cell.addEventListener('mouseenter', (e) => {
            if (this.isSelectingCells && (e.ctrlKey || e.metaKey)) this.addCellToSelection(cell);
        });

        cell.addEventListener('click', (e) => {
            e.stopPropagation();
            if (!this.isSelectingCells) setTimeout(() => this.showProperties(this.selectedElement), 0);
        });

        cell.appendChild(input);
    });

    this.showProperties(this.selectedElement);
    this.generateJSON();
    this.updateStatus('Column added');
};

TemplateEditor.prototype.removeTableRow = function() {
    if (!this.selectedElement || this.selectedElement.dataset.type !== 'table') return;
    const table = this.selectedElement.querySelector('.template-table');
    if (table.rows.length > 1) {
        table.deleteRow(-1);
        this.showProperties(this.selectedElement);
        this.generateJSON();
        this.updateStatus('Row removed');
    }
};

TemplateEditor.prototype.removeTableColumn = function() {
    if (!this.selectedElement || this.selectedElement.dataset.type !== 'table') return;
    const table = this.selectedElement.querySelector('.template-table');
    const rows = table.querySelectorAll('tr');
    if (rows.length > 0 && rows[0].cells.length > 1) {
        rows.forEach(row => row.deleteCell(-1));
        this.showProperties(this.selectedElement);
        this.generateJSON();
        this.updateStatus('Column removed');
    }
};
