import { useEffect, useRef } from 'react'

const BackgroundAnimation = () => {
  const canvasRef = useRef(null);

  useEffect(() => {
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
        this.size = Math.random() * 2 + 1; // Smaller particles
        this.speedX = (Math.random() - 0.5) * 0.2; // Slower speed
        this.speedY = (Math.random() - 0.5) * 0.2; // Slower speed
        this.density = Math.random() * 30 + 1;
        // Color palette: Subtle professional colors (slate/blueish)
        const colors = [
          `rgba(78, 205, 196, ${Math.random() * 0.3 + 0.1})`, // Reduced opacity
          `rgba(0, 122, 204, ${Math.random() * 0.2 + 0.1})`, // Reduced opacity
          `rgba(150, 150, 200, ${Math.random() * 0.2 + 0.1})`, // More neutral
        ];
        this.color = colors[Math.floor(Math.random() * colors.length)];
      }

      update() {
        // Gradually increase movement speed over first 60 frames
        const speedMultiplier = Math.min(1, frameCount / 60);

        // Move particles with wave motion (gentler)
        this.x += (this.speedX + Math.sin(Date.now() * 0.001 + this.baseY * 0.01) * 0.1) * speedMultiplier;
        this.y += (this.speedY + Math.cos(Date.now() * 0.001 + this.baseX * 0.01) * 0.1) * speedMultiplier;

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
            this.x -= Math.cos(angle) * force * 2; // Gentler repel
            this.y -= Math.sin(angle) * force * 2;
          }
        }
      }

      draw(globalOpacity) {
        ctx.beginPath();
        ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2);
        // Further reduce global opacity for a very subtle texture
        ctx.fillStyle = this.color.replace(/[\d.]+\)$/, `${parseFloat(this.color.match(/[\d.]+\)$/)[0]) * globalOpacity * 0.6})`);
        ctx.fill();
      }
    }

    // Create particles
    const particleCount = Math.min(80, Math.floor((canvas.width * canvas.height) / 15000)); // Fewer particles
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
            // Very faint lines
            ctx.strokeStyle = `rgba(150, 160, 180, ${0.08 * (1 - distance / 120) * globalOpacity})`;
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

export default BackgroundAnimation
