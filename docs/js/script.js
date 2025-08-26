document.addEventListener('DOMContentLoaded', function() {
    // Theme Toggle Functionality
    const themeToggle = document.getElementById('themeToggle');
    const themeIcon = document.getElementById('themeIcon');
    const body = document.body;

    // Check for saved theme preference or default to light mode
    const currentTheme = localStorage.getItem('theme') || 'light';
    
    // Apply the theme
    if (currentTheme === 'dark') {
        body.setAttribute('data-theme', 'dark');
        if (themeIcon) themeIcon.textContent = '☀️';
    } else {
        body.removeAttribute('data-theme');
        if (themeIcon) themeIcon.textContent = '🌙';
    }
    // Theme toggle click handler
    themeToggle.addEventListener('click', function() {
        const isDark = body.getAttribute('data-theme') === 'dark';
        
        if (isDark) {
            body.removeAttribute('data-theme');
            if (themeIcon) themeIcon.textContent = '🌙';
            localStorage.setItem('theme', 'light');
        } else {
            body.setAttribute('data-theme', 'dark');
            if (themeIcon) themeIcon.textContent = '☀️';
            localStorage.setItem('theme', 'dark');
        }
        
        // Add a little animation
        themeToggle.style.transform = 'rotate(360deg)';
        setTimeout(() => {
            themeToggle.style.transform = 'rotate(0deg)';
        }, 300);
    });
    });

    // Copy to Clipboard Functionality
    function copyToClipboard(text) {
        navigator.clipboard.writeText(text).then(() => {
            // Show feedback
            showCopyFeedback();
        }).catch(err => {
            console.error('Failed to copy text: ', err);
        });
    }

    // Scroll animations
    function initScrollAnimations() {
        const observerOptions = {
            threshold: 0.1,
            rootMargin: '0px 0px -50px 0px'
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('visible');
                }
            });
        }, observerOptions);

        document.querySelectorAll('.animate-on-scroll').forEach(el => {
            observer.observe(el);
        });
    }

    // Enhanced copy feedback
    function showCopyFeedback() {
        const feedback = document.createElement('div');
        feedback.innerHTML = '✅ Copied to clipboard!';
        feedback.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: linear-gradient(135deg, var(--success-color), #059669);
            color: white;
            padding: 12px 20px;
            border-radius: 12px;
            font-size: 14px;
            font-weight: 600;
            z-index: 1000;
            box-shadow: 0 8px 25px rgba(16, 185, 129, 0.4);
            animation: fadeInOut 3s ease-in-out;
            backdrop-filter: blur(10px);
        `;
        
        document.body.appendChild(feedback);
        
        setTimeout(() => {
            if (document.body.contains(feedback)) {
                document.body.removeChild(feedback);
            }
        }, 3000);
    }

    // Particle background
    function createParticles() {
        const hero = document.querySelector('.hero');
        if (!hero) return;

        const particlesContainer = document.createElement('div');
        particlesContainer.className = 'particles';
        hero.appendChild(particlesContainer);

        for (let i = 0; i < 50; i++) {
            const particle = document.createElement('div');
            particle.className = 'particle';
            particle.style.left = Math.random() * 100 + '%';
            particle.style.animationDelay = Math.random() * 6 + 's';
            particle.style.animationDuration = (Math.random() * 3 + 3) + 's';
            particlesContainer.appendChild(particle);
        }
    }

    // Enhanced API card expansion
    function initAPICards() {
        const apiCards = document.querySelectorAll('.api-card[data-expandable="true"]');
        
        apiCards.forEach((card, index) => {
            card.style.animationDelay = `${index * 0.1}s`;
            card.classList.add('animate-on-scroll');
            
            const expandBtn = card.querySelector('.api-expand');
            const details = card.querySelector('.api-details');
            
            expandBtn.addEventListener('click', () => {
                const isExpanded = card.classList.contains('expanded');
                
                // Close all other cards
                apiCards.forEach(otherCard => {
                    if (otherCard !== card) {
                        otherCard.classList.remove('expanded');
                        const otherBtn = otherCard.querySelector('.api-expand');
                        const otherDetails = otherCard.querySelector('.api-details');
                        otherBtn.textContent = '+';
                        otherDetails.style.maxHeight = '0';
                    }
                });
                
                if (isExpanded) {
                    card.classList.remove('expanded');
                    expandBtn.textContent = '+';
                    details.style.maxHeight = '0';
                } else {
                    card.classList.add('expanded');
                    expandBtn.textContent = '−';
                    details.style.maxHeight = details.scrollHeight + 'px';
                    
                    // Smooth scroll to card
                    setTimeout(() => {
                        card.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    }, 300);
                }
            });
        });
    }

    // Tab Functionality
    function initTabs() {
        const tabBtns = document.querySelectorAll('.tab-btn');
        const tabContents = document.querySelectorAll('.tab-content');
        
        tabBtns.forEach(btn => {
            btn.addEventListener('click', () => {
                const targetTab = btn.getAttribute('data-tab');
                
                // Remove active classes
                tabBtns.forEach(b => b.classList.remove('active'));
                tabContents.forEach(c => c.classList.remove('active'));
                
                // Add active classes
                btn.classList.add('active');
                document.getElementById(targetTab).classList.add('active');
            });
        });
    }

    // Smooth Scrolling for Navigation Links
    function initSmoothScrolling() {
        const navLinks = document.querySelectorAll('.nav-link[href^="#"]');
        
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const targetId = link.getAttribute('href').substring(1);
                const targetElement = document.getElementById(targetId);
                
                if (targetElement) {
                    targetElement.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }

    // Initialize everything when DOM is loaded
    document.addEventListener('DOMContentLoaded', () => {
        initAPICards();
        initTabs();
        initSmoothScrolling();
        
        // Add copy button listeners
        document.querySelectorAll('.copy-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const textToCopy = btn.getAttribute('data-copy');
                copyToClipboard(textToCopy);
            });
        });
        
        // Add new initializations
        initScrollAnimations();
        createParticles();
        
        // Enhanced copy button feedback
        document.querySelectorAll('.copy-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const textToCopy = btn.getAttribute('data-copy');
                copyToClipboard(textToCopy);
                
                // Visual feedback on button
                const originalText = btn.innerHTML;
                btn.innerHTML = '✅';
                btn.style.background = 'var(--success-color)';
                btn.style.color = 'white';
                
                setTimeout(() => {
                    btn.innerHTML = originalText;
                    btn.style.background = '';
                    btn.style.color = '';
                }, 1500);
            });
        });
        
        // Add enhanced animations CSS
        const style = document.createElement('style');
        style.textContent = `
            /* ...existing CSS... */
        
            .api-card {
                opacity: 0;
                transform: translateY(30px);
                animation: fadeInUp 0.6s ease-out forwards;
            }
        
            .tab-btn {
                transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
            }
        
            .tab-btn:hover {
                transform: translateY(-2px);
            }
        
            .carousel-item img {
                transition: transform 0.3s ease;
            }
        
            .carousel-item:hover img {
                transform: scale(1.02);
            }
        `;
        document.head.appendChild(style);
    });


// Export functions for potential use in other scripts
window.GoPdfSuitDocs = {
    toggleTheme: function() {
        document.getElementById('themeToggle').click();
    },
    
    scrollToSection: function(sectionId) {
        const section = document.getElementById(sectionId);
        if (section) {
            section.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    },
    
    copyToClipboard: function(text) {
        if (navigator.clipboard) {
            return navigator.clipboard.writeText(text);
        } else {
            // Fallback for older browsers
            const textarea = document.createElement('textarea');
            textarea.value = text;
            document.body.appendChild(textarea);
            textarea.select();
            const result = document.execCommand('copy');
            document.body.removeChild(textarea);
            return Promise.resolve(result);
        }
    }
};

