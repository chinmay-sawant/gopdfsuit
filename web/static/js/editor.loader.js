// Loader: inject split files in order and initialize the editor
(function(){
    const scripts = [
        'editor.core.js',
        'editor.setup.js',
        'editor.ui.js',
        'editor.model.js',
        'editor.elements.js',
        'editor.table.js'
    ];

    const staticBase = '/static/js/';
    const head = document.getElementsByTagName('head')[0];

    // Load scripts sequentially to preserve order (core first)
    function loadScriptAt(index) {
        if (index >= scripts.length) {
            // all loaded -> initialize
            if (window.TemplateEditor) {
                window.templateEditor = new TemplateEditor();
                try { window.templateEditor.init(); } catch (err) { console.error('Failed to initialize TemplateEditor', err); }
            } else {
                console.error('TemplateEditor is not defined after loading scripts');
            }
            return;
        }

        const filename = scripts[index];
        const s = document.createElement('script');
        s.src = staticBase + filename;
        s.onload = () => loadScriptAt(index + 1);
        s.onerror = () => {
            console.error('Failed to load', filename);
            // attempt to continue loading next scripts to surface more errors
            loadScriptAt(index + 1);
        };
        // ensure ordered execution
        s.async = false;
        head.appendChild(s);
    }

    loadScriptAt(0);
})();
