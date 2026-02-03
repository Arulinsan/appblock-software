@echo off
REM Refresh Icon Cache - Run ini jika icon tidak berubah

echo ========================================
echo Refreshing Windows Icon Cache
echo ========================================
echo.

echo [1/3] Stopping Explorer...
taskkill /F /IM explorer.exe >nul 2>&1

echo [2/3] Clearing icon cache...
cd /d %userprofile%\AppData\Local
if exist IconCache.db del /F /Q IconCache.db
if exist Microsoft\Windows\Explorer\iconcache*.db del /F /Q Microsoft\Windows\Explorer\iconcache*.db

echo [3/3] Restarting Explorer...
start explorer.exe

echo.
echo ✅ Icon cache cleared!
echo.
echo Now check appblock.exe in Explorer.
echo Icon should show correctly.
echo.
echo If still not showing:
echo 1. Right-click appblock.exe → Properties
echo 2. Check if icon shows there
echo 3. Reboot Windows (last resort)
echo.
pause
