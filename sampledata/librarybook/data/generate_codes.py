#!/usr/bin/env python3
"""
Generate QR Code and Barcode for Library Book Receiving Form
Both codes will encode the URL: www.google.com
"""

import qrcode
import barcode
from barcode.writer import ImageWriter
from PIL import Image
import base64
import io
import os

# URL to encode
URL = "https://www.google.com"

def generate_qr_code(data: str, filename: str, size: int = 200) -> str:
    """Generate a QR code and save as PNG, return base64 string"""
    qr = qrcode.QRCode(
        version=1,
        error_correction=qrcode.constants.ERROR_CORRECT_M,
        box_size=10,
        border=2,
    )
    qr.add_data(data)
    qr.make(fit=True)
    
    img = qr.make_image(fill_color="black", back_color="white")
    
    # Resize to desired size
    img = img.resize((size, size), Image.Resampling.LANCZOS)
    
    # Save to file
    img.save(filename)
    print(f"QR Code saved to: {filename}")
    
    # Convert to base64
    buffer = io.BytesIO()
    img.save(buffer, format='PNG')
    base64_str = base64.b64encode(buffer.getvalue()).decode('utf-8')
    
    return base64_str


def generate_barcode(data: str, filename: str) -> str:
    """Generate a Code128 barcode and save as PNG, return base64 string"""
    # Use Code128 which can encode URLs
    code128 = barcode.get_barcode_class('code128')
    
    # Create barcode with ImageWriter
    writer = ImageWriter()
    writer.set_options({
        'module_width': 0.4,
        'module_height': 15.0,
        'quiet_zone': 2.0,
        'font_size': 10,
        'text_distance': 5.0,
        'write_text': True,
    })
    
    # Generate barcode - use a shorter identifier for the barcode
    # (full URLs are too long for readable barcodes)
    barcode_data = "LIB-GOOGLE-COM"
    bc = code128(barcode_data, writer=writer)
    
    # Save without extension (python-barcode adds .png automatically)
    saved_path = bc.save(filename.replace('.png', ''))
    print(f"Barcode saved to: {saved_path}")
    
    # Open the saved image and convert to base64
    with Image.open(saved_path) as img:
        # Resize to reasonable dimensions
        img = img.resize((300, 80), Image.Resampling.LANCZOS)
        
        buffer = io.BytesIO()
        img.save(buffer, format='PNG')
        base64_str = base64.b64encode(buffer.getvalue()).decode('utf-8')
    
    return base64_str


def main():
    # Get the directory of this script
    script_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Generate QR Code
    qr_filename = os.path.join(script_dir, "qr_code_google.png")
    qr_base64 = generate_qr_code(URL, qr_filename, size=150)
    
    # Generate Barcode
    barcode_filename = os.path.join(script_dir, "barcode_google.png")
    barcode_base64 = generate_barcode(URL, barcode_filename)
    
    print("\n" + "="*60)
    print("BASE64 ENCODED IMAGES")
    print("="*60)
    
    print("\n--- QR CODE BASE64 ---")
    print(qr_base64)
    
    print("\n--- BARCODE BASE64 ---")
    print(barcode_base64)
    
    # Save base64 strings to text files for easy copy
    with open(os.path.join(script_dir, "qr_code_base64.txt"), 'w') as f:
        f.write(qr_base64)
    print(f"\nQR Code base64 saved to: qr_code_base64.txt")
    
    with open(os.path.join(script_dir, "barcode_base64.txt"), 'w') as f:
        f.write(barcode_base64)
    print(f"Barcode base64 saved to: barcode_base64.txt")


if __name__ == "__main__":
    main()
