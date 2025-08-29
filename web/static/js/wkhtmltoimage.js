// WKHTML to Image Converter JavaScript
document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('imageForm');
    const inputType = document.getElementById('inputType');
    const htmlGroup = document.getElementById('htmlGroup');
    const urlGroup = document.getElementById('urlGroup');
    const previewBtn = document.getElementById('previewBtn');
    const previewModal = document.getElementById('previewModal');
    const closeBtn = document.querySelector('.close');
    const previewFrame = document.getElementById('previewFrame');
    const qualityInput = document.getElementById('quality');

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

    // Quality indicator
    qualityInput.addEventListener('input', function() {
        updateQualityIndicator(this.value);
    });

    function updateQualityIndicator(quality) {
        let indicator = document.querySelector('.quality-indicator');
        if (!indicator) {
            const qualityGroup = qualityInput.parentNode;
            indicator = document.createElement('div');
            indicator.className = 'quality-indicator';
            indicator.innerHTML = `
                <span>Quality: ${quality}%</span>
                <div class="quality-bar">
                    <div class="quality-fill"></div>
                </div>
            `;
            qualityGroup.appendChild(indicator);
        } else {
            indicator.querySelector('span').textContent = `Quality: ${quality}%`;
        }

        const fill = indicator.querySelector('.quality-fill');
        fill.style.width = `${quality}%`;

        // Change color based on quality
        if (quality >= 80) {
            fill.style.background = '#10b981';
        } else if (quality >= 60) {
            fill.style.background = '#f59e0b';
        } else {
            fill.style.background = '#ef4444';
        }
    }

    // Initialize quality indicator
    updateQualityIndicator(qualityInput.value);

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

            const response = await fetch('/api/v1/wkhtmltoimage', {
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
            const format = document.getElementById('format').value;
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `converted.${format}`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);

            showMessage('Image conversion successful!', 'success');

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
            format: document.getElementById('format').value,
            quality: parseInt(document.getElementById('quality').value),
            zoom: parseFloat(document.getElementById('zoom').value)
        };

        // Optional dimensions
        const width = document.getElementById('width').value;
        const height = document.getElementById('height').value;
        if (width) data.width = parseInt(width);
        if (height) data.height = parseInt(height);

        // Optional crop settings
        const cropWidth = document.getElementById('cropWidth').value;
        const cropHeight = document.getElementById('cropHeight').value;
        const cropX = document.getElementById('cropX').value;
        const cropY = document.getElementById('cropY').value;
        if (cropWidth) data.crop_width = parseInt(cropWidth);
        if (cropHeight) data.crop_height = parseInt(cropHeight);
        if (cropX) data.crop_x = parseInt(cropX);
        if (cropY) data.crop_y = parseInt(cropY);

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
    <title>Sample Image</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            text-align: center;
            padding: 40px;
        }
        h1 { font-size: 2.5em; margin-bottom: 20px; }
        .content { font-size: 1.2em; line-height: 1.6; }
        .highlight { background: rgba(255,255,255,0.2); padding: 20px; border-radius: 10px; }
    </style>
</head>
<body>
    <h1>Sample Image Document</h1>
    <div class="content">
        <div class="highlight">
            <p>This HTML will be converted to an image.</p>
            <p>You can use CSS for styling and layout.</p>
        </div>
    </div>
</body>
</html>`;
});
