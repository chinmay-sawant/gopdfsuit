// Core: defines the TemplateEditor class and basic properties
class TemplateEditor {
    constructor() {
        // DOM refs will be initialized in init()
        this.canvas = null;
        this.propertiesPanel = null;
        this.jsonOutput = null;
        this.statusMessage = null;

        this.elementCounter = 0;
        this.selectedElement = null;
        this.selectedCells = new Set();
        this.isSelectingCells = false;

        // Theme gradients (restored from original editor.js)
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

        // Template model
        this.template = { config: { page: 'A4', pageAlignment: 1, pageBorder: '1:1:1:1', watermark: '' } };
    }

    init() {
        this.canvas = document.getElementById('canvas');
    // HTML uses id="propertiesContent" for the properties panel
    this.propertiesPanel = document.getElementById('propertiesContent');
        this.jsonOutput = document.getElementById('jsonOutput');
        this.statusMessage = document.getElementById('statusMessage');

        this.checkURLParams?.();
        this.setupThemeControls?.();
        this.setupEventListeners?.();
        this.setupDragAndDrop?.();
        this.generateJSON?.();
    }

    checkURLParams() {
        // optional placeholder - implementations may override or extend
        return;
    }
}

window.TemplateEditor = TemplateEditor;
