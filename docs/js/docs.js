document.addEventListener('DOMContentLoaded', function() {
    // Theme management
    const themeToggle = document.getElementById('themeToggle');
    const html = document.documentElement;
    
    // Initialize theme - default to dark
    const savedTheme = localStorage.getItem('gopdf_docs_theme') || 'dark';
    html.setAttribute('data-theme', savedTheme);
    updateThemeToggle(savedTheme);
    
    // Theme toggle functionality
    themeToggle.addEventListener('click', function() {
        const currentTheme = html.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        
        html.setAttribute('data-theme', newTheme);
        localStorage.setItem('gopdf_docs_theme', newTheme);
        updateThemeToggle(newTheme);
    });
    
    function updateThemeToggle(theme) {
        const icon = themeToggle.querySelector('i');
        const text = themeToggle.querySelector('span');
        
        if (theme === 'dark') {
            icon.className = 'fas fa-sun';
            text.textContent = 'Light';
        } else {
            icon.className = 'fas fa-moon';
            text.textContent = 'Dark';
        }
    }

    // Navigation functionality
    const navLinks = document.querySelectorAll('.nav-link');
    const contentSections = document.querySelectorAll('.content-section');
    const sidebar = document.getElementById('sidebar');
    const content = document.getElementById('content');
    const sidebarToggle = document.getElementById('sidebarToggle');

    // Navigation click handlers (single handler handles dynamic creation)
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();

            const targetSection = this.dataset.section;

            // If the section content is defined in sectionsData but not present in DOM, create it
            const existingSection = document.getElementById(targetSection + '-section');
            if (!existingSection && sectionsData[targetSection]) {
                const newSection = document.createElement('section');
                newSection.id = targetSection + '-section';
                newSection.className = 'content-section';
                newSection.innerHTML = sectionsData[targetSection].content;
                content.appendChild(newSection);
                // Add copy buttons for any code blocks inside the newly created section
                attachCopyButtons(newSection);
            }

            // Remove active class from all nav links and sections
            navLinks.forEach(nav => nav.classList.remove('active'));
            // Re-query content sections to include dynamically added ones
            const allSections = document.querySelectorAll('.content-section');
            allSections.forEach(section => section.classList.remove('active'));

            // Add active class to clicked nav link
            this.classList.add('active');

            // Show target section (now exists)
            const targetElement = document.getElementById(targetSection + '-section');
            if (targetElement) {
                targetElement.classList.add('active');
            }

            // Close sidebar on mobile
            if (window.innerWidth <= 768) {
                sidebar.classList.remove('show');
                content.classList.add('sidebar-hidden');
            }
        });
    });

    // Sidebar toggle functionality
    sidebarToggle.addEventListener('click', function() {
        if (window.innerWidth <= 768) {
            sidebar.classList.toggle('show');
        } else {
            sidebar.classList.toggle('hidden');
            content.classList.toggle('sidebar-hidden');
        }
    });

    // Carousel functionality
    let currentSlide = 0;
    const slides = document.querySelectorAll('.carousel-slide');
    const indicators = document.querySelectorAll('.indicator');
    const track = document.querySelector('.carousel-track');
    const prevBtn = document.getElementById('prevBtn');
    const nextBtn = document.getElementById('nextBtn');

    function updateCarousel() {
        // Update track position
        if (track) {
            track.style.transform = `translateX(-${currentSlide * 100}%)`;
        }
        
        // Update indicators
        indicators.forEach((indicator, index) => {
            indicator.classList.toggle('active', index === currentSlide);
        });
        
        // Update slide active state
        slides.forEach((slide, index) => {
            slide.classList.toggle('active', index === currentSlide);
        });
    }

    function nextSlide() {
        currentSlide = (currentSlide + 1) % slides.length;
        updateCarousel();
    }

    function prevSlide() {
        currentSlide = (currentSlide - 1 + slides.length) % slides.length;
        updateCarousel();
    }

    // Carousel event listeners
    if (nextBtn) {
        nextBtn.addEventListener('click', nextSlide);
    }
    
    if (prevBtn) {
        prevBtn.addEventListener('click', prevSlide);
    }

    // Indicator click handlers
    indicators.forEach((indicator, index) => {
        indicator.addEventListener('click', function() {
            currentSlide = index;
            updateCarousel();
        });
    });

    // Auto-advance carousel
    let carouselInterval = setInterval(nextSlide, 5000);

    // Pause auto-advance on hover
    const carousel = document.getElementById('carousel');
    if (carousel) {
        carousel.addEventListener('mouseenter', function() {
            clearInterval(carouselInterval);
        });

        carousel.addEventListener('mouseleave', function() {
            carouselInterval = setInterval(nextSlide, 5000);
        });
    }

    // Handle window resize
    window.addEventListener('resize', function() {
        if (window.innerWidth > 768) {
            sidebar.classList.remove('show');
            if (!sidebar.classList.contains('hidden')) {
                content.classList.remove('sidebar-hidden');
            }
        }
    });

    // Smooth scrolling for anchor links
    document.addEventListener('click', function(e) {
        if (e.target.tagName === 'A' && e.target.getAttribute('href').startsWith('#')) {
            e.preventDefault();
            const target = document.querySelector(e.target.getAttribute('href'));
            if (target) {
                target.scrollIntoView({ behavior: 'smooth' });
            }
        }
    });

    // Load additional content sections dynamically with enhanced details
    const sectionsData = {
        'pdf-viewer': {
            title: '👁️ PDF Viewer',
            content: `
                <h1>👁️ PDF Viewer</h1>
                <p class="lead">Interactive web-based PDF viewer with real-time template editing, syntax highlighting, and multi-theme support for seamless document creation.</p>
                
                <h2>✨ Key Features</h2>
                <div class="features-grid">
                    <div class="feature-card">
                        <i class="fas fa-edit"></i>
                        <h3>Real-time Editing</h3>
                        <p>Edit JSON templates with syntax highlighting and instant validation</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-eye"></i>
                        <h3>Live Preview</h3>
                        <p>Generate and preview PDFs instantly with page navigation controls</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-palette"></i>
                        <h3>Theme Support</h3>
                        <p>Multiple gradient themes with dark/light mode toggle</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-mobile-alt"></i>
                        <h3>Responsive Design</h3>
                        <p>Optimized for desktop, tablet, and mobile devices</p>
                    </div>
                </div>

                <h2>🔗 Access URLs</h2>
                <div class="code-block">
                    <pre><code># Direct access
http://localhost:8080/

# Load with template
http://localhost:8080/?file=temp_multiplepage.json

# With theme preferences
http://localhost:8080/?file=patient_form.json&theme=dark</code></pre>
                </div>

                <h2>🎯 Usage Workflow</h2>
                <ol>
                    <li><strong>Load Template:</strong> Enter a JSON template filename or use URL parameters</li>
                    <li><strong>Edit JSON:</strong> Modify the template using the syntax-highlighted editor</li>
                    <li><strong>Apply Changes:</strong> Use Ctrl+S or click "Apply JSON" to validate changes</li>
                    <li><strong>Generate PDF:</strong> Click "Generate PDF" to create and preview the document</li>
                    <li><strong>Navigate Pages:</strong> Use controls to browse multi-page documents</li>
                    <li><strong>Download:</strong> Save the generated PDF with one click</li>
                </ol>

                <h2>🎨 Theme Customization</h2>
                <ul>
                    <li><strong>Dark Themes:</strong> Deep Space, Nightfall, Purple Haze, Charcoal</li>
                    <li><strong>Light Themes:</strong> Default, Sunrise, Aqua, Mint</li>
                    <li><strong>Auto-save:</strong> Theme preferences saved to localStorage</li>
                    <li><strong>System Sync:</strong> Respects system dark/light mode preferences</li>
                </ul>
            `
        },
        'template-editor': {
            title: '🎨 Template Editor',
            content: `
                <h1>🎨 Template Editor</h1>
                <p class="lead">Visual drag-and-drop PDF template designer with live JSON generation, properties panel, and instant preview capabilities.</p>
                
                <h2>🎯 Core Features</h2>
                <div class="features-grid">
                    <div class="feature-card">
                        <i class="fas fa-mouse-pointer"></i>
                        <h3>Drag & Drop</h3>
                        <p>Intuitive visual interface for building templates without coding</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-toolbox"></i>
                        <h3>Component Library</h3>
                        <p>Rich set of components: titles, tables, footers, spacers, checkboxes</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-sliders-h"></i>
                        <h3>Properties Panel</h3>
                        <p>Detailed customization options for each component</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-code"></i>
                        <h3>Live JSON</h3>
                        <p>Real-time JSON generation as you build your template</p>
                    </div>
                </div>

                <h2>🔗 Access & URLs</h2>
                <div class="code-block">
                    <pre><code># Template editor
http://localhost:8080/editor

# Load existing template
http://localhost:8080/editor?file=temp_multiplepage.json

# Start with specific configuration
http://localhost:8080/editor?theme=dark&page=A4</code></pre>
                </div>

                <h2>🧩 Available Components</h2>
                <div class="features-grid">
                    <div class="feature-card">
                        <i class="fas fa-heading"></i>
                        <h3>Title Component</h3>
                        <p><strong>Features:</strong> Custom fonts, sizes, bold/italic/underline, alignment<br>
                        <strong>Props:</strong> fontSize, alignment, style flags</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-table"></i>
                        <h3>Table Component</h3>
                        <p><strong>Features:</strong> Dynamic rows/columns, cell borders, multi-select<br>
                        <strong>Controls:</strong> Add/remove rows/columns, border styling</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-align-center"></i>
                        <h3>Footer Component</h3>
                        <p><strong>Features:</strong> Page footers, auto page numbering<br>
                        <strong>Props:</strong> fontSize, alignment, positioning</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-check-square"></i>
                        <h3>Checkbox Component</h3>
                        <p><strong>Features:</strong> Interactive checkboxes for forms<br>
                        <strong>Usage:</strong> Drag into table cells for form creation</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-arrows-alt-v"></i>
                        <h3>Spacer Component</h3>
                        <p><strong>Features:</strong> Layout spacing control<br>
                        <strong>Usage:</strong> Add vertical space between components</p>
                    </div>
                </div>

                <h2>⚙️ Properties Panel Features</h2>
                <ul>
                    <li><strong>Text Styling:</strong> Font size, alignment, bold/italic/underline toggles</li>
                    <li><strong>Table Management:</strong> Cell selection, border controls, font properties</li>
                    <li><strong>Border Designer:</strong> Visual border controls with all/clear options</li>
                    <li><strong>Multi-Selection:</strong> Ctrl+click to select multiple table cells</li>
                    <li><strong>Live Preview:</strong> Changes reflected immediately in the canvas</li>
                </ul>

                <h2>🔄 Editor Workflow</h2>
                <ol>
                    <li><strong>Start Building:</strong> Drag components from toolbox to canvas</li>
                    <li><strong>Configure Properties:</strong> Select elements to edit in properties panel</li>
                    <li><strong>Table Editing:</strong> Use Ctrl+click for multi-cell selection and borders</li>
                    <li><strong>Live JSON:</strong> Monitor real-time JSON generation in the output panel</li>
                    <li><strong>Preview PDF:</strong> Generate instant PDF previews</li>
                    <li><strong>Save/Load:</strong> Export JSON or load existing templates</li>
                </ol>
            `
        },
        'pdf-filler': {
            title: '🖊️ PDF Filler',
            content: `
                <h1>🖊️ PDF Filler</h1>
                <p class="lead">Advanced PDF form filling service with AcroForm support, XFDF processing, and secure file handling for automated document completion.</p>
                
                <h2>✨ Key Capabilities</h2>
                <div class="features-grid">
                    <div class="feature-card">
                        <i class="fas fa-file-pdf"></i>
                        <h3>AcroForm Support</h3>
                        <p>Full compatibility with PDF AcroForm fields and form structures</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-file-code"></i>
                        <h3>XFDF Processing</h3>
                        <p>XML Forms Data Format parsing and field value injection</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-upload"></i>
                        <h3>File Upload</h3>
                        <p>Secure multipart file upload with validation and size limits</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-download"></i>
                        <h3>Instant Download</h3>
                        <p>Generated filled PDFs available for immediate download</p>
                    </div>
                </div>

                <h2>🔗 Access Points</h2>
                <div class="code-block">
                    <pre><code># Web interface
http://localhost:8080/filler

# API endpoint
POST http://localhost:8080/api/v1/fill</code></pre>
                </div>

                <h2>📡 API Usage</h2>
                <div class="code-block">
                    <h3>cURL Example - Multipart Upload</h3>
                    <pre><code>curl -X POST "http://localhost:8080/api/v1/fill" \\
  -F "pdf=@patient_form.pdf;type=application/pdf" \\
  -F "xfdf=@patient_data.xfdf;type=application/xml" \\
  --output filled_patient_form.pdf</code></pre>
                </div>

                <div class="code-block">
                    <h3>Python Example</h3>
                    <pre><code>import requests

url = "http://localhost:8080/api/v1/fill"
files = {
    'pdf': ('form.pdf', open('patient_form.pdf', 'rb'), 'application/pdf'),
    'xfdf': ('data.xfdf', open('patient_data.xfdf', 'rb'), 'application/xml')
}

response = requests.post(url, files=files)
with open('filled_form.pdf', 'wb') as f:
    f.write(response.content)</code></pre>
                </div>

                <h2>🔧 Technical Implementation</h2>
                <ul>
                    <li><strong>Field Detection:</strong> Heuristic parsing to locate AcroForm field names</li>
                    <li><strong>Value Injection:</strong> Direct byte-oriented field value insertion</li>
                    <li><strong>Appearance Handling:</strong> Sets /NeedAppearances for viewer regeneration</li>
                    <li><strong>Memory Processing:</strong> In-memory operations for security and speed</li>
                    <li><strong>Format Support:</strong> .pdf (source), .xfdf/.xml (data)</li>
                </ul>

                <h2>📋 Supported Form Types</h2>
                <div class="features-grid">
                    <div class="feature-card">
                        <i class="fas fa-edit"></i>
                        <h3>Text Fields</h3>
                        <p>Single-line and multi-line text input fields</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-check-square"></i>
                        <h3>Checkboxes</h3>
                        <p>Boolean checkbox fields with true/false values</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-dot-circle"></i>
                        <h3>Radio Buttons</h3>
                        <p>Single-selection radio button groups</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-list"></i>
                        <h3>Dropdown Lists</h3>
                        <p>Selection lists and combo boxes</p>
                    </div>
                </div>

                <h2>⚠️ Limitations & Notes</h2>
                <ul>
                    <li><strong>Compressed Streams:</strong> PDFs with compressed object streams may have limited support</li>
                    <li><strong>Appearance Streams:</strong> Complex appearance generation relies on viewer capabilities</li>
                    <li><strong>Indirect References:</strong> Best effort for complex reference structures</li>
                    <li><strong>Viewer Compatibility:</strong> Results may vary across different PDF viewers</li>
                </ul>

                <h2>🔒 Security Features</h2>
                <ul>
                    <li><strong>File Validation:</strong> MIME type checking and file extension validation</li>
                    <li><strong>Size Limits:</strong> Configurable upload size restrictions</li>
                    <li><strong>Memory Safety:</strong> No temporary file creation, all in-memory processing</li>
                    <li><strong>Input Sanitization:</strong> XFDF parsing with malformed data protection</li>
                </ul>
            `
        },
        'api-endpoints': {
            title: '🔌 API Endpoints',
            content: `
                <h1>🔌 API Endpoints</h1>
                <p class="lead">Complete API reference for all GoPdfSuit endpoints and their usage.</p>

                <h2>Template Data API</h2>
                <div class="code-block">
                    <h3>GET /api/v1/template-data</h3>
                    <pre><code>curl "http://localhost:8080/api/v1/template-data?file=temp_multiplepage.json"</code></pre>
                </div>
                
                <h2>PDF Generation API</h2>
                <div class="code-block">
                    <h3>POST /api/v1/generate/template-pdf</h3>
                    <pre><code>curl -X POST "http://localhost:8080/api/v1/generate/template-pdf" \\
  -H "Content-Type: application/json" \\
  -d '{
    "config": {
      "pageBorder": "1:1:1:1",
      "page": "A4",
      "pageAlignment": 1
    },
    "title": {
      "props": "font1:24:100:center:0:0:1:0",
      "text": "Sample Document"
    }
  }' \\
  --output document.pdf</code></pre>
                </div>

                <h2>PDF Filling API</h2>
                <div class="code-block">
                    <h3>POST /api/v1/fill</h3>
                    <pre><code>curl -X POST "http://localhost:8080/api/v1/fill" \\
  -F "pdf=@form.pdf" \\
  -F "xfdf=@data.xfdf" \\
  --output filled.pdf</code></pre>
                </div>
            `
        }
            ,
            'template-structure': {
                title: '🧩 Template Structure',
                content: `
                    <h1>🧩 Template Structure</h1>
                    <p class="lead">JSON-driven template format used by GoPdfSuit. Templates define the document layout, components and properties.</p>

                    <h2>Top-level Fields</h2>
                    <ul>
                        <li><strong>config</strong>: Page configuration (size, borders, orientation, watermark)</li>
                        <li><strong>title</strong>: Optional title component (props + text)</li>
                        <li><strong>table</strong>: Array of table blocks that contain rows and cells</li>
                        <li><strong>footer</strong>: Footer configuration (font, text)</li>
                    </ul>

                    <h2>Component Shapes (contract)</h2>
                    <p>Inputs/outputs and validation rules:</p>
                    <ul>
                        <li><strong>Cell</strong>: { props: string, text?: string, chequebox?: boolean }</li>
                        <li><strong>Row</strong>: { row: Cell[] }</li>
                        <li><strong>Table</strong>: { maxcolumns: number, rows: Row[] }</li>
                        <li><strong>Props format</strong>: "fontName:fontSize:style:alignment:left:right:top:bottom"</li>
                    </ul>

                    <h2>Style Flags</h2>
                    <p>Style is a 3-digit code: <code>bold(1/0)italic(1/0)underline(1/0)</code> e.g. <code>110</code> = bold+italic.</p>

                    <h2>Example Minimal Template</h2>
                    <div class="code-block">
                        <pre><code>{
      "config": { "pageBorder": "1:1:1:1", "page": "A4", "pageAlignment": 1 },
      "title": { "props": "font1:20:100:center:0:0:2:0", "text": "Sample Title" },
      "table": [ { "maxcolumns": 2, "rows": [ { "row": [ { "props": "font1:12:100:left:1:1:1:1", "text": "Label:" }, { "props": "font1:12:000:left:0:1:1:1", "text": "Value" } ] } ] } ],
      "footer": { "font": "font1:10:001:center", "text": "Page footer" }
    }</code></pre>
                    </div>
                `
            },
            'examples': {
                title: '📄 Examples',
                content: `
                    <h1>📄 Examples</h1>
                    <p class="lead">Practical sample templates and usage patterns. Sample data is available in the <code>sampledata/json</code> folder.</p>

                    <h2>Multi-page Example</h2>
                    <div class="code-block">
                        <pre><code>// See sampledata/json/temp_mutliplepage.json for a full example
    {
      "config": { "page": "A4", "pageBorder": "1:1:1:1", "pageAlignment": 1 },
      "title": { "props": "font1:22:100:center:0:0:2:0", "text": "Multi-Page Demo" },
      "table": [ { "maxcolumns": 4, "rows": [ /* many rows to demonstrate auto page breaks */ ] } ],
      "footer": { "font": "font1:10:000:center", "text": "Confidential" }
    }
    </code></pre>
                    </div>

                    <h2>Filler Example (XFDF)</h2>
                    <div class="code-block">
                        <pre><code>// Use the web UI or POST to /api/v1/fill with multipart form-data (pdf + xfdf)
    // Example XFDF content is available in the sampledata folder
    </code></pre>
                    </div>
                `
            },
            'configuration': {
                title: '⚙️ Configuration',
                content: `
                    <h1>⚙️ Configuration</h1>
                    <p class="lead">Configuration options for server and templates.</p>

                    <h2>Template Config</h2>
                    <ul>
                        <li><strong>page</strong>: Page size (A4, LETTER, LEGAL, A3, A5)</li>
                        <li><strong>pageAlignment</strong>: 1=portrait, 2=landscape</li>
                        <li><strong>pageBorder</strong>: "left:right:top:bottom" border widths in points</li>
                        <li><strong>watermark</strong>: Optional text rendered diagonally across pages</li>
                    </ul>

                    <h2>Server Config (dev)</h2>
                    <ul>
                        <li><strong>Upload size limits</strong>: Configure max upload size for /api/v1/fill</li>
                        <li><strong>Timeouts</strong>: Request and generation timeouts to prevent runaway jobs</li>
                        <li><strong>Logging</strong>: Verbose logging can be enabled in development</li>
                    </ul>

                    <h2>Performance Note</h2>
                    <p>GoPdfSuit generates PDFs entirely in-memory. Typical generation from JSON (tested locally) completes in milliseconds; filling a PDF via XFDF takes up to ~12ms depending on the PDF complexity and environment. Test data is in <code>sampledata/json</code>.</p>
                `
            },
            'development': {
                title: '🛠️ Development',
                content: `
                    <h1>🛠️ Development</h1>
                    <p class="lead">Developer guidance for running and extending GoPdfSuit.</p>

                    <h2>Build & Run</h2>
                    <div class="code-block">
                        <pre><code>go mod download
    # Run server
    go run ./cmd/gopdfsuit</code></pre>
                    </div>

                    <h2>Tests</h2>
                    <p>Run unit tests with:</p>
                    <div class="code-block"><pre><code>go test ./... -v</code></pre></div>

                    <h2>Project Layout</h2>
                    <ul>
                        <li><code>cmd/</code>: application entrypoints</li>
                        <li><code>internal/</code>: handlers, models, pdf generation</li>
                        <li><code>web/</code>: UI templates and static assets</li>
                        <li><code>sampledata/</code>: example templates and XFDF samples</li>
                    </ul>
                `
            },
            'contributing': {
                title: '🤝 Contributing',
                content: `
                    <h1>🤝 Contributing</h1>
                    <p class="lead">How to contribute to GoPdfSuit.</p>

                    <h2>Process</h2>
                    <ol>
                        <li>Fork the repo</li>
                        <li>Create a feature branch: <code>git checkout -b feature/my-change</code></li>
                        <li>Write tests for new behaviour</li>
                        <li>Open a PR and include a description and screenshots</li>
                    </ol>

                    <h2>Code Style</h2>
                    <ul>
                        <li>Follow Go idioms and formatting (<code>gofmt</code>)</li>
                        <li>Keep public APIs stable; add feature flags for breaking changes</li>
                    </ul>

                    <h2>Communication</h2>
                    <p>Open issues for bugs or proposals. Label your PRs with the relevant tags.</p>
                `
            },
            'comparison': {
                title: '🔍 Comparison',
                content: `
                    <h1>🔍 GoPdfSuit vs Licensed Tools</h1>
                    <p class="lead">A practical comparison between GoPdfSuit and commercial licensed PDF SDKs (e.g., Aspose, Unidoc commercial).</p>

                    <h2>At-a-glance Comparison</h2>
                    <div class="code-block">
                        <h3>Feature Comparison</h3>
                        <table style="width:100%; border-collapse: collapse;">
                            <thead>
                                <tr>
                                    <th style="border:1px solid var(--border-color); padding:8px; text-align:left;">Feature / Tool</th>
                                    <th style="border:1px solid var(--border-color); padding:8px; text-align:left;">GoPdfSuit (OSS)</th>
                                    <th style="border:1px solid var(--border-color); padding:8px; text-align:left;">Aspose / Commercial</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">In-memory generation</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">✅ Very fast (ms)</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">✅ Available</td>
                                </tr>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">JSON template driven</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">✅ Yes</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">❌ Usually API-based or proprietary templates</td>
                                </tr>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">AcroForm filling (XFDF)</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">✅ Yes (heuristic)</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">✅ Robust, appearance-ready</td>
                                </tr>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Cost</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">✅ Free (MIT)</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">❌ Paid license</td>
                                </tr>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Appearance streams</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Limited</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Full support</td>
                                </tr>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Speed (generate)</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Milliseconds (local tests)</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Variable</td>
                                </tr>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Deployment</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Single binary</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Library or service</td>
                                </tr>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Extensibility</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">✅ Open-source</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">✅ (with license)</td>
                                </tr>
                                <tr>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Support</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Community</td>
                                    <td style="border:1px solid var(--border-color); padding:8px;">Commercial SLA</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <h2>Benefits of GoPdfSuit</h2>
                    <ul>
                        <li>Generates PDFs entirely in-memory for speed and security</li>
                        <li>Millisecond-scale JSON-based generation (tested locally) which is suitable for high-throughput systems</li>
                        <li>Lightweight single-binary deployment with no external dependencies</li>
                        <li>Transparent open-source codebase for customization</li>
                    </ul>

                    <h2>When to choose a commercial SDK</h2>
                    <ul>
                        <li>Require guaranteed appearance stream generation across all viewers</li>
                        <li>Need enterprise support, SLAs and advanced features (PDF/A, advanced annotations, OCR)</li>
                        <li>Large-scale document transformations requiring optimized native routines</li>
                    </ul>
                `
            }
    };

    // Attach copy buttons to code blocks (for static and dynamic content)
    function attachCopyButtons(root = document) {
        const blocks = root.querySelectorAll('.code-block');
        blocks.forEach(block => {
            if (block.querySelector('.copy-btn')) return; // already added
            const pre = block.querySelector('pre');
            if (!pre) return;

            const btn = document.createElement('button');
            btn.className = 'copy-btn';
            btn.type = 'button';
            btn.innerHTML = '<i class="fas fa-copy"></i><span>Copy</span>';

            btn.addEventListener('click', async function() {
                const code = pre.innerText;
                try {
                    await navigator.clipboard.writeText(code);
                    btn.classList.add('copied');
                    btn.querySelector('span').textContent = 'Copied';
                    setTimeout(() => {
                        btn.classList.remove('copied');
                        btn.querySelector('span').textContent = 'Copy';
                    }, 1500);
                } catch (err) {
                    console.warn('Copy failed', err);
                }
            });

            block.appendChild(btn);
        });
    }

    // (Dynamic sections are created in the navigation click handler above)

    // Initialize carousel
    updateCarousel();
    // Attach copy buttons for any existing static code blocks
    attachCopyButtons();
});
