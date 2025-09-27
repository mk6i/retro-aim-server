#!/bin/bash

# BART Import Script for Retro AIM Server
# This script imports BART (Buddy ART) assets from an AIM client's bartcache
# directory (usually found under %APPDATA%\acccore\caches\bart) into Retro AIM
# Server via the management API.

set -e

# Default values
API_BASE_URL="http://localhost:8080"
VERBOSE=false
DRY_RUN=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 [OPTIONS] <directory_path>"
    echo ""
    echo "Import BART assets from a bartcache directory into Retro AIM Server"
    echo ""
    echo "Arguments:"
    echo "  directory_path    Path to bartcache directory containing BART assets"
    echo "                   Directory should contain subdirectories named by BART type (0, 1, 2, etc.)"
    echo "                   Each type directory should contain files named by their hash"
    echo ""
    echo "Options:"
    echo "  -u, --url URL     API base URL (default: http://localhost:8080)"
    echo "  -v, --verbose     Enable verbose output"
    echo "  -d, --dry-run     Show what would be uploaded without actually uploading"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Example:"
    echo "  $0 /Users/mike/Downloads/aim\\ barts/eigdbvye/bartcache"
    echo "  $0 --verbose --dry-run /path/to/bartcache"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_verbose() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${BLUE}[VERBOSE]${NC} $1"
    fi
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

check_prerequisites() {
    if ! command_exists curl; then
        log_error "curl is required but not installed"
        exit 1
    fi

    if ! command_exists jq; then
        log_warning "jq is not installed. JSON responses will not be formatted"
    fi
}

test_api() {
    log_info "Testing API connectivity..."

    local response
    if response=$(curl -s -w "%{http_code}" "$API_BASE_URL/bart?type=0" 2>/dev/null); then
        local http_code="${response: -3}"
        if [ "$http_code" = "200" ]; then
            log_success "API is accessible"
            return 0
        else
            log_error "API returned HTTP $http_code"
            return 1
        fi
    else
        log_error "Failed to connect to API at $API_BASE_URL"
        return 1
    fi
}

upload_bart_asset() {
    local file_path="$1"
    local bart_type="$2"
    local hash="$3"

    log_verbose "Uploading $file_path (type: $bart_type, hash: $hash)"

    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] Would upload: $file_path -> type=$bart_type, hash=$hash"
        return 0
    fi

    local response
    local http_code

    # Upload the file
    if response=$(curl -s -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/octet-stream" \
        --data-binary "@$file_path" \
        "$API_BASE_URL/bart/$hash?type=$bart_type" 2>/dev/null); then

        http_code="${response: -3}"
        response_body="${response%???}"

        case "$http_code" in
            201)
                log_success "Uploaded $hash (type: $bart_type)"
                if [ "$VERBOSE" = true ] && command_exists jq; then
                    echo "$response_body" | jq .
                fi
                return 0
                ;;
            409)
                log_warning "Asset $hash already exists (type: $bart_type)"
                return 0
                ;;
            400)
                log_error "Bad request for $hash (type: $bart_type)"
                if command_exists jq; then
                    echo "$response_body" | jq .
                else
                    echo "$response_body"
                fi
                return 1
                ;;
            413)
                log_error "File too large for $hash (type: $bart_type)"
                return 1
                ;;
            *)
                log_error "Upload failed for $hash (type: $bart_type) - HTTP $http_code"
                if command_exists jq; then
                    echo "$response_body" | jq .
                else
                    echo "$response_body"
                fi
                return 1
                ;;
        esac
    else
        log_error "Failed to upload $hash (type: $bart_type)"
        return 1
    fi
}

process_bart_type_directory() {
    local type_dir="$1"
    local bart_type="$2"
    local type_name="$3"

    log_info "Processing BART type $bart_type ($type_name)..."

    if [ ! -d "$type_dir" ]; then
        log_warning "Type directory $type_dir does not exist, skipping"
        return 0
    fi

    local file_count=0
    local success_count=0
    local error_count=0

    # Find all files in the type directory (excluding directories)
    while IFS= read -r -d '' file_path; do
        if [ -f "$file_path" ]; then
            local filename=$(basename "$file_path")
            file_count=$((file_count + 1))

            log_verbose "Found file: $filename"

            if upload_bart_asset "$file_path" "$bart_type" "$filename"; then
                success_count=$((success_count + 1))
            else
                error_count=$((error_count + 1))
            fi
        fi
    done < <(find "$type_dir" -maxdepth 1 -type f -print0 2>/dev/null)

    log_info "Type $bart_type ($type_name): $file_count files, $success_count successful, $error_count errors"

    return $error_count
}

process_directory() {
    local base_dir="$1"

    log_info "Processing BART assets from bartcache directory: $base_dir"

    if [ ! -d "$base_dir" ]; then
        log_error "Directory $base_dir does not exist"
        exit 1
    fi

    # BART type definitions (from the API spec)
    # Using a function to get type name for better shell compatibility
    get_bart_type_name() {
        case "$1" in
            "0") echo "buddy_icon_small" ;;
            "1") echo "buddy_icon" ;;
            "2") echo "status_str" ;;
            "3") echo "arrive_sound" ;;
            "4") echo "rich_text" ;;
            "5") echo "superbuddy_icon" ;;
            "6") echo "radio_station" ;;
            "12") echo "buddy_icon_big" ;;
            "13") echo "status_str_tod" ;;
            "15") echo "current_av_track" ;;
            "96") echo "depart_sound" ;;
            "129") echo "im_chrome" ;;
            "131") echo "im_sound" ;;
            "136") echo "im_chrome_xml" ;;
            "137") echo "im_chrome_immers" ;;
            "1024") echo "emoticon_set" ;;
            "1026") echo "encr_cert_chain" ;;
            "1027") echo "sign_cert_chain" ;;
            "1028") echo "gateway_cert" ;;
            *) echo "unknown_type_$1" ;;
        esac
    }

    local total_errors=0
    local total_files=0
    local total_success=0

    # Process each BART type directory
    # Check for common BART type directories
    local bart_types=("0" "1" "2" "3" "4" "5" "6" "12" "13" "15" "96" "129" "131" "136" "137" "1024" "1026" "1027" "1028")

    for bart_type in "${bart_types[@]}"; do
        local type_dir="$base_dir/$bart_type"
        local type_name=$(get_bart_type_name "$bart_type")

        if [ -d "$type_dir" ]; then
            log_verbose "Found type directory: $type_dir"

            local before_files=$total_files
            local before_success=$total_success
            local before_errors=$total_errors

            if process_bart_type_directory "$type_dir" "$bart_type" "$type_name"; then
                # Count files in this directory
                local type_file_count=$(find "$type_dir" -maxdepth 1 -type f | wc -l)
                total_files=$((total_files + type_file_count))
            else
                total_errors=$((total_errors + 1))
            fi
        else
            log_verbose "Type directory $type_dir not found, skipping"
        fi
    done

    # Summary
    log_info "Import completed!"
    log_info "Total files processed: $total_files"
    log_info "Total errors: $total_errors"

    if [ "$DRY_RUN" = true ]; then
        log_info "This was a dry run - no files were actually uploaded"
    fi

    return $total_errors
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--url)
            API_BASE_URL="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        -*)
            log_error "Unknown option $1"
            usage
            exit 1
            ;;
        *)
            if [ -z "$TARGET_DIR" ]; then
                TARGET_DIR="$1"
            else
                log_error "Multiple directory arguments provided"
                usage
                exit 1
            fi
            shift
            ;;
    esac
done

if [ -z "$TARGET_DIR" ]; then
    log_error "No directory path provided"
    usage
    exit 1
fi

main() {
    log_info "BART Import Script for Retro AIM Server"
    log_info "========================================"

    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY RUN MODE - No files will be uploaded"
    fi

    check_prerequisites

    if [ "$DRY_RUN" = false ]; then
        test_api
    fi

    process_directory "$TARGET_DIR"

    local exit_code=$?
    if [ $exit_code -eq 0 ]; then
        log_success "All operations completed successfully"
    else
        log_error "Some operations failed"
    fi

    exit $exit_code
}

main
