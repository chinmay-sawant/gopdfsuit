class TemplateEditor {
    constructor() {
        // Core UI Elements
        this.canvas = document.getElementById('canvas');
        this.propertiesPanel = document.getElementById('propertiesContent');
        this.jsonOutput = document.getElementById('jsonOutput');
        this.statusMessage = document.getElementById('statusMessage');
        
        // State Management
        this.selectedElement = null;
        this.elementCounter = 0;
        this.template = {
            config: {
                pageBorder: "1:1:1:1",
                page: "A4",
                pageAlignment: 1, // 1 for Portrait, 2 for Landscape
                watermark: ""
            },
            title: null,
            table: [],
            footer: null
        };

        // Theme Gradients
        this.lightGradients = [
            { name: 'Default', value: 'linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%)' },
            { name: 'Sunrise', value: 'linear-gradient(120deg, #ff9a9e 0%, #fad0c4 100%)' },
            { name: 'Aqua', value: 'linear-gradient(120deg, #a1c4fd 0%, #c2e9fb 100%)' },
            { name: 'Mint', value: 'linear-gradient(120deg, #d4fc79 0%, #96e6a1 100%)' }
        ];
        this.darkGradients = [
            { name: 'Deep Space', value: 'linear-gradient(135deg, #000428 0%, #004e92 100%)' },
            { name: 'Nightfall', value: 'linear-gradient(135deg, #232526 0%, #414345 100%)' },
            { name: 'Purple Haze', value: 'linear-gradient(135deg, #42275a 0%, #734b6d 100%)' },
            { name: 'Charcoal', value: 'linear-gradient(135deg, #141E30 0%, #243B55 100%)' }
        ];

        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupDragAndDrop();
        this.setupThemeControls();
        this.clearCanvas(); // Sets initial placeholder and state
        this.generateJSON();
        this.checkURLParams(); // Load from URL after setup
        this.updateStatus('Ready - Start building your template');
    }

    checkURLParams() {
        const urlParams = new URLSearchParams(window.location.search);
        const file = urlParams.get('file');
        
        if (file) {
            document.getElementById('templateInput').value = file;
            this.loadTemplate();
        }
    }

    setupEventListeners() {
        // Template controls
        document.getElementById('loadTemplateBtn').addEventListener('click', () => this.loadTemplate());
        document.getElementById('pasteTemplateBtn').addEventListener('click', () => this.showPasteModal());
        document.getElementById('generateTemplateBtn').addEventListener('click', () => this.generateJSON());
        document.getElementById('previewPdfBtn').addEventListener('click', () => this.previewPDF());
        document.getElementById('copyTemplateBtn').addEventListener('click', () => this.copyJSON());
        document.getElementById('formatJsonBtn').addEventListener('click', () => this.formatJSON());

        // Document settings
        document.getElementById('pageSize').addEventListener('change', () => this.updateDocumentSettings());
        document.getElementById('pageOrientation').addEventListener('change', () => this.updateDocumentSettings());
        document.getElementById('pageBorder').addEventListener('input', () => this.updateDocumentSettings());
        document.getElementById('watermark').addEventListener('input', () => this.updateDocumentSettings());

        // Table tools
        document.getElementById('addRowBtn').addEventListener('click', () => this.addTableRow());
        document.getElementById('addColumnBtn').addEventListener('click', () => this.addTableColumn());
        document.getElementById('removeRowBtn').addEventListener('click', () => this.removeTableRow());
        document.getElementById('removeColumnBtn').addEventListener('click', () => this.removeTableColumn());

        // Theme controls
        document.getElementById('themeToggle').addEventListener('click', () => this.toggleTheme());
        document.getElementById('gradientSelect').addEventListener('change', () => this.applyGradient());

        // Modal controls
        document.getElementById('confirmPasteBtn').addEventListener('click', () => this.loadFromPaste());
        document.getElementById('cancelPasteBtn').addEventListener('click', () => this.hidePasteModal());
        document.querySelector('.modal .close').addEventListener('click', () => this.hidePasteModal());

        // Enter key on template input
        document.getElementById('templateInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.loadTemplate();
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                this.generateJSON();
            }
            if (e.key === 'Delete' && this.selectedElement) {
                this.deleteElement(this.selectedElement);
            }
            if (e.key === 'Escape') {
                this.hidePasteModal();
            }
        });

        // Click outside modal to close
        window.addEventListener('click', (e) => {
            const modal = document.getElementById('pasteModal');
            if (e.target === modal) this.hidePasteModal();
        });

        // Deselect element when clicking on canvas background
        this.canvas.addEventListener('click', (e) => {
            if (e.target === this.canvas) this.deselectAllElements();
        });
    }

    setupDragAndDrop() {
        const draggableItems = document.querySelectorAll('.draggable-item');
        
        draggableItems.forEach(item => {
            item.draggable = true;
            item.addEventListener('dragstart', (e) => {
                e.dataTransfer.setData('text/plain', item.dataset.type);
                item.classList.add('dragging');
            });
            item.addEventListener('dragend', () => item.classList.remove('dragging'));
        });

        this.canvas.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.canvas.classList.add('drag-over');
        });
        this.canvas.addEventListener('dragleave', () => this.canvas.classList.remove('drag-over'));
        this.canvas.addEventListener('drop', (e) => {
            e.preventDefault();
            this.canvas.classList.remove('drag-over');
            const elementType = e.dataTransfer.getData('text/plain');
            if (elementType) this.createElement(elementType);
        });
    }

    setupThemeControls() {
        const savedTheme = localStorage.getItem('gopdf_editor_theme') || 'dark';
        const isDark = savedTheme === 'dark';
        document.body.classList.toggle('theme-dark', isDark);
        document.body.classList.toggle('theme-light', !isDark);
        document.getElementById('themeToggle').textContent = isDark ? '🌞' : '🌙';
        
        this.populateGradients(isDark);
        
        const savedGradient = localStorage.getItem('gopdf_editor_gradient');
        if (savedGradient) {
            document.getElementById('gradientSelect').value = savedGradient;
        } else {
            const defaultGradient = isDark ? this.darkGradients[0].value : this.lightGradients[0].value;
            document.getElementById('gradientSelect').value = defaultGradient;
        }
        this.applyGradient();
    }

    createElement(type, data = {}) {
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

        const content = document.createElement('div');
        content.className = 'element-content';

        switch (type) {
            case 'title':
                header.innerHTML = `<i class="fas fa-heading"></i> Title`;
                content.innerHTML = `<input type="text" class="title-text" value="${this.escapeHtml(data.text || 'Document Title')}" placeholder="Enter title text">`;
                // Set default data attributes
                element.dataset.fontSize = data.fontSize || '18';
                element.dataset.alignment = data.alignment || 'center';
                element.dataset.bold = data.bold !== undefined ? data.bold : 'true';
                element.dataset.italic = data.italic || 'false';
                element.dataset.underline = data.underline || 'false';
                break;

            case 'table':
                header.innerHTML = `<i class="fas fa-table"></i> Table`;
                content.innerHTML = `<table class="template-table"></table>`;
                this.populateTable(content.querySelector('table'), data);
                // Set default data attributes for cells
                element.dataset.cellFontSize = data.cellFontSize || '12';
                element.dataset.cellAlignment = data.cellAlignment || 'left';
                break;

            case 'footer':
                header.innerHTML = `<i class="fas fa-align-center"></i> Footer`;
                content.innerHTML = `<input type="text" class="footer-text" value="${this.escapeHtml(data.text || 'Document Footer')}" placeholder="Enter footer text">`;
                // Set default data attributes
                element.dataset.fontSize = data.fontSize || '10';
                element.dataset.alignment = data.alignment || 'center';
                break;
        }

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
    }
    
    selectElement(element) {
        this.deselectAllElements();
        element.classList.add('selected');
        this.selectedElement = element;
        this.showProperties(element);
        this.updateTableTools();
    }
    
    deselectAllElements() {
        document.querySelectorAll('.canvas-element').forEach(el => el.classList.remove('selected'));
        this.selectedElement = null;
        this.propertiesPanel.innerHTML = '<p class="no-selection">Select an element to edit its properties</p>';
        this.updateTableTools();
    }

    showProperties(element) {
        const type = element.dataset.type;
        let propertiesHTML = '';

        switch (type) {
            case 'title':
                propertiesHTML = `
                    <h4>Title Properties</h4>
                    <div class="property-group">
                        <label>Font Size:</label>
                        <input type="number" id="propFontSize" value="${element.dataset.fontSize}" min="8" max="72">
                    </div>
                    <div class="property-group">
                        <label>Alignment:</label>
                        <select id="propAlignment">
                            <option value="left" ${element.dataset.alignment === 'left' ? 'selected' : ''}>Left</option>
                            <option value="center" ${element.dataset.alignment === 'center' ? 'selected' : ''}>Center</option>
                            <option value="right" ${element.dataset.alignment === 'right' ? 'selected' : ''}>Right</option>
                        </select>
                    </div>
                    <div class="property-group">
                        <label>Style:</label>
                        <div class="style-checkboxes">
                            <label><input type="checkbox" id="propBold" ${element.dataset.bold === 'true' ? 'checked' : ''}> Bold</label>
                            <label><input type="checkbox" id="propItalic" ${element.dataset.italic === 'true' ? 'checked' : ''}> Italic</label>
                            <label><input type="checkbox" id="propUnderline" ${element.dataset.underline === 'true' ? 'checked' : ''}> Underline</label>
                        </div>
                    </div>
                `;
                break;

            case 'table':
                const table = element.querySelector('.template-table');
                const rows = table.querySelectorAll('tr');
                const cols = rows.length > 0 ? rows[0].children.length : 0;
                propertiesHTML = `
                    <h4>Table Properties</h4>
                    <div class="property-group"><label>Columns:</label><span>${cols}</span></div>
                    <div class="property-group"><label>Rows:</label><span>${rows.length}</span></div>
                    <div class="property-group">
                        <label>Cell Font Size:</label>
                        <input type="number" id="propCellFontSize" value="${element.dataset.cellFontSize}" min="6" max="24">
                    </div>
                    <div class="property-group">
                        <label>Cell Alignment:</label>
                        <select id="propCellAlignment">
                            <option value="left" ${element.dataset.cellAlignment === 'left' ? 'selected' : ''}>Left</option>
                            <option value="center" ${element.dataset.cellAlignment === 'center' ? 'selected' : ''}>Center</option>
                            <option value="right" ${element.dataset.cellAlignment === 'right' ? 'selected' : ''}>Right</option>
                        </select>
                    </div>
                `;
                break;

            case 'footer':
                propertiesHTML = `
                    <h4>Footer Properties</h4>
                    <div class="property-group">
                        <label>Font Size:</label>
                        <input type="number" id="propFontSize" value="${element.dataset.fontSize}" min="6" max="24">
                    </div>
                    <div class="property-group">
                        <label>Alignment:</label>
                        <select id="propAlignment">
                            <option value="left" ${element.dataset.alignment === 'left' ? 'selected' : ''}>Left</option>
                            <option value="center" ${element.dataset.alignment === 'center' ? 'selected' : ''}>Center</option>
                            <option value="right" ${element.dataset.alignment === 'right' ? 'selected' : ''}>Right</option>
                        </select>
                    </div>
                `;
                break;
        }

        this.propertiesPanel.innerHTML = propertiesHTML;
        this.attachPropertyListeners();
    }

    attachPropertyListeners() {
        this.propertiesPanel.querySelectorAll('input, select').forEach(input => {
            input.addEventListener('input', () => this.updateElementFromProperties());
        });
    }

    updateElementFromProperties() {
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
                break;
            case 'table':
                this.selectedElement.dataset.cellFontSize = document.getElementById('propCellFontSize')?.value;
                this.selectedElement.dataset.cellAlignment = document.getElementById('propCellAlignment')?.value;
                break;
        }
        this.generateJSON();
    }

    updateDocumentSettings() {
        this.template.config.page = document.getElementById('pageSize').value;
        this.template.config.pageAlignment = parseInt(document.getElementById('pageOrientation').value);
        this.template.config.pageBorder = document.getElementById('pageBorder').value;
        this.template.config.watermark = document.getElementById('watermark').value;
        this.generateJSON();
        this.updateStatus('Document settings updated');
    }

    generateJSON() {
        // Reset template structure
        this.template.title = null;
        this.template.table = [];
        this.template.footer = null;

        const elements = this.canvas.querySelectorAll('.canvas-element');
        
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
                                const props = `font1:${element.dataset.cellFontSize}:000:${element.dataset.cellAlignment}:1:1:1:1`;
                                if (input.type === 'checkbox') {
                                    return { props: props, chequebox: input.checked };
                                } else {
                                    return { props: props, text: input.value };
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
            }
        });

        this.jsonOutput.value = JSON.stringify(this.template, null, 2);
        document.getElementById('previewPdfBtn').disabled = elements.length === 0;
    }

    async loadTemplate() {
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
    }

    loadTemplateData(templateData) {
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
        if (this.canvas.children.length > 1) this.selectElement(this.canvas.children[1]);
    }
    
    populateTable(tableElement, tableData) {
        tableElement.innerHTML = ''; // Clear existing content
        const rows = tableData.rows || [{ row: [{ text: "Field:" }, { text: "Value" }] }];

        rows.forEach(rowData => {
            const tr = document.createElement('tr');
            rowData.row?.forEach(cellData => {
                const td = document.createElement('td');
                const input = document.createElement('input');
                if (cellData.chequebox !== undefined) {
                    input.type = 'checkbox';
                    input.checked = cellData.chequebox;
                } else {
                    input.type = 'text';
                    input.value = this.escapeHtml(cellData.text || '');
                    input.className = 'cell-input';
                }
                input.addEventListener('input', () => this.generateJSON());
                td.appendChild(input);
                tr.appendChild(td);
            });
            tableElement.appendChild(tr);
        });
    }

    async previewPDF() {
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
    }

    updateTableTools() {
        const isTableSelected = this.selectedElement?.dataset.type === 'table';
        document.getElementById('addRowBtn').disabled = !isTableSelected;
        document.getElementById('addColumnBtn').disabled = !isTableSelected;
        document.getElementById('removeRowBtn').disabled = !isTableSelected;
        document.getElementById('removeColumnBtn').disabled = !isTableSelected;
    }

    clearCanvas() {
        this.canvas.innerHTML = '<div class="canvas-placeholder"><i class="fas fa-mouse-pointer"></i><p>Drag components from the toolbox to start building your template</p></div>';
        this.deselectAllElements();
    }

    clearPlaceholder() {
        const placeholder = this.canvas.querySelector('.canvas-placeholder');
        if (placeholder) placeholder.remove();
    }

    deleteElement(element) {
        if (element) {
            element.remove();
            if (this.selectedElement === element) this.deselectAllElements();
            this.generateJSON();
            this.updateStatus('Element deleted');
            if (this.canvas.querySelectorAll('.canvas-element').length === 0) {
                this.clearCanvas();
            }
        }
    }
    
    showPasteModal() {
        document.getElementById('pasteModal').style.display = 'block';
        document.getElementById('pasteJsonArea').focus();
    }

    hidePasteModal() {
        document.getElementById('pasteModal').style.display = 'none';
        document.getElementById('pasteJsonArea').value = '';
    }

    loadFromPaste() {
        const jsonText = document.getElementById('pasteJsonArea').value.trim();
        try {
            const templateData = JSON.parse(jsonText);
            this.loadTemplateData(templateData);
            this.hidePasteModal();
            this.updateStatus('Template loaded from JSON');
        } catch (error) {
            this.updateStatus('Invalid JSON format', true);
        }
    }

    async copyJSON() {
        try {
            await navigator.clipboard.writeText(this.jsonOutput.value);
            this.updateStatus('JSON copied to clipboard');
        } catch (error) {
            this.updateStatus('Failed to copy JSON', true);
        }
    }

    formatJSON() {
        try {
            const parsed = JSON.parse(this.jsonOutput.value);
            this.jsonOutput.value = JSON.stringify(parsed, null, 2);
            this.updateStatus('JSON formatted');
        } catch (error) {
            this.updateStatus('Invalid JSON, cannot format', true);
        }
    }

    escapeHtml(text) {
        if (typeof text !== 'string') return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    toggleTheme() {
        const isDark = document.body.classList.toggle('theme-dark');
        document.body.classList.toggle('theme-light', !isDark);
        document.getElementById('themeToggle').textContent = isDark ? '🌞' : '🌙';
        localStorage.setItem('gopdf_editor_theme', isDark ? 'dark' : 'light');
        this.populateGradients(isDark);
        this.applyGradient(true);
    }

    populateGradients(isDark) {
        const gradients = isDark ? this.darkGradients : this.lightGradients;
        const select = document.getElementById('gradientSelect');
        select.innerHTML = '';
        gradients.forEach(gradient => {
            const option = document.createElement('option');
            option.value = gradient.value;
            option.textContent = gradient.name;
            select.appendChild(option);
        });
    }

    applyGradient(themeChanged = false) {
        const select = document.getElementById('gradientSelect');
        if (themeChanged) {
            const isDark = document.body.classList.contains('theme-dark');
            select.value = isDark ? this.darkGradients[0].value : this.lightGradients[0].value;
        }
        const gradient = select.value;
        document.body.style.background = gradient;
        localStorage.setItem('gopdf_editor_gradient', gradient);
    }
    
    updateStatus(message, isError = false) {
        this.statusMessage.textContent = message;
        this.statusMessage.className = isError ? 'status-error' : 'status-success';
    }

    // Table manipulation methods
    addTableRow() {
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
            cell.appendChild(input);
        }
        this.showProperties(this.selectedElement); // To update row count
        this.generateJSON();
        this.updateStatus('Row added');
    }

    addTableColumn() {
        if (!this.selectedElement || this.selectedElement.dataset.type !== 'table') return;
        
        const table = this.selectedElement.querySelector('.template-table');
        table.querySelectorAll('tr').forEach(row => {
            const cell = row.insertCell();
            const input = document.createElement('input');
            input.type = 'text';
            input.className = 'cell-input';
            input.addEventListener('input', () => this.generateJSON());
            cell.appendChild(input);
        });
        
        this.showProperties(this.selectedElement); // To update column count
        this.generateJSON();
        this.updateStatus('Column added');
    }

    removeTableRow() {
        if (!this.selectedElement || this.selectedElement.dataset.type !== 'table') return;
        
        const table = this.selectedElement.querySelector('.template-table');
        if (table.rows.length > 1) {
            table.deleteRow(-1);
            this.showProperties(this.selectedElement);
            this.generateJSON();
            this.updateStatus('Row removed');
        }
    }

    removeTableColumn() {
        if (!this.selectedElement || this.selectedElement.dataset.type !== 'table') return;
        
        const table = this.selectedElement.querySelector('.template-table');
        const rows = table.querySelectorAll('tr');
        if (rows.length > 0 && rows[0].cells.length > 1) {
            rows.forEach(row => row.deleteCell(-1));
            this.showProperties(this.selectedElement);
            this.generateJSON();
            this.updateStatus('Column removed');
        }
    }
}

// Initialize the editor when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new TemplateEditor();
});