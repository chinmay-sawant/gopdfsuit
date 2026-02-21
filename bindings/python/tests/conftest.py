"""
Pytest configuration and fixtures for pypdfsuit tests.
"""

import os
import subprocess
from pathlib import Path

import pytest


def _should_rebuild(lib_path: Path, source_roots: list[Path]) -> bool:
    if not lib_path.exists():
        return True

    lib_mtime = lib_path.stat().st_mtime
    for root in source_roots:
        if not root.exists():
            continue
        for path in root.rglob("*.go"):
            if path.stat().st_mtime > lib_mtime:
                return True
    return False


def pytest_sessionstart(session):
    """Ensure tests run against a freshly built shared library."""
    if os.getenv("PYPDFSUIT_SKIP_AUTO_BUILD") == "1":
        return

    python_dir = Path(__file__).resolve().parents[1]
    repo_root = python_dir.parents[1]
    lib_name = "libgopdfsuit.so"
    lib_path = python_dir / "pypdfsuit" / "lib" / lib_name
    source_roots = [
        repo_root / "bindings" / "python" / "cgo",
        repo_root / "pkg" / "gopdflib",
        repo_root / "internal" / "pdf",
    ]

    if not _should_rebuild(lib_path, source_roots):
        return

    build_script = python_dir / "build.sh"
    subprocess.run([str(build_script)], check=True, cwd=str(python_dir))


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
