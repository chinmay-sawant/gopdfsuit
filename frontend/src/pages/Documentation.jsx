import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Sidebar } from '../components/documentation/Sidebar'
import { DocContent } from '../components/documentation/DocContent'
import { docSections } from '../components/documentation/content'

const Documentation = () => {
    const [searchParams, setSearchParams] = useSearchParams()
    
    // Find item based on URL or default to first
    const [activeItem, setActiveItem] = useState(() => {
        const itemParam = searchParams.get('item')
        if (itemParam) {
            for (const section of docSections) {
                const item = section.items.find(i => i.id === itemParam || i.id === itemParam.toLowerCase())
                if (item) return item
            }
        }
        return docSections[0].items[0]
    })

    // Sync state with URL if it changes externally
    useEffect(() => {
        const itemParam = searchParams.get('item')
        if (itemParam) {
            for (const section of docSections) {
                const item = section.items.find(i => i.id === itemParam || i.id === itemParam.toLowerCase())
                if (item && item.id !== activeItem?.id) {
                    setActiveItem(item)
                    break
                }
            }
        }
    }, [searchParams, activeItem?.id])

    const handleItemClick = (item) => {
        setActiveItem(item)
        setSearchParams({ item: item.id })
    }

    // Scroll to top when item changes
    useEffect(() => {
        const contentContainer = document.getElementById('doc-content-scrollArea')
        const codeContainer = document.getElementById('doc-code-scrollArea')
        
        if (contentContainer) {
            contentContainer.scrollTo({ top: 0, behavior: 'smooth' })
        }
        if (codeContainer) {
            codeContainer.scrollTo({ top: 0, behavior: 'smooth' })
        }
        
        // Also ensure sidebar item is visible
        const sidebarItem = document.getElementById(`sidebar-item-${activeItem?.id}`)
        if (sidebarItem) {
            sidebarItem.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
        }
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
                onItemClick={handleItemClick}
            />
            <div style={{ flex: 1, height: '100%', overflow: 'hidden' }}>
                <DocContent item={activeItem} />
            </div>
        </div>
    )
}

export default Documentation
