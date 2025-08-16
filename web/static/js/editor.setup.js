// Setup: event listeners, theme controls, and drag/drop
TemplateEditor.prototype.setupEventListeners = function() {
    // Template controls
    document.getElementById('loadTemplateBtn')?.addEventListener('click', () => this.loadTemplate());
    document.getElementById('pasteTemplateBtn')?.addEventListener('click', () => this.showPasteModal());
    document.getElementById('generateTemplateBtn')?.addEventListener('click', () => this.generateJSON());
    document.getElementById('previewPdfBtn')?.addEventListener('click', () => this.previewPDF());
    document.getElementById('copyTemplateBtn')?.addEventListener('click', () => this.copyJSON());
    document.getElementById('formatJsonBtn')?.addEventListener('click', () => this.formatJSON());

    document.getElementById('watermark')?.addEventListener('input', () => this.updateDocumentSettings());
    // Document settings: ensure changes to page size/orientation/border update the model
    document.getElementById('pageSize')?.addEventListener('change', () => this.updateDocumentSettings());
    document.getElementById('pageOrientation')?.addEventListener('change', () => this.updateDocumentSettings());
    document.getElementById('pageBorder')?.addEventListener('input', () => this.updateDocumentSettings());

    // Table tools
    document.getElementById('addRowBtn')?.addEventListener('click', () => this.addTableRow());
    document.getElementById('addColumnBtn')?.addEventListener('click', () => this.addTableColumn());
    document.getElementById('removeRowBtn')?.addEventListener('click', () => this.removeTableRow());
    document.getElementById('removeColumnBtn')?.addEventListener('click', () => this.removeTableColumn());

    // Theme controls
    document.getElementById('themeToggle')?.addEventListener('click', () => this.toggleTheme());
    document.getElementById('gradientSelect')?.addEventListener('change', () => this.applyGradient());

    // Modal controls
    document.getElementById('confirmPasteBtn')?.addEventListener('click', () => this.loadFromPaste());
    document.getElementById('cancelPasteBtn')?.addEventListener('click', () => this.hidePasteModal());
    document.querySelector('.modal .close')?.addEventListener('click', () => this.hidePasteModal());

    // Enter key on template input
    document.getElementById('templateInput')?.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') this.loadTemplate();
    });

    // Keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
            e.preventDefault();
            // generate JSON on Ctrl+S like original
            this.generateJSON();
        }
        if (e.key === 'Delete' && this.selectedElement) {
            this.deleteElement(this.selectedElement);
        }
        if (e.key === 'Escape') {
            this.hidePasteModal();
            this.clearCellSelection?.();
        }
    });

    // Stop cell selection on mouseup
    document.addEventListener('mouseup', () => { this.isSelectingCells = false; });

    // Click outside modal to close
    window.addEventListener('click', (e) => {
        const modal = document.getElementById('pasteModal');
        if (e.target === modal) this.hidePasteModal();
    });

    // Deselect element when clicking on canvas background
    this.canvas?.addEventListener('click', (e) => {
        if (e.target === this.canvas) this.deselectAllElements();
    });
};

TemplateEditor.prototype.setupThemeControls = function() {
    const savedTheme = localStorage.getItem('gopdf_editor_theme') || 'light';
    const savedGradient = localStorage.getItem('gopdf_editor_gradient');
    const isDark = savedTheme === 'dark';

    document.body.classList.toggle('theme-dark', isDark);
    document.body.classList.toggle('theme-light', !isDark);
    const themeToggle = document.getElementById('themeToggle');
    if (themeToggle) themeToggle.textContent = isDark ? '🌞' : '🌙';

    this.populateGradients(isDark);

    const gradientSelect = document.getElementById('gradientSelect');
    if (savedGradient && gradientSelect) {
        gradientSelect.value = savedGradient;
    }

    this.applyGradient();
};

TemplateEditor.prototype.setupDragAndDrop = function() {
    const draggableItems = document.querySelectorAll('.draggable-item');
    draggableItems.forEach(item => {
        item.setAttribute('draggable', 'true');
        item.addEventListener('dragstart', (e) => {
            e.dataTransfer.setData('text/plain', item.dataset.type || '');
            if (item.dataset.type === 'cell-component' || item.dataset.type === 'checkbox') {
                e.dataTransfer.setData('text/cell-component', '1');
            }
            this.canvas.classList.add('drag-over');
        });
        item.addEventListener('dragend', () => this.canvas.classList.remove('drag-over'));

        // Click to add quickly (except for checkbox which is contextual)
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const type = item.dataset.type;
            if (type === 'checkbox') {
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
                    this.updateStatus('Drag the checkbox into a table cell to add it');
                }
            } else {
                this.createElement(type);
            }
        });
    });

    this.canvas?.addEventListener('dragover', (e) => { e.preventDefault(); this.canvas.classList.add('drag-over'); });
    this.canvas?.addEventListener('dragleave', () => this.canvas.classList.remove('drag-over'));
    this.canvas?.addEventListener('drop', (e) => {
        e.preventDefault();
        this.canvas.classList.remove('drag-over');
        const type = e.dataTransfer.getData('text/plain');
        if (type) this.createElement(type);
    });

    // Setup cell-specific drag and drop for checkboxes
    this.setupCellDragAndDrop();
};

TemplateEditor.prototype.setupCellDragAndDrop = function() {
    this.canvas?.addEventListener('dragover', (e) => {
        const cell = e.target.closest('td');
        if (cell && e.dataTransfer.types.includes('text/cell-component')) {
            e.preventDefault();
            cell.classList.add('cell-drag-over');
        }
    });

    this.canvas?.addEventListener('dragleave', (e) => {
        const cell = e.target.closest('td');
        if (cell) cell.classList.remove('cell-drag-over');
    });

    this.canvas?.addEventListener('drop', (e) => {
        const cell = e.target.closest('td');
        if (cell && e.dataTransfer.types.includes('text/cell-component')) {
            e.preventDefault();
            cell.classList.remove('cell-drag-over');
            this.addComponentToCell(cell, 'checkbox');
        }
    });
};

TemplateEditor.prototype.addComponentToCell = function(cell, componentType) {
    const input = cell.querySelector('input');
    if (!input) return;

    if (componentType === 'checkbox') {
        input.type = 'checkbox';
        input.checked = false;
        input.className = 'cell-checkbox';
        this.generateJSON();
        this.updateStatus('Checkbox added to cell');
    }
};
