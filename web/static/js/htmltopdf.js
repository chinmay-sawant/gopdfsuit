// HTML to PDF Converter JavaScript (powered by gochromedp)
document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('pdfForm');
    const inputType = document.getElementById('inputType');
    const htmlGroup = document.getElementById('htmlGroup');
    const urlGroup = document.getElementById('urlGroup');
    const previewBtn = document.getElementById('previewBtn');
    const previewModal = document.getElementById('previewModal');
    const closeBtn = document.querySelector('.close');
    const previewFrame = document.getElementById('previewFrame');

    // Toggle input type
    inputType.addEventListener('change', function() {
        if (this.value === 'html') {
            htmlGroup.style.display = 'block';
            urlGroup.style.display = 'none';
        } else {
            htmlGroup.style.display = 'none';
            urlGroup.style.display = 'block';
        }
    });

    // Preview HTML
    previewBtn.addEventListener('click', function() {
        const inputType = document.getElementById('inputType').value;
        let htmlContent = '';

        if (inputType === 'html') {
            htmlContent = document.getElementById('htmlContent').value;
        } else {
            const url = document.getElementById('urlInput').value;
            if (!url) {
                showMessage('Please enter a URL', 'error');
                return;
            }
            // For URL preview, we'll create an iframe that loads the URL
            previewFrame.src = url;
            previewModal.style.display = 'block';
            return;
        }

        if (!htmlContent.trim()) {
            showMessage('Please enter HTML content', 'error');
            return;
        }

        // Create a blob URL for the HTML content
        const blob = new Blob([htmlContent], { type: 'text/html' });
        const url = URL.createObjectURL(blob);
        previewFrame.src = url;
        previewModal.style.display = 'block';
    });

    // Close modal
    closeBtn.addEventListener('click', function() {
        previewModal.style.display = 'none';
        previewFrame.src = '';
    });

    // Close modal when clicking outside
    window.addEventListener('click', function(event) {
        if (event.target === previewModal) {
            previewModal.style.display = 'none';
            previewFrame.src = '';
        }
    });

    // Form submission
    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        const submitBtn = form.querySelector('.convert-btn');
        const originalText = submitBtn.textContent;
        submitBtn.textContent = 'Converting...';
        submitBtn.classList.add('loading');
        submitBtn.disabled = true;

        try {
            const formData = getFormData();

            const response = await fetch('/api/v1/htmltopdf', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Conversion failed');
            }

            // Create download link
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'converted.pdf';
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);

            showMessage('PDF conversion successful!', 'success');

        } catch (error) {
            console.error('Conversion error:', error);
            showMessage('Conversion failed: ' + error.message, 'error');
        } finally {
            submitBtn.textContent = originalText;
            submitBtn.classList.remove('loading');
            submitBtn.disabled = false;
        }
    });

    function getFormData() {
        const inputType = document.getElementById('inputType').value;
        const data = {
            page_size: document.getElementById('pageSize').value,
            orientation: document.getElementById('orientation').value,
            dpi: parseInt(document.getElementById('dpi').value),
            margin_top: document.getElementById('marginTop').value,
            margin_right: document.getElementById('marginRight').value,
            margin_bottom: document.getElementById('marginBottom').value,
            margin_left: document.getElementById('marginLeft').value,
            grayscale: document.getElementById('grayscale').checked,
            low_quality: document.getElementById('lowQuality').checked
        };

        if (inputType === 'html') {
            data.html = document.getElementById('htmlContent').value;
        } else {
            data.url = document.getElementById('urlInput').value;
        }

        return data;
    }

    function showMessage(message, type) {
        // Remove existing messages
        const existingMessages = document.querySelectorAll('.message');
        existingMessages.forEach(msg => msg.remove());

        // Create new message
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${type}`;
        messageDiv.textContent = message;

        // Insert at the top of the form
        const form = document.querySelector('.converter-form');
        form.insertBefore(messageDiv, form.firstChild);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.remove();
            }
        }, 5000);
    }

    // Add some sample HTML for testing
    document.getElementById('htmlContent').value = `<!DOCTYPE html>
<html>
<head>
    <title>Sample Document</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #2563eb; }
        .content { line-height: 1.6; }
    </style>
</head>
<body>
    <h1>Sample PDF Document</h1>
    <div class="content">
        <p>This is a sample HTML document that will be converted to PDF.</p>
        <p>You can include any HTML content here, including:</p>
        <ul>
            <li>Text formatting</li>
            <li>Images</li>
            <li>Tables</li>
            <li>CSS styling</li>
        </ul>
        <p>The PDF will preserve the layout and styling as much as possible.</p>
    </div>
</body>
</html>`;
});
