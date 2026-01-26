import json
import logging
from pathlib import Path
from typing import Union, Optional
import requests
from .models import PdfRequest

class PdfClient:
    def __init__(self, api_url: str = "http://localhost:8080/api/v1/generate/template-pdf"):
        self.api_url = api_url
        self.logger = logging.getLogger(__name__)

    def generate_pdf(self, request_data: Union[PdfRequest, dict]) -> Optional[bytes]:
        """
        Sends a PDF generation request to the API.
        
        Args:
            request_data: PdfRequest model instance or dictionary data
            
        Returns:
            bytes: The PDF content if successful, None otherwise
        """
        # Validate and convert if it's a dict
        if isinstance(request_data, dict):
            try:
                request_data = PdfRequest(**request_data)
            except Exception as e:
                self.logger.error(f"Validation Error: {e}")
                return None

        payload = request_data.model_dump(mode='json')

        self.logger.info(f"Sending request to {self.api_url}...")
        try:
            response = requests.post(self.api_url, json=payload)
            
            if response.status_code == 200:
                self.logger.info("PDF generated successfully.")
                return response.content
            else:
                self.logger.error(f"Request failed with status code: {response.status_code}")
                self.logger.error(f"Response: {response.text}")
                return None
                
        except requests.RequestException as e:
            self.logger.error(f"Connection Error: {e}")
            return None

    def generate_from_file(self, json_path: Union[str, Path]) -> Optional[bytes]:
        """
        Loads JSON from a file and generates a PDF.
        """
        try:
            with open(json_path, "r") as f:
                data = json.load(f)
            return self.generate_pdf(data)
        except Exception as e:
            self.logger.error(f"Error reading file {json_path}: {e}")
            return None
