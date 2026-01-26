import json
import requests
from pathlib import Path
from models import PdfRequest

# Configuration
API_URL = "http://localhost:8080/api/v1/generate/template-pdf"
# Use path relative to this script
BASE_DIR = Path(__file__).parent
JSON_SOURCE = BASE_DIR / "../editor/financial_digitalsignature.json"
OUTPUT_FILE = BASE_DIR / "output.pdf"

def main():
    if not JSON_SOURCE.exists():
        print(f"Error: Source file {JSON_SOURCE} not found.")
        return

    print(f"Reading JSON from {JSON_SOURCE}...")
    try:
        with open(JSON_SOURCE, "r") as f:
            data = json.load(f)
    except json.JSONDecodeError as e:
        print(f"Error decoding JSON: {e}")
        return

    print("Validating data with Pydantic models...")
    try:
        # validate the data against the Pydantic model
        request_model = PdfRequest(**data)
        # convert back to json compatible dict (handles any defaults or conversions)
        payload = request_model.model_dump(mode='json') # use .dict() for pydantic v1
    except Exception as e:
        print(f"Validation Error: {e}")
        return

    print(f"Sending request to {API_URL}...")
    try:
        response = requests.post(API_URL, json=payload)
        
        if response.status_code == 200:
            print(f"Success! Saving PDF to {OUTPUT_FILE}...")
            with open(OUTPUT_FILE, "wb") as f:
                f.write(response.content)
            print("Done.")
        else:
            print(f"Request failed with status code: {response.status_code}")
            print(f"Response: {response.text}")
            
    except requests.RequestException as e:
        print(f"Connection Error: {e}")
        print("Ensure the API server is running at http://localhost:8080")

if __name__ == "__main__":
    main()
