"""
Pytest configuration and fixtures for pypdfsuit tests.
"""

import pytest


@pytest.fixture
def simple_html():
    """Simple HTML content for testing."""
    return "<html><body><h1>Test</h1></body></html>"


@pytest.fixture
def simple_xfdf():
    """Simple XFDF content for testing."""
    return b"""<?xml version="1.0" encoding="UTF-8"?>
<xfdf xmlns="http://ns.adobe.com/xfdf/">
    <fields>
        <field name="Name"><value>John Doe</value></field>
        <field name="Email"><value>john@example.com</value></field>
    </fields>
</xfdf>"""
