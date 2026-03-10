@echo off
setlocal enabledelayedexpansion

rem Build script for gopdfsuit Python bindings shared library on Windows.

set "SCRIPT_DIR=%~dp0"
if "%SCRIPT_DIR:~-1%"=="\" set "SCRIPT_DIR=%SCRIPT_DIR:~0,-1%"

for %%I in ("%SCRIPT_DIR%\..\..") do set "PROJECT_ROOT=%%~fI"
set "OUTPUT_DIR=%SCRIPT_DIR%\pypdfsuit\lib"
set "OUTPUT_FILE=%OUTPUT_DIR%\gopdfsuit.dll"

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

where go >nul 2>nul
if errorlevel 1 (
    echo Go was not found in PATH.
    exit /b 1
)

echo Building shared library for Windows...
echo Output: %OUTPUT_FILE%

pushd "%PROJECT_ROOT%"
if errorlevel 1 exit /b 1

set "CGO_ENABLED=1"
go build -buildmode=c-shared -o "%OUTPUT_FILE%" .\bindings\python\cgo\
if errorlevel 1 (
    popd
    exit /b 1
)

if exist "%OUTPUT_DIR%\gopdfsuit.h" del /f /q "%OUTPUT_DIR%\gopdfsuit.h"

popd

echo Build complete: %OUTPUT_FILE%
endlocal
