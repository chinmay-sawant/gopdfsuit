import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  FileText,
  Edit,
  Merge,
  FileCheck,
  Globe,
  Image,
  Zap,
  Shield,
  Download,
  CheckCircle,
  Star,
  Github,
  ChevronDown,
  ArrowRight,
  Sparkles
} from 'lucide-react'

const Home = () => {
  const [isVisible, setIsVisible] = useState({})
  const [typewriterText, setTypewriterText] = useState('')

  const fullText = "  A powerful Go web service that generates template-based PDF documents on-the-fly with multi-page support, PDF merge capabilities, and HTML to PDF/Image conversion."

  // Typewriter effect
  useEffect(() => {
    let i = 0
    const timer = setInterval(() => {
      if (i < fullText.length) {
        setTypewriterText(prev => prev + fullText.charAt(i))
        i++
      } else {
        clearInterval(timer)
      }
    }, 30)

    return () => clearInterval(timer)
  }, [])

  // Intersection Observer for scroll animations
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setIsVisible(prev => ({ ...prev, [entry.target.id]: true }))
          }
        })
      },
      { threshold: 0.1 }
    )

    // Observe all sections
    const sections = document.querySelectorAll('[id^="section-"]')
    sections.forEach((section) => observer.observe(section))

    return () => observer.disconnect()
  }, [])
  const features = [
    {
      icon: <FileText size={32} />,
      title: 'Template-based PDF Generation',
      description: 'JSON-driven PDF creation with multi-page support, automatic page breaks, and custom layouts.',
      link: '/viewer',
      color: 'teal',
      size: 'large'
    },
    {
      icon: <Edit size={32} />,
      title: 'Visual PDF Editor',
      description: 'Drag-and-drop interface for building PDF templates with live preview and real-time JSON generation.',
      link: '/editor',
      color: 'blue',
      size: 'normal'
    },
    {
      icon: <Merge size={32} />,
      title: 'PDF Merge',
      description: 'Combine multiple PDF files with intuitive drag-and-drop reordering and live preview.',
      link: '/merge',
      color: 'purple',
      size: 'normal'
    },
    {
      icon: <FileCheck size={32} />,
      title: 'Form Filling',
      description: 'AcroForm and XFDF support for filling PDF forms programmatically.',
      link: '/filler',
      color: 'yellow',
      size: 'normal'
    },
    {
      icon: <Globe size={32} />,
      title: 'HTML to PDF',
      description: 'Convert HTML content or web pages to PDF using Chromium with full control over page settings.',
      link: '/htmltopdf',
      color: 'green',
      size: 'normal'
    },
    {
      icon: <Image size={32} />,
      title: 'HTML to Image',
      description: 'Convert HTML content to PNG, JPG, or SVG images with custom dimensions and quality settings.',
      link: '/htmltoimage',
      color: 'blue',
      size: 'wide'
    }
  ]

  const highlights = [
    { icon: <Zap />, title: 'Ultra Fast', desc: 'Average 0.8ms response time for PDF generation' },
    { icon: <Shield />, title: 'Secure', desc: 'Path traversal protection and input validation' },
    { icon: <Download />, title: 'Self-contained', desc: 'Single binary deployment with zero dependencies' },
  ]

  // Interactive dots canvas background (Antigravity-style)
  const BackgroundAnimation = () => {
    const canvasRef = React.useRef(null);

    React.useEffect(() => {
      const canvas = canvasRef.current;
      if (!canvas) return;

      const ctx = canvas.getContext('2d');
      let animationFrameId;
      let mouse = { x: null, y: null, radius: 150 };
      let opacity = 0; // Start with 0 opacity for fade-in
      let frameCount = 0; // Track frames for smooth startup

      // Set canvas size
      const resize = () => {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
      };
      resize();
      window.addEventListener('resize', resize);

      // Track mouse
      const handleMouseMove = (e) => {
        mouse.x = e.clientX;
        mouse.y = e.clientY;
      };
      const handleMouseLeave = () => {
        mouse.x = null;
        mouse.y = null;
      };
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseleave', handleMouseLeave);

      // Particle class
      class Particle {
        constructor() {
          this.x = Math.random() * canvas.width;
          this.y = Math.random() * canvas.height;
          this.baseX = this.x;
          this.baseY = this.y;
          this.size = Math.random() * 3 + 1;
          this.speedX = (Math.random() - 0.5) * 0.3; // Reduced initial speed
          this.speedY = (Math.random() - 0.5) * 0.3; // Reduced initial speed
          this.density = Math.random() * 30 + 1;
          // Color palette: teal, blue, purple variations
          const colors = [
            `rgba(78, 205, 196, ${Math.random() * 0.5 + 0.3})`,
            `rgba(0, 122, 204, ${Math.random() * 0.4 + 0.2})`,
            `rgba(240, 147, 251, ${Math.random() * 0.3 + 0.2})`,
          ];
          this.color = colors[Math.floor(Math.random() * colors.length)];
        }

        update() {
          // Gradually increase movement speed over first 60 frames
          const speedMultiplier = Math.min(1, frameCount / 60);

          // Move particles with wave motion (gentler)
          this.x += (this.speedX + Math.sin(Date.now() * 0.001 + this.baseY * 0.01) * 0.2) * speedMultiplier;
          this.y += (this.speedY + Math.cos(Date.now() * 0.001 + this.baseX * 0.01) * 0.15) * speedMultiplier;

          // Wrap around screen
          if (this.x > canvas.width + 50) this.x = -50;
          if (this.x < -50) this.x = canvas.width + 50;
          if (this.y > canvas.height + 50) this.y = -50;
          if (this.y < -50) this.y = canvas.height + 50;

          // Mouse interaction - repel particles
          if (mouse.x !== null && mouse.y !== null) {
            const dx = mouse.x - this.x;
            const dy = mouse.y - this.y;
            const distance = Math.sqrt(dx * dx + dy * dy);

            if (distance < mouse.radius) {
              const force = (mouse.radius - distance) / mouse.radius;
              const angle = Math.atan2(dy, dx);
              this.x -= Math.cos(angle) * force * 3;
              this.y -= Math.sin(angle) * force * 3;
            }
          }
        }

        draw(globalOpacity) {
          ctx.beginPath();
          ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2);
          ctx.fillStyle = this.color.replace(/[\d.]+\)$/, `${parseFloat(this.color.match(/[\d.]+\)$/)[0]) * globalOpacity})`);
          ctx.fill();
        }
      }

      // Create particles
      const particleCount = Math.min(100, Math.floor((canvas.width * canvas.height) / 15000));
      const particles = [];
      for (let i = 0; i < particleCount; i++) {
        particles.push(new Particle());
      }

      // Draw connections between nearby particles
      const connectParticles = (globalOpacity) => {
        for (let i = 0; i < particles.length; i++) {
          for (let j = i + 1; j < particles.length; j++) {
            const dx = particles[i].x - particles[j].x;
            const dy = particles[i].y - particles[j].y;
            const distance = Math.sqrt(dx * dx + dy * dy);

            if (distance < 120) {
              ctx.beginPath();
              ctx.strokeStyle = `rgba(78, 205, 196, ${0.15 * (1 - distance / 120) * globalOpacity})`;
              ctx.lineWidth = 0.5;
              ctx.moveTo(particles[i].x, particles[i].y);
              ctx.lineTo(particles[j].x, particles[j].y);
              ctx.stroke();
            }
          }
        }
      };

      // Animation loop
      const animate = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        // Smoothly fade in over first 30 frames
        if (opacity < 1) {
          opacity = Math.min(1, opacity + 0.033);
        }
        frameCount++;

        // Draw and update particles with current opacity
        for (const particle of particles) {
          particle.update();
          particle.draw(opacity);
        }

        // Draw connections with current opacity
        connectParticles(opacity);

        animationFrameId = requestAnimationFrame(animate);
      };

      // Start animation after a small delay for smoother initial load
      const startTimeout = setTimeout(() => {
        animate();
      }, 50);

      // Cleanup
      return () => {
        clearTimeout(startTimeout);
        cancelAnimationFrame(animationFrameId);
        window.removeEventListener('resize', resize);
        window.removeEventListener('mousemove', handleMouseMove);
        window.removeEventListener('mouseleave', handleMouseLeave);
      };
    }, []);

    return (
      <>
        <canvas
          ref={canvasRef}
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            pointerEvents: 'none',
            zIndex: -1,
          }}
        />
        <style>
          {`
            @keyframes fadeInUp {
              from {
                opacity: 0;
                transform: translate3d(0, 40px, 0);
              }
              to {
                opacity: 1;
                transform: translate3d(0, 0, 0);
              }
            }
            
            @keyframes fadeInScale {
              from {
                opacity: 0;
                transform: scale(0.8);
              }
              to {
                opacity: 1;
                transform: scale(1);
              }
            }
            
            @keyframes slideInLeft {
              from {
                opacity: 0;
                transform: translate3d(-100px, 0, 0);
              }
              to {
                opacity: 1;
                transform: translate3d(0, 0, 0);
              }
            }
            
            @keyframes slideInRight {
              from {
                opacity: 0;
                transform: translate3d(100px, 0, 0);
              }
              to {
                opacity: 1;
                transform: translate3d(0, 0, 0);
              }
            }
            
            @keyframes blink {
              0%, 50% {
                opacity: 1;
              }
              51%, 100% {
                opacity: 0;
              }
            }
            
            .animate-fadeInUp {
              animation: fadeInUp 0.8s ease-out forwards;
            }
            
            .animate-fadeInScale {
              animation: fadeInScale 0.6s ease-out forwards;
            }
            
            .animate-slideInLeft {
              animation: slideInLeft 0.8s ease-out forwards;
            }
            
            .animate-slideInRight {
              animation: slideInRight 0.8s ease-out forwards;
            }
            
            .stagger-animation {
              opacity: 0;
            }
            
            .stagger-animation.visible {
              opacity: 1;
            }
            
            /* Custom Scrollbar Styles */
            .custom-scrollbar::-webkit-scrollbar {
              width: 8px;
            }
            
            .custom-scrollbar::-webkit-scrollbar-track {
              background: rgba(0, 0, 0, 0.3);
              border-radius: 4px;
            }
            
            .custom-scrollbar::-webkit-scrollbar-thumb {
              background: rgba(78, 205, 196, 0.5);
              border-radius: 4px;
            }
            
            .custom-scrollbar::-webkit-scrollbar-thumb:hover {
              background: rgba(78, 205, 196, 0.8);
            }
          `}
        </style>
      </>
    )
  }

  return (
    <div style={{ minHeight: '100vh', position: 'relative' }}>
      <BackgroundAnimation />

      {/* Hero Section */}
      <section
        id="section-hero"
        className="hero-section"
        style={{
          padding: '6rem 0 4rem',
          textAlign: 'center',
        }}
      >
        <div className="container">
          {/* Sparkle badge */}
          <div
            className="animate-fadeInUp"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.5rem 1rem',
              background: 'rgba(78, 205, 196, 0.1)',
              border: '1px solid rgba(78, 205, 196, 0.3)',
              borderRadius: '50px',
              marginBottom: '2rem',
              color: '#4ecdc4',
              fontSize: '0.9rem',
              fontWeight: '500',
            }}
          >
            <Sparkles size={16} />
            Open Source PDF Generation Engine
          </div>

          {/* Main Title */}
          <h1
            className="hero-title gradient-text animate-fadeInUp"
            style={{
              animationDelay: '0.1s',
            }}
          >
            GoPdfSuit
          </h1>

          {/* Typewriter subtitle */}
          <div
            className="hero-subtitle animate-fadeInUp"
            style={{
              marginBottom: '3rem',
              color: 'hsl(var(--muted-foreground))',
              animationDelay: '0.2s',
              minHeight: '4rem',
            }}
          >
            <span style={{ position: 'relative' }}>
              {typewriterText}
              <span
                style={{
                  opacity: typewriterText.length < fullText.length ? 1 : 0,
                  animation: 'blink 1s infinite',
                  marginLeft: '2px',
                  color: '#4ecdc4',
                }}
              >
                |
              </span>
            </span>
          </div>

          {/* CTA Buttons */}
          <div
            className="animate-fadeInUp"
            style={{
              display: 'flex',
              gap: '1.5rem',
              justifyContent: 'center',
              flexWrap: 'wrap',
              marginBottom: '4rem',
              animationDelay: '0.3s',
            }}
          >
            <Link
              to="/viewer"
              className="btn-glow glow-on-hover"
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.75rem',
                textDecoration: 'none',
              }}
            >
              <FileText size={20} />
              Try PDF Generator
              <ArrowRight size={18} />
            </Link>
            <a
              href="https://github.com/chinmay-sawant/gopdfsuit"
              target="_blank"
              rel="noopener noreferrer"
              className="btn-outline-glow"
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.75rem',
                textDecoration: 'none',
              }}
            >
              <Github size={20} />
              View on GitHub
              <Star size={16} />
            </a>
          </div>

          {/* Quick Stats - Glass Cards */}
          <div
            className="grid grid-3"
            style={{ marginTop: '2rem' }}
          >
            {highlights.map((highlight, index) => (
              <div
                key={index}
                className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-hero'] ? 'visible' : ''}`}
                style={{
                  textAlign: 'center',
                  padding: '2rem 1.5rem',
                  animationDelay: `${0.4 + index * 0.15}s`,
                }}
              >
                <div
                  className={`feature-icon-box ${index === 0 ? 'teal' : index === 1 ? 'blue' : 'purple'}`}
                  style={{
                    margin: '0 auto 1rem',
                  }}
                >
                  {React.cloneElement(highlight.icon, {
                    size: 28,
                  })}
                </div>
                <h3 style={{
                  marginBottom: '0.5rem',
                  fontSize: '1.3rem',
                  fontWeight: '700',
                }}>
                  {highlight.title}
                </h3>
                <p style={{
                  color: 'hsl(var(--muted-foreground))',
                  marginBottom: 0,
                  fontSize: '0.95rem',
                  lineHeight: '1.6',
                }}>
                  {highlight.desc}
                </p>
              </div>
            ))}
          </div>
        </div>

        {/* Scroll indicator */}
        <div
          className="scroll-indicator"
          onClick={() => document.getElementById('section-features')?.scrollIntoView({ behavior: 'smooth' })}
        >
          <ChevronDown size={32} color="#4ecdc4" />
        </div>
      </section>

      {/* Features Section */}
      <section
        id="section-features"
        style={{ padding: '5rem 0' }}
      >
        <div className="container">
          <div
            className={`text-center animate-fadeInUp stagger-animation ${isVisible['section-features'] ? 'visible' : ''}`}
            style={{ marginBottom: '3rem' }}
          >
            <h2
              className="gradient-text"
              style={{
                fontSize: '2.5rem',
                marginBottom: '1rem',
              }}
            >
              Powerful Features
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.1rem',
              maxWidth: '600px',
              margin: '0 auto',
            }}>
              Everything you need for professional PDF workflows
            </p>
          </div>

          <div className="bento-grid">
            {features.map((feature, index) => {
              const sizeClass = feature.size === 'large' ? 'bento-item-large' :
                feature.size === 'wide' ? 'bento-item-wide' : '';
              return (
                <Link
                  key={index}
                  to={feature.link}
                  style={{ textDecoration: 'none', color: 'inherit' }}
                  className={sizeClass}
                >
                  <div
                    className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-features'] ? 'visible' : ''}`}
                    style={{
                      height: '100%',
                      padding: feature.size === 'large' ? '2.5rem' : '2rem',
                      cursor: 'pointer',
                      animationDelay: `${0.2 + index * 0.1}s`,
                      display: 'flex',
                      flexDirection: 'column',
                    }}
                  >
                    <div
                      className={`feature-icon-box ${feature.color}`}
                      style={{ marginBottom: '1.5rem' }}
                    >
                      {feature.icon}
                    </div>
                    <h3 style={{
                      marginBottom: '0.75rem',
                      color: 'hsl(var(--foreground))',
                      fontSize: feature.size === 'large' ? '1.5rem' : '1.25rem',
                      fontWeight: '700',
                    }}>
                      {feature.title}
                    </h3>
                    <p style={{
                      color: 'hsl(var(--muted-foreground))',
                      marginBottom: '1rem',
                      lineHeight: 1.7,
                      flex: 1,
                      fontSize: feature.size === 'large' ? '1.05rem' : '0.95rem',
                    }}>
                      {feature.description}
                    </p>
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      color: '#4ecdc4',
                      fontSize: '0.9rem',
                      fontWeight: '600',
                    }}>
                      Try it now
                      <ArrowRight size={16} />
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        </div>
      </section>

      <div className="section-divider container" />

      {/* Quick Start Section */}
      <section
        id="section-quickstart"
        style={{ padding: '5rem 0' }}
      >
        <div className="container">
          <div className="split-layout">
            {/* Left side - Text content */}
            <div
              className={`animate-slideInLeft stagger-animation ${isVisible['section-quickstart'] ? 'visible' : ''}`}
            >
              <h2
                className="gradient-text"
                style={{
                  fontSize: '2.5rem',
                  marginBottom: '1.5rem',
                }}
              >
                Get Started in Seconds
              </h2>
              <p style={{
                color: 'hsl(var(--muted-foreground))',
                fontSize: '1.1rem',
                marginBottom: '2rem',
                lineHeight: '1.7',
              }}>
                Clone the repository and start generating PDFs immediately.
                No complex setup required.
              </p>

              <ul className="check-list">
                <li>Zero external dependencies</li>
                <li>Single binary deployment</li>
                <li>Docker ready out of the box</li>
                <li>Cross-platform support</li>
              </ul>

              <Link
                to="/editor"
                className="btn-glow"
                style={{
                  marginTop: '2rem',
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.75rem',
                  textDecoration: 'none',
                }}
              >
                <Edit size={20} />
                Open Editor
              </Link>
            </div>

            {/* Right side - Terminal mockup */}
            <div
              className={`terminal-window animate-slideInRight stagger-animation ${isVisible['section-quickstart'] ? 'visible' : ''}`}
              style={{ animationDelay: '0.2s' }}
            >
              <div className="terminal-header">
                <span className="terminal-dot red"></span>
                <span className="terminal-dot yellow"></span>
                <span className="terminal-dot green"></span>
                <span style={{ color: '#888', marginLeft: '1rem', fontSize: '0.85rem' }}>Terminal</span>
              </div>
              <div className="terminal-body">
                <div style={{ marginBottom: '0.5rem' }}>
                  <span className="terminal-prompt">$ </span>
                  <span className="terminal-command">git clone https://github.com/chinmay-sawant/gopdfsuit.git</span>
                </div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <span className="terminal-prompt">$ </span>
                  <span className="terminal-command">cd gopdfsuit</span>
                </div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <span className="terminal-prompt">$ </span>
                  <span className="terminal-command">make run</span>
                </div>
                <div style={{ marginTop: '1rem' }}>
                  <span className="terminal-success">✓ Server listening on http://localhost:8080</span>
                </div>
                <div style={{ marginTop: '0.5rem' }}>
                  <span className="terminal-success">✓ Ready for PDF generation!</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div className="section-divider container" />

      {/* API Overview */}
      <section
        id="section-api"
        style={{ padding: '5rem 0' }}
      >
        <div className="container">
          <div
            className={`text-center animate-fadeInUp stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
            style={{ marginBottom: '3rem' }}
          >
            <h2
              className="gradient-text"
              style={{ fontSize: '2.5rem', marginBottom: '1rem' }}
            >
              API Endpoints
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.1rem',
            }}>
              RESTful API for seamless integration
            </p>
          </div>

          <div className="grid grid-2">
            <div
              className={`glass-card animate-slideInLeft stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
              style={{ padding: '2rem' }}
            >
              <div className="feature-icon-box blue" style={{ marginBottom: '1.5rem' }}>
                <Zap size={28} />
              </div>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', fontSize: '1.4rem' }}>REST API</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                {[
                  { method: 'POST', path: '/api/v1/generate/template-pdf', desc: 'Generate PDF' },
                  { method: 'POST', path: '/api/v1/merge', desc: 'Merge PDFs' },
                  { method: 'POST', path: '/api/v1/split', desc: 'Split PDFs' },
                  { method: 'POST', path: '/api/v1/fill', desc: 'Fill forms' },
                  { method: 'GET', path: '/api/v1/template-data', desc: 'Get template data' },
                  { method: 'GET', path: '/api/v1/fonts', desc: 'List fonts' },
                  { method: 'POST', path: '/api/v1/htmltopdf', desc: 'HTML to PDF' },
                  { method: 'POST', path: '/api/v1/htmltoimage', desc: 'HTML to Image' }
                ].map((api, index) => (
                  <div
                    key={index}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.75rem',
                      padding: '0.5rem 0',
                      borderBottom: index < 7 ? '1px solid rgba(255,255,255,0.05)' : 'none',
                    }}
                  >
                    <span style={{
                      background: 'rgba(0, 122, 204, 0.2)',
                      color: '#007acc',
                      padding: '0.2rem 0.5rem',
                      borderRadius: '4px',
                      fontSize: '0.7rem',
                      fontWeight: '700',
                    }}>
                      {api.method}
                    </span>
                    <code style={{ color: '#4ecdc4', fontSize: '0.85rem', flex: 1 }}>{api.path}</code>
                  </div>
                ))}
              </div>
            </div>

            <div
              className={`glass-card animate-slideInRight stagger-animation ${isVisible['section-api'] ? 'visible' : ''}`}
              style={{ padding: '2rem', animationDelay: '0.2s' }}
            >
              <div className="feature-icon-box purple" style={{ marginBottom: '1.5rem' }}>
                <Globe size={28} />
              </div>
              <h3 style={{ color: 'hsl(var(--foreground))', marginBottom: '1.5rem', fontSize: '1.4rem' }}>Web Interfaces</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                {[
                  { path: '/viewer', desc: 'PDF Viewer & Template Processor' },
                  { path: '/editor', desc: 'Drag-and-drop Template Editor' },
                  { path: '/merge', desc: 'PDF Merge Interface' },
                  { path: '/split', desc: 'PDF Split Interface' },
                  { path: '/filler', desc: 'PDF Form Filler' },
                  { path: '/htmltopdf', desc: 'HTML to PDF Converter' },
                  { path: '/htmltoimage', desc: 'HTML to Image Converter' },
                  { path: '/screenshots', desc: 'Screenshots Page' },
                  { path: '/comparison', desc: 'Feature Comparison' }
                ].map((route, index) => (
                  <div
                    key={index}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.75rem',
                      padding: '0.5rem 0',
                      borderBottom: index < 8 ? '1px solid rgba(255,255,255,0.05)' : 'none',
                    }}
                  >
                    <span style={{
                      background: 'rgba(240, 147, 251, 0.2)',
                      color: '#f093fb',
                      padding: '0.2rem 0.5rem',
                      borderRadius: '4px',
                      fontSize: '0.7rem',
                      fontWeight: '700',
                    }}>
                      GET
                    </span>
                    <code style={{ color: '#4ecdc4', fontSize: '0.85rem' }}>{route.path}</code>
                    <span style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.85rem' }}>- {route.desc}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      <div className="section-divider container" />

      {/* Performance Section */}
      <section
        id="section-performance"
        style={{ padding: '4rem 0' }}
      >
        <div className="container">
          <div
            className={`card card-hover animate-fadeInScale stagger-animation ${isVisible['section-performance'] ? 'visible' : ''}`}
            style={{
              textAlign: 'center',
              maxWidth: '800px',
              margin: '0 auto',
              animationDelay: '0.2s',
            }}
          >
            <h2
              className={`animate-fadeInUp stagger-animation ${isVisible['section-performance'] ? 'visible' : ''}`}
              style={{
                color: 'hsl(var(--foreground))',
                marginBottom: '1rem',
                animationDelay: '0.4s',
              }}
            >
              🏃‍♂️ Performance
            </h2>
            <p
              className={`animate-fadeInUp stagger-animation ${isVisible['section-performance'] ? 'visible' : ''}`}
              style={{
                color: 'hsl(var(--muted-foreground))',
                marginBottom: '2rem',
                animationDelay: '0.6s',
              }}
            >
              Ultra-fast PDF generation with in-memory processing
            </p>

            {/* Performance Stats */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '1rem', marginBottom: '2rem' }}>
              {[
                { value: '42.3 ms', label: 'Average Response', color: '#4ecdc4', bg: 'rgba(78, 205, 196, 0.1)', border: 'rgba(78, 205, 196, 0.3)' },
                { value: '39.0 ms', label: 'Min Response', color: '#007acc', bg: 'rgba(0, 122, 204, 0.1)', border: 'rgba(0, 122, 204, 0.3)' },
                { value: '66.8 ms', label: 'Max Response', color: '#ffc107', bg: 'rgba(255, 193, 7, 0.1)', border: 'rgba(255, 193, 7, 0.3)' }
              ].map((stat, index) => (
                <div
                  key={index}
                  className={`animate-fadeInScale stagger-animation ${isVisible['section-performance'] ? 'visible' : ''}`}
                  style={{
                    background: stat.bg,
                    padding: '1rem',
                    borderRadius: '8px',
                    border: `1px solid ${stat.border}`,
                    transition: 'all 0.3s ease',
                    animationDelay: `${0.8 + index * 0.2}s`,
                  }}
                >
                  <div
                    className="animate-pulse"
                    style={{
                      fontSize: '1.5rem',
                      fontWeight: 'bold',
                      color: stat.color,
                      animationDelay: `${2 + index * 0.5}s`,
                    }}
                  >
                    {stat.value}
                  </div>
                  <div style={{ fontSize: '0.8rem', color: 'hsl(var(--muted-foreground))' }}>
                    {stat.label}
                  </div>
                </div>
              ))}
            </div>

            {/* Sample Logs */}
            <div style={{
              background: 'hsl(var(--card))',
              border: '1px solid hsl(var(--border))',
              padding: '1rem',
              borderRadius: '8px',
              fontFamily: 'monospace',
              color: '#4ecdc4',
              fontSize: '0.8rem',
              textAlign: 'left',
              maxHeight: '200px',
              overflowY: 'auto',
              scrollbarWidth: 'thin',
              scrollbarColor: 'rgba(78, 205, 196, 0.5) hsl(var(--card))',
            }}
              className="custom-scrollbar"
            >
              <div style={{ marginBottom: '0.5rem', fontWeight: 'bold' }}>Recent Performance Logs:</div>
              [GIN] 2026/01/19 - 22:45:10 | 200 |      41.25ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:11 | 200 |      43.82ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:12 | 200 |      45.12ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:13 | 200 |      66.79ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:13 | 200 |      42.05ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:14 | 200 |      39.56ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:14 | 200 |      40.11ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:15 | 200 |      44.30ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:15 | 200 |      42.98ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:16 | 200 |      41.77ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:16 | 200 |      48.55ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:17 | 200 |      52.12ms |             ::1 | POST     "/api/v1/generate/template-pdf"<br />
              [GIN] 2026/01/19 - 22:45:17 | 200 |      40.88ms |             ::1 | POST     "/api/v1/generate/template-pdf"
            </div>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              marginTop: '1rem',
              fontSize: '0.85rem',
              marginBottom: 0,
              lineHeight: 1.6,
            }}>
              Benchmarks for PDF generation with PDF/A compliance, font embedding, digital signatures, bookmarks, and internal links.
              <br />
              <span style={{ fontSize: '0.75rem', fontStyle: 'italic', opacity: 0.8 }}>
                * Results may vary based on selected options, hardware configuration, data complexity, and network conditions.
              </span>
            </p>
          </div>
        </div>
      </section>

      {/* Comparison Preview Section */}
      <section
        id="section-comparison-preview"
        style={{ padding: '5rem 0' }}
      >
        <div className="container">
          <div
            className={`text-center animate-fadeInUp stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
            style={{ marginBottom: '3rem' }}
          >
            <h2
              className="gradient-text"
              style={{ fontSize: '2.5rem', marginBottom: '1rem' }}
            >
              Why Choose GoPdfSuit?
            </h2>
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1.1rem',
              maxWidth: '700px',
              margin: '0 auto',
            }}>
              Enterprise features at zero cost — compare with iTextPDF, PDFLib, and commercial solutions
            </p>
          </div>

          {/* Quick Stats */}
          <div
            className="grid grid-3"
            style={{ marginBottom: '2.5rem', maxWidth: '900px', margin: '0 auto 2.5rem' }}
          >
            {[
              { value: 'Free', label: 'vs $2K-4K/dev/year', color: '#4ecdc4', icon: <CheckCircle size={28} /> },
              { value: '< 100ms', label: 'Response time', color: '#007acc', icon: <Zap size={28} /> },
              { value: '0 deps', label: 'Pure Go binary', color: '#f093fb', icon: <Download size={28} /> }
            ].map((stat, index) => (
              <div
                key={index}
                className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
                style={{
                  textAlign: 'center',
                  padding: '1.5rem',
                  animationDelay: `${0.2 + index * 0.1}s`,
                }}
              >
                <div style={{ color: stat.color, marginBottom: '0.75rem', display: 'flex', justifyContent: 'center' }}>
                  {stat.icon}
                </div>
                <div className="stat-value" style={{ color: stat.color, marginBottom: '0.25rem', fontSize: '1.8rem' }}>
                  {stat.value}
                </div>
                <div style={{ fontSize: '0.85rem', color: 'hsl(var(--muted-foreground))' }}>
                  {stat.label}
                </div>
              </div>
            ))}
          </div>

          {/* Feature Comparison */}
          <div
            className={`glass-card animate-fadeInScale stagger-animation ${isVisible['section-comparison-preview'] ? 'visible' : ''}`}
            style={{ width: '100%', padding: '2.5rem' }}
          >
            <h3 style={{
              color: 'hsl(var(--foreground))',
              marginBottom: '1.5rem',
              fontSize: '1.3rem',
              textAlign: 'center',
            }}>
              Built-in Enterprise Features
            </h3>

            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
              gap: '1rem',
              marginBottom: '2rem',
            }}>
              {[
                { name: 'PDF/A-4 Compliance', desc: 'Archival standard with sRGB ICC profiles', color: '#4ecdc4' },
                { name: 'PDF/UA-2 Accessibility', desc: 'Universal accessibility compliance', color: '#007acc' },
                { name: 'AES-128 Encryption', desc: 'Password protection with permissions', color: '#f093fb' },
                { name: 'Digital Signatures', desc: 'PKCS#7 certificates with visual appearance', color: '#ffc107' },
                { name: 'Font Subsetting', desc: 'TrueType embedding with glyph optimization', color: '#4ecdc4' },
                { name: 'PDF Merge', desc: 'Combine multiple PDFs, preserve forms', color: '#007acc' },
                { name: 'XFDF Form Filling', desc: 'Advanced field detection and population', color: '#f093fb' },
                { name: 'Bookmarks & Links', desc: 'Outlines with internal/external hyperlinks', color: '#ffc107' },
              ].map((feature, index) => (
                <div
                  key={index}
                  style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: '0.75rem',
                    padding: '0.75rem',
                    background: 'rgba(255,255,255,0.02)',
                    borderRadius: '8px',
                    border: '1px solid rgba(255,255,255,0.05)',
                  }}
                >
                  <CheckCircle size={18} style={{ color: feature.color, flexShrink: 0, marginTop: '2px' }} />
                  <div>
                    <div style={{ color: 'hsl(var(--foreground))', fontWeight: '600', fontSize: '0.9rem' }}>
                      {feature.name}
                    </div>
                    <div style={{ color: 'hsl(var(--muted-foreground))', fontSize: '0.8rem' }}>
                      {feature.desc}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            <div style={{ textAlign: 'center' }}>
              <Link
                to="/comparison"
                className="btn-glow"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.75rem',
                  textDecoration: 'none',
                }}
              >
                View Full Comparison
                <ArrowRight size={18} />
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer
        id="section-footer"
        style={{
          padding: '4rem 0 2rem',
          marginTop: '2rem',
          background: 'linear-gradient(0deg, rgba(78,205,196,0.03) 0%, transparent 100%)',
        }}
      >
        <div className="container">
          <div className="section-divider" style={{ margin: '0 0 3rem' }} />

          <div
            className={`animate-fadeInUp stagger-animation ${isVisible['section-footer'] ? 'visible' : ''}`}
            style={{
              textAlign: 'center',
            }}
          >
            {/* Quick Links */}
            <div style={{
              display: 'flex',
              justifyContent: 'center',
              gap: '2rem',
              marginBottom: '2rem',
              flexWrap: 'wrap',
            }}>
              <a
                href="https://github.com/chinmay-sawant/gopdfsuit"
                target="_blank"
                rel="noopener noreferrer"
                className="btn-outline-glow"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  textDecoration: 'none',
                  padding: '0.75rem 1.5rem',
                }}
              >
                <Github size={18} />
                GitHub
              </a>
              <Link
                to="/viewer"
                className="btn-outline-glow"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  textDecoration: 'none',
                  padding: '0.75rem 1.5rem',
                }}
              >
                <FileText size={18} />
                Documentation
              </Link>
            </div>

            {/* Credits */}
            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '1rem',
              marginBottom: '0.5rem',
            }}>
              Made with ❤️ and ☕ by{' '}
              <a
                href="https://github.com/chinmay-sawant"
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  color: '#4ecdc4',
                  textDecoration: 'none',
                  fontWeight: '600',
                }}
              >
                Chinmay Sawant
              </a>
            </p>

            <p style={{
              color: 'hsl(var(--muted-foreground))',
              fontSize: '0.9rem',
              marginBottom: 0,
              opacity: 0.7,
            }}>
              <Star size={14} style={{ display: 'inline', marginRight: '0.5rem', color: '#ffc107' }} />
              Star this repo if you find it helpful!
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default Home