#!/usr/bin/env bash
set -euo pipefail
trap 'rm -f temp_manifest.json temp_script.js' EXIT

# Default values
DEFAULT_REPO_URL="https://github.com/iamadamdev/bypass-paywalls-chrome"
DEFAULT_WS_ADDRESS="ws://127.0.0.1:4343"
DEFAULT_EXTENSION_PATH=""
DESTINATION_FOLDER="external_extension"
INJECT_SCRIPT="extension/src/bg/background.js"
EXTRA_FILES=("extension/src/bg/lame.min.js" "extension/src/bg/RecordRTC.min.js")
ORIGINAL_MANIFEST="extension/manifest.json"  # Path to the original extension's manifest.json

# Function to build target paths
build_target_path() {
    if [ -z "$EXTENSION_PATH" ]; then
        echo "$DESTINATION_FOLDER/$1"
    else
        echo "$DESTINATION_FOLDER/$EXTENSION_PATH/$1"
    fi
}

# Check dependencies
check_dependencies() {
    local missing_deps=()

    for dep in git jq javascript-obfuscator; do
        if ! command -v "$dep" &>/dev/null; then
            missing_deps+=("$dep")
        fi
    done

    if [ ${#missing_deps[@]} -ne 0 ]; then
        echo "Error: Missing dependencies:"
        printf "  - %s\n" "${missing_deps[@]}"
        echo "Please install the missing dependencies before running this script."
        exit 1
    fi
}

# Prompt user for input
prompt_for_input() {
    local prompt="$1"
    local default="$2"
    local user_input

    read -p "$prompt [$default]: " user_input
    echo "${user_input:-$default}"
}

# Clone repository
clone_repository() {
    if [ ! -d "$DESTINATION_FOLDER" ]; then
        echo "Cloning repository..."
        if ! git clone "$REPO_URL" "$DESTINATION_FOLDER"; then
            echo "Error: Failed to clone repository."
            echo "Possible reasons:"
            echo "  - The repository URL may be incorrect."
            echo "  - You may not have an internet connection."
            echo "  - You may not have the necessary permissions."
            echo "Please check the URL and your network connection, then try again."
            exit 1
        fi
    else
        echo "Repository already exists, skipping clone."
    fi
}

# Inject script
inject_script() {
    local source_script="extension/src/bg/background.js"
    local background_script="src/bg/background.js"
    local target_file
    target_file=$(build_target_path "$background_script")

    mkdir -p "$(dirname "$target_file")"

    sed "s|$DEFAULT_WS_ADDRESS|$WS_ADDRESS|g" "$source_script" > temp_script.js

    javascript-obfuscator temp_script.js --output "$target_file" \
        --compact true \
        --self-defending true \
        --disable-console-output true

    rm temp_script.js

    echo "Injected script: $background_script"
}

# Copy extra files
copy_extra_files() {
    for file in "${EXTRA_FILES[@]}"; do
        if [ -f "$file" ]; then
            local filename
            filename=$(basename "$file")
            local target
            target=$(build_target_path "src/bg/$filename")
            mkdir -p "$(dirname "$target")"
            cp "$file" "$target"
            echo "Copied extra file: src/bg/$filename"
        else
            echo "Warning: Extra file not found: $file"
        fi
    done
}

# Update manifest file
update_manifest() {
    local manifest
    manifest=$(build_target_path "manifest.json")
    local temp_manifest="temp_manifest.json"

    # Check manifest version
    local manifest_version
    manifest_version=$(jq -r '.manifest_version' "$manifest")

    if [ "$manifest_version" -eq 2 ]; then
        # Read existing background.scripts
        local existing_bg_scripts
        existing_bg_scripts=$(jq '.background.scripts // []' "$manifest")

        # Merge extra_files and existing_bg_scripts, remove duplicates but keep order
        local merged_scripts
        merged_scripts=$(jq -n --argjson existing "$existing_bg_scripts" --argjson extra '["src/bg/RecordRTC.min.js", "src/bg/lame.min.js"]' '
            ($existing + $extra) | unique')

        # Add src/bg/background.js at the end
        jq --argjson scripts "$merged_scripts" \
           '.background.scripts = ($scripts + ["src/bg/background.js"]) | .background.persistent = true' "$manifest" > "$temp_manifest"

    elif [ "$manifest_version" -eq 3 ]; then
        # Check if service_worker already exists
        local existing_service_worker
        existing_service_worker=$(jq -r '.background.service_worker // empty' "$manifest")

        if [ "$existing_service_worker" != "src/bg/background.js" ]; then
            jq --arg script "src/bg/background.js" \
                '.background.service_worker = $script' "$manifest" > "$temp_manifest"
        else
            cp "$manifest" "$temp_manifest"
        fi
    else
        echo "Error: Unsupported manifest version: $manifest_version"
        exit 1
    fi

    mv "$temp_manifest" "$manifest"

    # Merge permissions from the original extension
    local original_permissions
    original_permissions=$(jq '.permissions // []' "$ORIGINAL_MANIFEST")

    jq --argjson orig_permissions "$original_permissions" \
        '.permissions = (.permissions + $orig_permissions) | .permissions |= unique' "$manifest" > "$temp_manifest"
    mv "$temp_manifest" "$manifest"

    # Merge optional_permissions
    local original_optional_permissions
    original_optional_permissions=$(jq '.optional_permissions // []' "$ORIGINAL_MANIFEST")

    jq --argjson orig_optional_permissions "$original_optional_permissions" \
        '.optional_permissions = (.optional_permissions + $orig_optional_permissions) | .optional_permissions |= unique' "$manifest" > "$temp_manifest"
    mv "$temp_manifest" "$manifest"

    # Merge host_permissions (for Manifest V3)
    if [ "$manifest_version" -eq 3 ]; then
        local original_host_permissions
        original_host_permissions=$(jq '.host_permissions // []' "$ORIGINAL_MANIFEST")

        jq --argjson orig_host_permissions "$original_host_permissions" \
            '.host_permissions = (.host_permissions + $orig_host_permissions) | .host_permissions |= unique' "$manifest" > "$temp_manifest"
        mv "$temp_manifest" "$manifest"
    fi

    # Merge content_scripts without duplicates
    local original_content_scripts
    original_content_scripts=$(jq '.content_scripts // []' "$ORIGINAL_MANIFEST")

    jq --argjson orig_content_scripts "$original_content_scripts" \
        '.content_scripts = (.content_scripts + $orig_content_scripts)' "$manifest" > "$temp_manifest"
    mv "$temp_manifest" "$manifest"

    # Merge web_accessible_resources without duplicates
    local original_war
    original_war=$(jq '.web_accessible_resources // []' "$ORIGINAL_MANIFEST")

    jq --argjson orig_war "$original_war" \
        '.web_accessible_resources = (.web_accessible_resources + $orig_war) | .web_accessible_resources |= unique' "$manifest" > "$temp_manifest"
    mv "$temp_manifest" "$manifest"

    echo "Updated manifest.json with merged permissions and resources"
}

# Create CRX file
create_crx() {
    local crx_file="${DESTINATION_FOLDER}-injected.crx"
    local pem_file="${DESTINATION_FOLDER}-injected.pem"
    local extension_dir
    extension_dir=$(build_target_path "")

    # Detect browser command
    local browser_cmd
    if command -v chromium &>/dev/null; then
        browser_cmd="chromium"
    elif command -v chromium-browser &>/dev/null; then
        browser_cmd="chromium-browser"
    elif command -v google-chrome &>/dev/null; then
        browser_cmd="google-chrome"
    else
        echo "Error: Chromium or Chrome is not installed, cannot create CRX file."
        return 1
    fi

    # Generate new PEM file if it doesn't exist
    if [ ! -f "$pem_file" ]; then
        openssl genrsa -out "$pem_file" 2048
    fi

    # Pack the extension
    "$browser_cmd" --pack-extension="$extension_dir" --pack-extension-key="$pem_file"

    # Move the generated CRX file
    local generated_crx="${extension_dir}.crx"
    if [ -f "$generated_crx" ]; then
        mv "$generated_crx" "$crx_file"
        echo "Created CRX file: $crx_file"
    else
        echo "Error: Failed to create CRX file."
    fi
}

# Main execution flow
check_dependencies

REPO_URL=$(prompt_for_input "Enter the repository URL" "$DEFAULT_REPO_URL")
WS_ADDRESS=$(prompt_for_input "Enter the WebSocket address" "$DEFAULT_WS_ADDRESS")
EXTENSION_PATH=$(prompt_for_input "Enter the extension path (leave empty if in root)" "$DEFAULT_EXTENSION_PATH")

# Validate user input
if ! [[ $REPO_URL =~ ^https?:// ]]; then
    echo "Error: Invalid repository URL."
    exit 1
fi

if ! [[ $WS_ADDRESS =~ ^ws:// ]]; then
    echo "Error: Invalid WebSocket address."
    exit 1
fi

echo "Configuration:"
echo "  Repository URL: $REPO_URL"
echo "  WebSocket Address: $WS_ADDRESS"
echo "  Extension Path: ${EXTENSION_PATH:-Root directory}"
echo

read -p "Proceed with this configuration? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Operation cancelled."
    exit 1
fi

clone_repository
inject_script
copy_extra_files
update_manifest

read -p "Do you want to create a CRX file? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    create_crx
fi

echo "Script execution completed."
