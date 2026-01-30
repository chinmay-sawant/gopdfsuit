import { useEffect, useState } from 'react'
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react'

export default function Toast({ message, type = 'success', duration = 3000, onClose }) {
    const [isVisible, setIsVisible] = useState(true)
    const [isExiting, setIsExiting] = useState(false)

    useEffect(() => {
        const timer = setTimeout(() => {
            setIsExiting(true)
            setTimeout(() => {
                setIsVisible(false)
                onClose?.()
            }, 300)
        }, duration)

        return () => clearTimeout(timer)
    }, [duration, onClose])

    const handleClose = () => {
        setIsExiting(true)
        setTimeout(() => {
            setIsVisible(false)
            onClose?.()
        }, 300)
    }

    if (!isVisible) return null

    const icons = {
        success: <CheckCircle size={20} />,
        error: <AlertCircle size={20} />,
        info: <Info size={20} />
    }

    const colors = {
        success: {
            bg: 'hsl(142 71% 45%)',
            icon: 'hsl(142 71% 45%)',
            text: 'white'
        },
        error: {
            bg: 'hsl(0 84.2% 60.2%)',
            icon: 'hsl(0 84.2% 60.2%)',
            text: 'white'
        },
        info: {
            bg: 'hsl(199 89% 48%)',
            icon: 'hsl(199 89% 48%)',
            text: 'white'
        }
    }

    const colorScheme = colors[type] || colors.info

    return (
        <div
            style={{
                position: 'fixed',
                top: '80px',
                right: '20px',
                zIndex: 9999,
                minWidth: '300px',
                maxWidth: '400px',
                background: 'hsl(var(--card))',
                border: '1px solid hsl(var(--border))',
                borderRadius: '8px',
                boxShadow: '0 10px 25px rgba(0, 0, 0, 0.15)',
                padding: '1rem',
                display: 'flex',
                alignItems: 'flex-start',
                gap: '0.75rem',
                animation: isExiting ? 'slideOut 0.3s ease-out forwards' : 'slideIn 0.3s ease-out',
                transform: isExiting ? 'translateX(100%)' : 'translateX(0)',
                opacity: isExiting ? 0 : 1,
                transition: 'all 0.3s ease-out'
            }}
        >
            <div style={{ color: colorScheme.icon, flexShrink: 0 }}>
                {icons[type]}
            </div>
            <div style={{ flex: 1, fontSize: '0.9rem', color: 'hsl(var(--foreground))', lineHeight: '1.5' }}>
                {message}
            </div>
            <button
                onClick={handleClose}
                style={{
                    background: 'transparent',
                    border: 'none',
                    color: 'hsl(var(--muted-foreground))',
                    cursor: 'pointer',
                    padding: '0',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    transition: 'color 0.2s'
                }}
                onMouseEnter={(e) => e.currentTarget.style.color = 'hsl(var(--foreground))'}
                onMouseLeave={(e) => e.currentTarget.style.color = 'hsl(var(--muted-foreground))'}
            >
                <X size={16} />
            </button>

            <style jsx>{`
                @keyframes slideIn {
                    from {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                    to {
                        transform: translateX(0);
                        opacity: 1;
                    }
                }
                
                @keyframes slideOut {
                    from {
                        transform: translateX(0);
                        opacity: 1;
                    }
                    to {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                }
            `}</style>
        </div>
    )
}

// Toast Container Component to manage multiple toasts
export function ToastContainer({ toasts, removeToast }) {
    return (
        <>
            {toasts.map((toast, index) => (
                <div key={toast.id} style={{ top: `${80 + index * 100}px`, position: 'fixed', right: '20px' }}>
                    <Toast
                        message={toast.message}
                        type={toast.type}
                        duration={toast.duration}
                        onClose={() => removeToast(toast.id)}
                    />
                </div>
            ))}
        </>
    )
}
