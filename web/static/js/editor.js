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
        // multi-cell selection state for table editing
        this.isSelectingCells = false;
        this.selectedCells = new Set();
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
        this.setupThemeControls(); // Fixed: This function is now defined
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
                // clear multi-cell selection when pressing Escape
                this.clearCellSelection();
            }
        });

        // Stop cell selection on mouseup
        document.addEventListener('mouseup', () => {
            this.isSelectingCells = false;
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

    // --- ADDED: Missing setupThemeControls method ---
    setupThemeControls() {
        const savedTheme = localStorage.getItem('gopdf_editor_theme') || 'light';
        const savedGradient = localStorage.getItem('gopdf_editor_gradient');
        const isDark = savedTheme === 'dark';

        document.body.classList.toggle('theme-dark', isDark);
        document.body.classList.toggle('theme-light', !isDark);
        document.getElementById('themeToggle').textContent = isDark ? '🌞' : '🌙';

        this.populateGradients(isDark);

        const gradientSelect = document.getElementById('gradientSelect');
        if (savedGradient) {
            const currentGradients = isDark ? this.darkGradients : this.lightGradients;
            const gradientExists = currentGradients.some(g => g.value === savedGradient);
            if (gradientExists) {
                gradientSelect.value = savedGradient;
            }
        }
        
        this.applyGradient(); 
    }

    setupDragAndDrop() {
        const draggableItems = document.querySelectorAll('.draggable-item');
        
        draggableItems.forEach(item => {
            item.draggable = true;
            item.addEventListener('dragstart', (e) => {
                const type = item.dataset.type;
                if (type === 'checkbox') {
                    e.dataTransfer.setData('text/cell-component', type);
                } else {
                    e.dataTransfer.setData('text/plain', type);
                }
                item.classList.add('dragging');
            });
            item.addEventListener('dragend', () => item.classList.remove('dragging'));
            // Click the toolbox item to add it directly to the canvas (except checkbox)
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const type = item.dataset.type;
                if (type === 'checkbox') {
                    // If a table is selected, allow quick add to selected cells or first cell
                    if (this.selectedElement && this.selectedElement.dataset.type === 'table') {
                        if (this.selectedCells && this.selectedCells.size > 0) {
                            this.selectedCells.forEach(cell => this.addComponentToCell(cell, 'checkbox'));
                            this.updateStatus('Checkbox added to selected cell(s)');
                        } else {
                            const firstCell = this.selectedElement.querySelector('td');
                            if (firstCell) {
                                this.addComponentToCell(firstCell, 'checkbox');
                                this.updateStatus('Checkbox added to first table cell');
                            } else {
                                this.updateStatus('No table cell found to add checkbox', true);
                            }
                        }
                    } else {
                        // Not added directly; guide the user to drag into a table cell
                        this.updateStatus('Drag the checkbox into a table cell to add it');
                    }
                } else {
                    this.createElement(type);
                }
            });
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

        // Setup cell-specific drag and drop for checkboxes
        this.setupCellDragAndDrop();
    }

    setupCellDragAndDrop() {
        // This will be called dynamically when tables are created
        this.canvas.addEventListener('dragover', (e) => {
            const cell = e.target.closest('td');
            if (cell && e.dataTransfer.types.includes('text/cell-component')) {
                e.preventDefault();
                cell.classList.add('cell-drop-target');
            }
        });

        this.canvas.addEventListener('dragleave', (e) => {
            const cell = e.target.closest('td');
            if (cell) {
                cell.classList.remove('cell-drop-target');
            }
        });

        this.canvas.addEventListener('drop', (e) => {
            const cell = e.target.closest('td');
            if (cell && e.dataTransfer.types.includes('text/cell-component')) {
                e.preventDefault();
                cell.classList.remove('cell-drop-target');
                const componentType = e.dataTransfer.getData('text/cell-component');
                this.addComponentToCell(cell, componentType);
            }
        });
    }

    addComponentToCell(cell, componentType) {
        const input = cell.querySelector('input');
        if (!input) return;

        if (componentType === 'checkbox') {
            input.type = 'checkbox';
            input.checked = false;
            input.className = 'cell-checkbox';
            this.generateJSON();
            this.updateStatus('Checkbox added to cell');
        }
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
                // Apply visual styles to the title input so it matches other inputs/buttons
                // (will be overwritten later by property changes)
                // Use a short timeout to ensure the element is in the DOM when queried
                setTimeout(() => {
                    const tInput = element.querySelector('.title-text');
                    if (tInput) this.applyTextElementStyles(element, tInput);
                }, 0);
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
                // Apply visual styles to the footer input
                setTimeout(() => {
                    const fInput = element.querySelector('.footer-text');
                    if (fInput) this.applyTextElementStyles(element, fInput);
                }, 0);
                break;

            case 'spacer':
                header.innerHTML = `<i class="fas fa-arrows-alt-v"></i> Spacer`;
                // Spacer has no content, it's just for layout
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
        // clear any multi-cell selection state
        if (this.selectedCells && this.selectedCells.size > 0) {
            this.selectedCells.forEach(td => td.classList.remove('selected-cell'));
            this.selectedCells.clear();
        }
    }

    showProperties(element) {
        const type = element.dataset.type;
        let propertiesHTML = '';

        switch (type) {
            case 'title':
                propertiesHTML = `
                    <h4><i class="fas fa-heading"></i> Title Properties</h4>
                    <div class="property-group">
                        <div class="table-property-group">
                            <label><i class="fas fa-text-height"></i> Font Size:</label>
                            <input type="number" id="propFontSize" value="${element.dataset.fontSize}" min="8" max="72">
                        </div>
                    </div>
                    <div class="property-group">
                        <div class="table-property-group">
                            <label><i class="fas fa-align-center"></i> Alignment:</label>
                            <select id="propAlignment">
                                <option value="left" ${element.dataset.alignment === 'left' ? 'selected' : ''}>Left</option>
                                <option value="center" ${element.dataset.alignment === 'center' ? 'selected' : ''}>Center</option>
                                <option value="right" ${element.dataset.alignment === 'right' ? 'selected' : ''}>Right</option>
                            </select>
                        </div>
                    </div>
                    <div class="property-group">
                        <label><i class="fas fa-bold"></i> Style:</label>
                        <div class="style-checkboxes">
                            <label><input type="checkbox" id="propBold" ${element.dataset.bold === 'true' ? 'checked' : ''}> <i class="fas fa-bold"></i> Bold</label>
                            <label><input type="checkbox" id="propItalic" ${element.dataset.italic === 'true' ? 'checked' : ''}> <i class="fas fa-italic"></i> Italic</label>
                            <label><input type="checkbox" id="propUnderline" ${element.dataset.underline === 'true' ? 'checked' : ''}> <i class="fas fa-underline"></i> Underline</label>
                        </div>
                    </div>
                `;
                break;

            case 'table':
                const table = element.querySelector('.template-table');
                const rows = table.querySelectorAll('tr');
                const cols = rows.length > 0 ? rows[0].children.length : 0;
                const selectedCellsCount = this.selectedCells.size;
                
                propertiesHTML = `
                    <h4><i class="fas fa-table"></i> Table Properties</h4>
                    <div class="property-stats">
                        <div class="stat-item">
                            <i class="fas fa-columns"></i>
                            <span>Columns: ${cols}</span>
                        </div>
                        <div class="stat-item">
                            <i class="fas fa-grip-lines"></i>
                            <span>Rows: ${rows.length}</span>
                        </div>
                        ${selectedCellsCount > 0 ? `
                        <div class="stat-item selected-cells-info">
                            <i class="fas fa-check-square"></i>
                            <span>Selected: ${selectedCellsCount} cells</span>
                        </div>` : ''}
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
                    
                    ${selectedCellsCount > 0 ? `
                    <div class="property-group">
                        <div class="table-property-group">
                            <label><i class="fas fa-border-style"></i> Selected Cells Border:</label>
                            <div class="border-controls">
                                <div class="border-grid">
                                    <div></div>
                                    <button type="button" class="border-btn" data-border="top" title="Top Border">
                                        <i class="fas fa-minus"></i>
                                    </button>
                                    <div></div>
                                    <button type="button" class="border-btn" data-border="left" title="Left Border">
                                        <i class="fas fa-minus" style="transform: rotate(90deg);"></i>
                                    </button>
                                    <div class="border-center">
                                        <i class="fas fa-th"></i>
                                    </div>
                                    <button type="button" class="border-btn" data-border="right" title="Right Border">
                                        <i class="fas fa-minus" style="transform: rotate(90deg);"></i>
                                    </button>
                                    <div></div>
                                    <button type="button" class="border-btn" data-border="bottom" title="Bottom Border">
                                        <i class="fas fa-minus"></i>
                                    </button>
                                    <div></div>
                                </div>
                                <div class="border-actions">
                                    <button type="button" id="clearBordersBtn" class="btn-secondary">
                                        <i class="fas fa-eraser"></i> Clear
                                    </button>
                                    <button type="button" id="allBordersBtn" class="btn-primary">
                                        <i class="fas fa-border-all"></i> All
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>` : `
                    <div class="property-group">
                        <div class="table-property-group">
                            <label><i class="fas fa-border-style"></i> Table Border:</label>
                            <button type="button" id="propBorderLeft" class="table-property-btn">
                                <i class="fas fa-minus" style="transform: rotate(90deg);"></i> Left
                            </button>
                            <button type="button" id="propBorderRight" class="table-property-btn">
                                <i class="fas fa-minus" style="transform: rotate(90deg);"></i> Right
                            </button>
                            <button type="button" id="propBorderTop" class="table-property-btn">
                                <i class="fas fa-minus"></i> Top
                            </button>
                            <button type="button" id="propBorderBottom" class="table-property-btn">
                                <i class="fas fa-minus"></i> Bottom
                            </button>
                            <div class="border-hint">
                                <i class="fas fa-info-circle"></i>
                                Select cells with Ctrl+click or Ctrl+drag for individual cell borders
                            </div>
                        </div>
                    </div>`
                }
                `;
                break;

            case 'footer':
                propertiesHTML = `
                    <h4><i class="fas fa-align-center"></i> Footer Properties</h4>
                    <div class="property-group">
                        <div class="table-property-group">
                            <label><i class="fas fa-text-height"></i> Font Size:</label>
                            <input type="number" id="propFontSize" value="${element.dataset.fontSize}" min="6" max="24">
                        </div>
                    </div>
                    <div class="property-group">
                        <div class="table-property-group">
                            <label><i class="fas fa-align-center"></i> Alignment:</label>
                            <select id="propAlignment">
                                <option value="left" ${element.dataset.alignment === 'left' ? 'selected' : ''}>Left</option>
                                <option value="center" ${element.dataset.alignment === 'center' ? 'selected' : ''}>Center</option>
                                <option value="right" ${element.dataset.alignment === 'right' ? 'selected' : ''}>Right</option>
                            </select>
                        </div>
                    </div>
                `;
                break;
        }

        this.propertiesPanel.innerHTML = propertiesHTML;

        if (type === 'table') {
            this.setupBorderControls();
            this.initializeTableBorderButtons();
        }

        this.attachPropertyListeners();
    }
setupBorderControls() {
    // Border button controls for selected cells
    const borderBtns = this.propertiesPanel.querySelectorAll('.border-btn');
    borderBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const borderType = btn.dataset.border;
            this.toggleSelectedCellsBorder(borderType);
            this.updateBorderButtonStates();
        });
    });

    // Clear all borders button
    const clearBtn = this.propertiesPanel.querySelector('#clearBordersBtn');
    if (clearBtn) {
        clearBtn.addEventListener('click', () => {
            this.clearSelectedCellsBorders();
            this.updateBorderButtonStates();
        });
    }

    // All borders button
    const allBtn = this.propertiesPanel.querySelector('#allBordersBtn');
    if (allBtn) {
        allBtn.addEventListener('click', () => {
            this.setAllSelectedCellsBorders();
            this.updateBorderButtonStates();
        });
    }

    // Update initial states
    this.updateBorderButtonStates();
}

setAllSelectedCellsBorders() {
    if (this.selectedCells.size === 0) return;

    this.selectedCells.forEach(cell => {
        const input = cell.querySelector('input');
        const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
        const parts = existing.split(':');
        while (parts.length < 8) parts.push('0');
    // set left,right,top,bottom = 1
    parts[4] = parts[5] = parts[6] = parts[7] = '1';
        const updated = parts.join(':');
        cell.dataset.props = updated;
        if (input) input.dataset.props = updated;
        cell.style.borderTop = cell.style.borderRight = cell.style.borderBottom = cell.style.borderLeft = '2px solid var(--primary-color)';
    });

    this.generateJSON();
}

    initializeTableBorderButtons() {
        const buttons = {
            // Use props order left(4), right(5), top(6), bottom(7)
            'propBorderLeft': 4,
            'propBorderRight': 5,
            'propBorderTop': 6,
            'propBorderBottom': 7
        };

        Object.entries(buttons).forEach(([buttonId, borderIndex]) => {
            const button = document.getElementById(buttonId);
            if (button) {
                // Check if this border is active for all table cells
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
    }
clearSelectedCellsBorders() {
        if (this.selectedCells.size === 0) return;

        this.selectedCells.forEach(cell => {
            const input = cell.querySelector('input');
            const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
            const parts = existing.split(':');
            while (parts.length < 8) parts.push('0');
            
            // clear left,right,top,bottom
            parts[4] = parts[5] = parts[6] = parts[7] = '0';
            const updated = parts.join(':');
            cell.dataset.props = updated;
            if (input) input.dataset.props = updated;
            
            // Clear visual feedback
            cell.style.borderTop = cell.style.borderRight = cell.style.borderBottom = cell.style.borderLeft = '';
        });

        this.generateJSON();
    }
    
    toggleTableBorder(borderIndex, activate) {
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
    }

    toggleSelectedCellsBorder(borderType) {
        if (this.selectedCells.size === 0) return;

    // props order: left(4), right(5), top(6), bottom(7)
    const borderMap = { left: 4, right: 5, top: 6, bottom: 7 };
        const index = borderMap[borderType];

        // Check if all selected cells have this border
        let allHaveBorder = true;
        this.selectedCells.forEach(cell => {
            const input = cell.querySelector('input');
            const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
            const parts = existing.split(':');
            while (parts.length < 8) parts.push('0');
            if (parts[index] !== '1') {
                allHaveBorder = false;
            }
        });

        // Toggle based on current state
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
            
            // Visual feedback
            const borderSide = borderType.charAt(0).toUpperCase() + borderType.slice(1);
            cell.style[`border${borderSide}`] = newState ? '2px solid var(--primary-color)' : '';
        });

        this.generateJSON();
    }

    updateBorderButtonStates() {
        if (this.selectedCells.size === 0) return;
    // props order: left(4), right(5), top(6), bottom(7)
    const borderMap = { left: 4, right: 5, top: 6, bottom: 7 };
        const borderBtns = this.propertiesPanel.querySelectorAll('.border-btn');

        borderBtns.forEach(btn => {
            const borderType = btn.dataset.border;
            const index = borderMap[borderType];
            // Mark as active if ANY selected cell has this border (more intuitive for mixed selections)
            let anySelected = false;
            this.selectedCells.forEach(cell => {
                const input = cell.querySelector('input');
                const existing = cell.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
                const parts = existing.split(':');
                while (parts.length < 8) parts.push('0');
                if (parts[index] === '1') {
                    anySelected = true;
                }
            });

            btn.classList.toggle('active', anySelected);
        });
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
                // Update visual appearance of the title/footer input to match properties
                const inputEl = this.selectedElement.querySelector('.title-text, .footer-text');
                if (inputEl) this.applyTextElementStyles(this.selectedElement, inputEl);
                break;
            case 'table':
                this.selectedElement.dataset.cellFontSize = document.getElementById('propCellFontSize')?.value;
                this.selectedElement.dataset.cellAlignment = document.getElementById('propCellAlignment')?.value;
                // Border flags (top,right,bottom,left) apply to selected cells or all cells
                const bt = document.getElementById('propBorderTop')?.checked;
                const br = document.getElementById('propBorderRight')?.checked;
                const bb = document.getElementById('propBorderBottom')?.checked;
                const bl = document.getElementById('propBorderLeft')?.checked;

                        // Helper to apply border flags into the props string (preserve other parts)
                        const applyBorderToCell = (td) => {
                            if (!td) return;
                            // get existing props from td or input, or default
                            const input = td.querySelector('input');
                            const existing = td.dataset.props || input?.dataset?.props || `font1:${this.selectedElement.dataset.cellFontSize}:000:${this.selectedElement.dataset.cellAlignment}:0:0:0:0`;
                            const parts = existing.split(':');
                            // parts expected like [font1, size, style, align, left,right,top,bottom]
                            while (parts.length < 8) parts.push('0');
                            // map flags: left(index4), right(5), top(6), bottom(7)
                            parts[4] = bl ? '1' : '0';
                            parts[5] = br ? '1' : '0';
                            parts[6] = bt ? '1' : '0';
                            parts[7] = bb ? '1' : '0';
                            const updated = parts.join(':');
                            td.dataset.props = updated;
                            if (input) input.dataset.props = updated;
                            // visual feedback: toggle CSS border styles
                            td.style.borderTop = bt ? '2px solid var(--primary-color)' : '';
                            td.style.borderRight = br ? '2px solid var(--primary-color)' : '';
                            td.style.borderBottom = bb ? '2px solid var(--primary-color)' : '';
                            td.style.borderLeft = bl ? '2px solid var(--primary-color)' : '';
                        };

                const table = this.selectedElement.querySelector('.template-table');
                if (this.selectedCells.size > 0) {
                    this.selectedCells.forEach(cell => applyBorderToCell(cell));
                } else {
                    // apply to all cells
                    table.querySelectorAll('td').forEach(td => applyBorderToCell(td));
                }
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

    // Apply visual styles to title/footer inputs so they match other form controls
    applyTextElementStyles(element, inputEl) {
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
                                // Prefer props stored on the cell (from loaded template) or on the input
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
                    // Add a dummy table with one empty row to create space
                    this.template.table.push({
                        maxcolumns: 1,
                        rows: [{ row: [{ props: "font1:12:000:center:0:0:0:0", text: "" }] }]
                    });
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
        const rows = tableData.rows || [{ row: [{ text: "" }, { text: "" }] }];

        rows.forEach(rowData => {
            const tr = document.createElement('tr');
            rowData.row?.forEach(cellData => {
                const td = document.createElement('td');
                const input = document.createElement('input');
                // preserve any per-cell props on the td for later serialization
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
                }
                input.addEventListener('input', () => this.generateJSON());
                td.appendChild(input);
                
                // Enhanced multi-select behavior with better event handling
                td.addEventListener('mousedown', (e) => {
                    if (e.ctrlKey || e.metaKey) {
                        e.preventDefault();
                        e.stopPropagation();
                        this.isSelectingCells = true;
                        this.toggleCellSelection(td);
                    } else {
                        // Normal click: select the table element and make this cell the single selected cell
                        const parentElement = td.closest('.canvas-element');
                        if (parentElement) {
                            // Select the table element first (this clears element-level selection)
                            this.selectElement(parentElement);
                        }
                        // Clear any previous cell selection and select this one
                        this.clearCellSelection();
                        this.selectedCells.add(td);
                        td.classList.add('selected-cell');
                        // Update properties panel to show the selected cell's borders
                        setTimeout(() => this.showProperties(this.selectedElement), 0);
                    }
                });
                
                td.addEventListener('mouseenter', (e) => {
                    if (this.isSelectingCells && (e.ctrlKey || e.metaKey)) {
                        this.addCellToSelection(td);
                    }
                });
                
                // Prevent clicks inside table cells from deselecting the table
                td.addEventListener('click', (e) => {
                    e.stopPropagation();
                    // If user clicked without holding Ctrl, we've already set this cell as selected in mousedown.
                    // If we're in multi-select mode, just refresh properties to update counts.
                    if (this.isSelectingCells) {
                        setTimeout(() => this.showProperties(this.selectedElement), 0);
                    }
                });
                
                tr.appendChild(td);
            });
            tableElement.appendChild(tr);
        });
    }

    toggleCellSelection(td) {
        if (this.selectedCells.has(td)) {
            this.selectedCells.delete(td);
            td.classList.remove('selected-cell');
        } else {
            this.selectedCells.add(td);
            td.classList.add('selected-cell');
        }
        // Update properties panel to show current selection
        setTimeout(() => this.showProperties(this.selectedElement), 0);
    }

    addCellToSelection(td) {
        if (!this.selectedCells.has(td)) {
            this.selectedCells.add(td);
            td.classList.add('selected-cell');
            // Update properties panel
            setTimeout(() => this.showProperties(this.selectedElement), 0);
        }
    }

    clearCellSelection() {
        if (!this.selectedCells) return;
        this.selectedCells.forEach(td => td.classList.remove('selected-cell'));
        this.selectedCells.clear();
        this.isSelectingCells = false;
        // Update properties panel if table is selected
        if (this.selectedElement?.dataset.type === 'table') {
            setTimeout(() => this.showProperties(this.selectedElement), 0);
        }
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

    // --- ADDED: Missing clearCanvas and clearPlaceholder methods ---
    clearCanvas() {
        this.canvas.innerHTML = '<div class="canvas-placeholder">Drag elements here to start building</div>';
        this.deselectAllElements();
        this.elementCounter = 0; 
        this.generateJSON();
    }

    clearPlaceholder() {
        const placeholder = this.canvas.querySelector('.canvas-placeholder');
        if (placeholder) {
            placeholder.remove();
        }
    }

    // --- ADDED: Missing updateTableTools method ---
    updateTableTools() {
        const isTableSelected = this.selectedElement && this.selectedElement.dataset.type === 'table';
        document.getElementById('addRowBtn').disabled = !isTableSelected;
        document.getElementById('addColumnBtn').disabled = !isTableSelected;
        document.getElementById('removeRowBtn').disabled = !isTableSelected;
        document.getElementById('removeColumnBtn').disabled = !isTableSelected;
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
            
            // Add cell selection handlers
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
                if (this.isSelectingCells && (e.ctrlKey || e.metaKey)) {
                    this.addCellToSelection(cell);
                }
            });
            
            cell.addEventListener('click', (e) => {
                e.stopPropagation();
                if (!this.isSelectingCells) {
                    setTimeout(() => this.showProperties(this.selectedElement), 0);
                }
            });
            
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
            
            // Add cell selection handlers
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
                if (this.isSelectingCells && (e.ctrlKey || e.metaKey)) {
                    this.addCellToSelection(cell);
                }
            });
            
            cell.addEventListener('click', (e) => {
                e.stopPropagation();
                if (!this.isSelectingCells) {
                    setTimeout(() => this.showProperties(this.selectedElement), 0);
                }
            });
            
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