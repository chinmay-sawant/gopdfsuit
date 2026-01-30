import { useState, useEffect } from 'react'
import { Sidebar } from '../components/documentation/Sidebar'
import { DocContent } from '../components/documentation/DocContent'
import { docSections } from '../components/documentation/content'

const Documentation = () => {
    const [activeItem, setActiveItem] = useState(docSections[0].items[0])

    // Scroll to top when item changes
    useEffect(() => {
        // This might be handled by the scrolling container in DocContent, 
        // but if we had a global scroll we'd do it here.
    }, [activeItem])

    return (
        <div style={{
            display: 'flex',
            height: 'calc(100vh - 80px)', // Adjusting for Navbar height
            background: 'hsl(var(--background))',
            overflow: 'hidden', // Prevent whole page scroll, let components scroll
            borderTop: '1px solid hsl(var(--border))'
        }}>
            <Sidebar
                sections={docSections}
                activeId={activeItem?.id}
                onItemClick={setActiveItem}
            />
            <div style={{ flex: 1, height: '100%', overflow: 'hidden' }}>
                <DocContent item={activeItem} />
            </div>
        </div>
    )
}

export default Documentation
