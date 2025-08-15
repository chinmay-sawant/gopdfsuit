class PDFViewer {
    constructor() {
        this.currentTemplate = null;
        this.pdfDoc = null;
        this.currentPage = 1;
        this.scale = 1.2;
        this.canvas = document.getElementById('pdfCanvas');
        this.ctx = this.canvas.getContext('2d');
        
        this.lightGradients = [
            { name: 'Default', value: 'linear-gradient(135deg,#667eea 0%,#764ba2 100%)' },
            { name: 'Sunrise', value: 'linear-gradient(120deg,#ff9a9e 0%,#fad0c4 100%)' },
            { name: 'Aqua', value: 'linear-gradient(120deg,#a1c4fd 0%,#c2e9fb 100%)' },
            { name: 'Mint', value: 'linear-gradient(120deg,#43cea2 0%,#185a9d 100%)' }
        ];
        this.darkGradients = [
            { name: 'Default', value: 'linear-gradient(135deg,#141E30 0%,#243B55 100%)' },
            { name: 'Nightfall', value: 'linear-gradient(135deg,#232526 0%,#414345 100%)' },
            { name: 'Purple Haze', value: 'linear-gradient(135deg,#42275a 0%,#734b6d 100%)' },
            { name: 'Deep Space', value: 'linear-gradient(135deg,#000428 0%,#004e92 100%)' }
        ];
        this.initializeElements();
        this.initThemeFromStorage();
        this.attachEventListeners();
        this.checkURLParams();
    }

    initializeElements() {
        this.fileInput = document.getElementById('fileInput');
        this.loadBtn = document.getElementById('loadBtn');
        this.generateBtn = document.getElementById('generateBtn');
        this.jsonDisplay = document.getElementById('jsonDisplay');
        this.copyBtn = document.getElementById('copyBtn');
        this.prevPageBtn = document.getElementById('prevPage');
        this.nextPageBtn = document.getElementById('nextPage');
        this.pageInfo = document.getElementById('pageInfo');
        this.downloadBtn = document.getElementById('downloadBtn');
        this.statusMessage = document.getElementById('statusMessage');
        this.loadingIndicator = document.getElementById('loadingIndicator');
        this.errorIndicator = document.getElementById('errorIndicator');
        this.themeToggle = document.getElementById('themeToggle');
        this.gradientSelect = document.getElementById('gradientSelect');
        this.applyJsonBtn = document.getElementById('applyJsonBtn');
        this.jsonCode = document.getElementById('jsonCode');
    }

    attachEventListeners() {
        this.loadBtn.addEventListener('click', () => this.loadTemplate());
        this.generateBtn.addEventListener('click', () => this.generatePDF());
        this.copyBtn.addEventListener('click', () => this.copyJSON());
        this.prevPageBtn.addEventListener('click', () => this.previousPage());
        this.nextPageBtn.addEventListener('click', () => this.nextPage());
        this.downloadBtn.addEventListener('click', () => this.downloadPDF());
        this.themeToggle.addEventListener('click', () => this.toggleTheme());
        this.gradientSelect.addEventListener('change', () => this.applyGradient());
        this.applyJsonBtn.addEventListener('click', () => this.applyJSONChanges());
        
        this.fileInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.loadTemplate();
            }
        });

        // Auto-load on file input change
        this.fileInput.addEventListener('input', () => {
            if (this.fileInput.value.trim()) {
                this.loadBtn.disabled = false;
            } else {
                this.loadBtn.disabled = true;
            }
        });

        this.jsonCode.addEventListener('input', () => {
            this.applyJsonBtn.disabled = false;
        });

        // Ctrl/Cmd + S to apply
        this.jsonCode.addEventListener('keydown', (e) => {
            if ((e.metaKey || e.ctrlKey) && e.key === 's') {
                e.preventDefault();
                this.applyJSONChanges();
            }
        });
    }

    initThemeFromStorage() {
        const savedTheme = localStorage.getItem('gopdf_theme') || (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
        const savedGradient = localStorage.getItem('gopdf_gradient');
        document.body.classList.toggle('theme-dark', savedTheme === 'dark');
        document.body.classList.toggle('theme-light', savedTheme !== 'dark');
        this.themeToggle.textContent = savedTheme === 'dark' ? '🌞' : '🌙';
        this.populateGradients(savedTheme === 'dark');
        if (savedGradient) {
            document.body.style.background = savedGradient;
            this.gradientSelect.value = savedGradient;
        } else {
            this.applyGradient(); // set default
        }
    }

    updateStatus(message, isError = false) {
        this.statusMessage.textContent = message;
        this.statusMessage.style.color = isError ? '#dc3545' : '#495057';
    }

    async loadTemplate() {
        const filename = this.fileInput.value.trim();
        
        if (!filename) {
            this.updateStatus('Please enter a filename', true);
            return;
        }

        this.updateStatus('Loading template...');
        this.loadBtn.disabled = true;

        try {
            const response = await fetch(`/api/v1/template-data?file=${encodeURIComponent(filename)}`);
            
            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Failed to load template');
            }

            this.currentTemplate = await response.json();
            this.displayJSON();
            this.generateBtn.disabled = false;
            this.updateStatus(`Template "${filename}" loaded successfully`);
            
            // Update URL without refresh
            const newURL = new URL(window.location);
            newURL.searchParams.set('file', filename);
            window.history.pushState({}, '', newURL);

        } catch (error) {
            this.updateStatus(`Error loading template: ${error.message}`, true);
            this.generateBtn.disabled = true;
            this.currentTemplate = null;
        } finally {
            this.loadBtn.disabled = false;
        }
    }

    displayJSON() {
        if (!this.currentTemplate) return;
        const formattedJSON = JSON.stringify(this.currentTemplate, null, 2);
        this.jsonCode.textContent = formattedJSON;
        this.applyJsonBtn.disabled = true;
        if (window.Prism) {
            Prism.highlightElement(this.jsonCode);
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    async copyJSON() {
        if (!this.currentTemplate) return;

        try {
            const jsonText = JSON.stringify(this.currentTemplate, null, 2);
            await navigator.clipboard.writeText(jsonText);
            
            // Visual feedback
            const originalText = this.copyBtn.textContent;
            this.copyBtn.textContent = '✓';
            setTimeout(() => {
                this.copyBtn.textContent = originalText;
            }, 1000);
            
            this.updateStatus('JSON copied to clipboard');
        } catch (error) {
            this.updateStatus('Failed to copy JSON', true);
        }
    }

    async generatePDF() {
        // apply pending edits automatically
        if (!this.applyJsonBtn.disabled) {
            this.applyJSONChanges();
            if (!this.applyJsonBtn.disabled) return; // still invalid
        }

        if (!this.currentTemplate) return;

        this.updateStatus('Generating PDF...');
        this.generateBtn.disabled = true;
        this.showLoading();

        try {
            const response = await fetch('/api/v1/generate/template-pdf', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(this.currentTemplate)
            });

            if (!response.ok) {
                throw new Error('Failed to generate PDF');
            }

            const pdfBlob = await response.blob();
            const pdfUrl = URL.createObjectURL(pdfBlob);
            
            await this.loadPDFFromBlob(pdfBlob);
            this.updateStatus('PDF generated successfully');
            
            // Store PDF blob for download
            this.currentPDFBlob = pdfBlob;
            this.downloadBtn.disabled = false;

        } catch (error) {
            this.updateStatus(`Error generating PDF: ${error.message}`, true);
            this.showError();
        } finally {
            this.generateBtn.disabled = false;
        }
    }

    async loadPDFFromBlob(blob) {
        try {
            const arrayBuffer = await blob.arrayBuffer();
            const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
            
            this.pdfDoc = pdf;
            this.currentPage = 1;
            this.updatePageControls();
            await this.renderPage();
            this.hideLoading();
            
        } catch (error) {
            console.error('Error loading PDF:', error);
            this.showError();
        }
    }

    async renderPage() {
        if (!this.pdfDoc) return;

        try {
            const page = await this.pdfDoc.getPage(this.currentPage);
            const viewport = page.getViewport({ scale: this.scale });
            
            this.canvas.width = viewport.width;
            this.canvas.height = viewport.height;
            
            const renderContext = {
                canvasContext: this.ctx,
                viewport: viewport
            };
            
            await page.render(renderContext).promise;
            this.updatePageInfo();
            
        } catch (error) {
            console.error('Error rendering page:', error);
        }
    }

    updatePageControls() {
        if (!this.pdfDoc) return;

        this.prevPageBtn.disabled = this.currentPage <= 1;
        this.nextPageBtn.disabled = this.currentPage >= this.pdfDoc.numPages;
    }

    updatePageInfo() {
        if (!this.pdfDoc) return;
        
        this.pageInfo.textContent = `Page ${this.currentPage} of ${this.pdfDoc.numPages}`;
    }

    async previousPage() {
        if (this.currentPage <= 1) return;
        
        this.currentPage--;
        this.updatePageControls();
        await this.renderPage();
    }

    async nextPage() {
        if (!this.pdfDoc || this.currentPage >= this.pdfDoc.numPages) return;
        
        this.currentPage++;
        this.updatePageControls();
        await this.renderPage();
    }

    downloadPDF() {
        if (!this.currentPDFBlob) return;

        const url = URL.createObjectURL(this.currentPDFBlob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `template-pdf-${Date.now()}.pdf`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        
        this.updateStatus('PDF downloaded');
    }

    showLoading() {
        this.loadingIndicator.style.display = 'block';
        this.errorIndicator.style.display = 'none';
        this.canvas.style.display = 'none';
    }

    hideLoading() {
        this.loadingIndicator.style.display = 'none';
        this.canvas.style.display = 'block';
    }

    showError() {
        this.loadingIndicator.style.display = 'none';
        this.errorIndicator.style.display = 'block';
        this.canvas.style.display = 'none';
    }

    toggleTheme() {
        const darkNow = !document.body.classList.contains('theme-dark');
        document.body.classList.toggle('theme-dark', darkNow);
        document.body.classList.toggle('theme-light', !darkNow);
        this.themeToggle.textContent = darkNow ? '🌞' : '🌙';
        this.populateGradients(darkNow);
        this.applyGradient();
        localStorage.setItem('gopdf_theme', darkNow ? 'dark' : 'light');
    }

    populateGradients(isDark) {
        const list = isDark ? this.darkGradients : this.lightGradients;
        const prev = this.gradientSelect.value;
        this.gradientSelect.innerHTML = '';
        list.forEach(g => {
            const opt = document.createElement('option');
            opt.value = g.value;
            opt.textContent = g.name;
            this.gradientSelect.appendChild(opt);
        });
        // restore if compatible
        if ([...this.gradientSelect.options].some(o => o.value === prev)) {
            this.gradientSelect.value = prev;
        }
    }

    applyGradient() {
        const val = this.gradientSelect.value || (document.body.classList.contains('theme-dark') ? this.darkGradients[0].value : this.lightGradients[0].value);
        document.body.style.background = val;
        localStorage.setItem('gopdf_gradient', val);
    }

    applyJSONChanges() {
        const raw = this.jsonCode.textContent;
        try {
            const parsed = JSON.parse(raw);
            this.currentTemplate = parsed;
            this.applyJsonBtn.disabled = true;
            this.updateStatus('JSON applied');
            // re-highlight
            if (window.Prism) Prism.highlightElement(this.jsonCode);
        } catch (e) {
            this.updateStatus('Invalid JSON: ' + e.message, true);
        }
    }
}

// Configure PDF.js worker
pdfjsLib.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js';

// Initialize the PDF viewer when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new PDFViewer();
});
