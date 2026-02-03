@echo off
REM Rebuild APPBlock with GUI support

echo =========================================
echo APPBlock - Build Script
echo =========================================
echo.

echo [1/3] Cleaning old build...
if exist appblock.exe (
    taskkill /F /IM appblock.exe >nul 2>&1
    timeout /t 1 /nobreak >nul
    del /F appblock.exe
)

echo [2/3] Building appblock.exe...
go build -ldflags "-s -w -H windowsgui" -o appblock.exe
if %errorlevel% neq 0 (
    echo.
    echo ❌ Build FAILED!
    echo Check errors above.
    pause
    exit /b 1
)

echo [3/3] Checking executable...
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
