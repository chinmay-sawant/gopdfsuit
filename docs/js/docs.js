document.addEventListener('DOMContentLoaded', function() {
    // Navigation functionality
    const navLinks = document.querySelectorAll('.nav-link');
    const contentSections = document.querySelectorAll('.content-section');
    const sidebar = document.getElementById('sidebar');
    const content = document.getElementById('content');
    const sidebarToggle = document.getElementById('sidebarToggle');

    // Navigation click handlers
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetSection = this.dataset.section;
            
            // Remove active class from all nav links and sections
            navLinks.forEach(nav => nav.classList.remove('active'));
            contentSections.forEach(section => section.classList.remove('active'));
            
            // Add active class to clicked nav link
            this.classList.add('active');
            
            // Show target section
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

    // Load additional content sections dynamically
    const sectionsData = {
        'pdf-viewer': {
            title: '👁️ PDF Viewer',
            content: `
                <h1>👁️ PDF Viewer</h1>
                <p class="lead">Interactive web-based PDF viewer with real-time template editing and preview capabilities.</p>
                
                <h2>Features</h2>
                <ul>
                    <li>🔄 Real-time JSON template editing with syntax highlighting</li>
                    <li>📄 Live PDF preview with page navigation</li>
                    <li>📱 Mobile-responsive design</li>
                    <li>🎨 Multiple theme support (dark/light mode)</li>
                    <li>📋 Copy/paste JSON templates</li>
                    <li>⬇️ One-click PDF download</li>
                </ul>

                <h2>Access</h2>
                <div class="code-block">
                    <pre><code>http://localhost:8080/
http://localhost:8080/?file=temp_multiplepage.json</code></pre>
                </div>

                <h2>Usage</h2>
                <ol>
                    <li>Enter a JSON template filename in the input field</li>
                    <li>Click "Load Template" to fetch and display the JSON</li>
                    <li>Edit the JSON template directly in the syntax-highlighted editor</li>
                    <li>Click "Generate PDF" to create and preview the PDF</li>
                    <li>Use navigation controls to browse multi-page documents</li>
                    <li>Download the PDF using the download button</li>
                </ol>
            `
        },
        'template-editor': {
            title: '🎨 Template Editor',
            content: `
                <h1>🎨 Template Editor</h1>
                <p class="lead">Visual drag-and-drop PDF template designer with live JSON generation and instant preview.</p>
                
                <h2>Features</h2>
                <ul>
                    <li>🎯 Drag-and-drop interface for building templates</li>
                    <li>📊 Component toolbox (titles, tables, footers, spacers)</li>
                    <li>🔧 Properties panel for customizing components</li>
                    <li>💾 Real-time JSON template generation</li>
                    <li>👁️ Live PDF preview</li>
                    <li>📱 Responsive design for all devices</li>
                </ul>

                <h2>Access</h2>
                <div class="code-block">
                    <pre><code>http://localhost:8080/editor
http://localhost:8080/editor?file=temp_multiplepage.json</code></pre>
                </div>

                <h2>Components</h2>
                <div class="features-grid">
                    <div class="feature-card">
                        <i class="fas fa-heading"></i>
                        <h3>Title</h3>
                        <p>Add headers and titles with customizable fonts and alignment</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-table"></i>
                        <h3>Table</h3>
                        <p>Create complex table layouts with dynamic rows and columns</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-align-center"></i>
                        <h3>Footer</h3>
                        <p>Add page footers with automatic page numbering</p>
                    </div>
                    <div class="feature-card">
                        <i class="fas fa-check-square"></i>
                        <h3>Checkbox</h3>
                        <p>Interactive checkboxes for forms and surveys</p>
                    </div>
                </div>
            `
        },
        'pdf-filler': {
            title: '🖊️ PDF Filler',
            content: `
                <h1>🖊️ PDF Filler</h1>
                <p class="lead">Upload PDF documents and XFDF form data to generate filled PDFs with AcroForm support.</p>
                
                <h2>Features</h2>
                <ul>
                    <li>📄 PDF form filling with XFDF data</li>
                    <li>📤 Simple drag-and-drop file upload</li>
                    <li>⚡ Fast in-memory processing</li>
                    <li>⬇️ Instant download of filled PDFs</li>
                    <li>🔒 Secure file handling</li>
                </ul>

                <h2>Access</h2>
                <div class="code-block">
                    <pre><code>http://localhost:8080/filler</code></pre>
                </div>

                <h2>API Usage</h2>
                <div class="code-block">
                    <h3>cURL Example</h3>
                    <pre><code>curl -X POST "http://localhost:8080/api/v1/fill" \\
  -F "pdf=@patient.pdf;type=application/pdf" \\
  -F "xfdf=@patient.xfdf;type=application/xml" \\
  --output filled.pdf</code></pre>
                </div>

                <h2>Supported Formats</h2>
                <ul>
                    <li><strong>PDF:</strong> Any PDF with AcroForm fields</li>
                    <li><strong>XFDF:</strong> XML Forms Data Format files (.xfdf, .xml)</li>
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
    };

    // Load content for sections that don't exist in HTML
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            const sectionName = this.dataset.section;
            const existingSection = document.getElementById(sectionName + '-section');
            
            if (!existingSection && sectionsData[sectionName]) {
                // Create section dynamically
                const newSection = document.createElement('section');
                newSection.id = sectionName + '-section';
                newSection.className = 'content-section';
                newSection.innerHTML = sectionsData[sectionName].content;
                content.appendChild(newSection);
            }
        });
    });

    // Initialize carousel
    updateCarousel();
});
