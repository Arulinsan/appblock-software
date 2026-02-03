@echo off
REM Rebuild APPBlock with GUI support

echo =========================================
echo APPBlock - Build Script
echo =========================================
echo.

echo [1/4] Cleaning old build...
if exist appblock.exe (
    taskkill /F /IM appblock.exe >nul 2>&1
    timeout /t 1 /nobreak >nul
    del /F appblock.exe
)

echo [2/4] Generating resources with icon...
rsrc -manifest appblock.manifest -ico icon.ico -o rsrc.syso
if %errorlevel% neq 0 (
    echo.
    echo ❌ Resource generation FAILED!
    echo Make sure rsrc is installed: go install github.com/akavel/rsrc@latest
    pause
    exit /b 1
)

echo [3/4] Building appblock.exe...
go build -ldflags "-s -w -H windowsgui" -o appblock.exe
if %errorlevel% neq 0 (
    echo.
    echo ❌ Build FAILED!
    echo Check errors above.
    pause
    exit /b 1
)

echo [4/4] Checking executable...
if exist appblock.exe (
    for %%A in (appblock.exe) do set size=%%~zA
    echo.
    echo ✅ Build SUCCESS!
    echo    File: appblock.exe
    echo    Size: %size% bytes
    echo.
    echo Ready to run! Double-click appblock.exe or run:
    echo    start appblock.exe
) else (
    echo ❌ Build failed - executable not found
    pause
    exit /b 1
)

echo.
pause
