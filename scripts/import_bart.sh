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
BART_TYPE=""
TARGET_DIRS=()

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 [OPTIONS] -t <type> <directory_path> [directory_path...]"
    echo ""
    echo "Import BART assets from bartcache directories into Retro AIM Server"
    echo ""
    echo "Arguments:"
    echo "  directory_path    Path to bartcache directory containing BART assets"
    echo "                   Directory should contain subdirectories named by BART type (0, 1, 2, etc.)"
    echo "                   Each type directory should contain files named by their hash"
    echo "                   Multiple directories can be specified for bulk import"
    echo ""
    echo "Options:"
    echo "  -t, --type TYPE   BART type to import (required)"
    echo "                   Valid types: buddy_icon_small, buddy_icon, status_str, arrive_sound,"
    echo "                   rich_text, superbuddy_icon, radio_station, buddy_icon_big,"
    echo "                   status_str_tod, current_av_track, depart_sound, im_chrome,"
    echo "                   im_sound, im_chrome_xml, im_chrome_immers, emoticon_set,"
    echo "                   encr_cert_chain, sign_cert_chain, gateway_cert"
    echo "  -u, --url URL     API base URL (default: http://localhost:8080)"
    echo "  -v, --verbose     Enable verbose output"
    echo "  -d, --dry-run     Show what would be uploaded without actually uploading"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -t buddy_icon /Users/mike/Downloads/aim\\ barts/eigdbvye/bartcache"
    echo "  $0 --type status_str --verbose --dry-run /path/to/bartcache"
    echo "  $0 -t arrive_sound /path/to/bartcache/*"
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

get_bart_type_number() {
    case "$1" in
        "buddy_icon_small") echo "0" ;;
        "buddy_icon") echo "1" ;;
        "status_str") echo "2" ;;
        "arrive_sound") echo "3" ;;
        "rich_text") echo "4" ;;
        "superbuddy_icon") echo "5" ;;
        "radio_station") echo "6" ;;
        "buddy_icon_big") echo "12" ;;
        "status_str_tod") echo "13" ;;
        "current_av_track") echo "15" ;;
        "depart_sound") echo "96" ;;
        "im_chrome") echo "129" ;;
        "im_sound") echo "131" ;;
        "im_chrome_xml") echo "136" ;;
        "im_chrome_immers") echo "137" ;;
        "emoticon_set") echo "1024" ;;
        "encr_cert_chain") echo "1026" ;;
        "sign_cert_chain") echo "1027" ;;
        "gateway_cert") echo "1028" ;;
        *) echo "" ;;
    esac
}

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
                log_success "Uploaded $hash"
                if [ "$VERBOSE" = true ] && command_exists jq; then
                    echo "$response_body" | jq .
                fi
                return 0
                ;;
            409)
                log_warning "Asset $hash already exists"
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

    if [ ! -d "$base_dir" ] && [ ! -f "$base_dir" ]; then
        log_error "Path $base_dir does not exist"
        return 1
    fi

    # Check if this is a single file
    if [ -f "$base_dir" ]; then
        log_info "Processing file: $base_dir"
        local filename=$(basename "$base_dir")
        if upload_bart_asset "$base_dir" "$BART_TYPE_NUMBER" "$filename"; then
            log_success "Successfully processed file: $base_dir"
            return 0
        else
            log_error "Failed to process file: $base_dir"
            return 1
        fi
    fi

    # For directories, show the BART type info
    log_info "Processing BART assets from directory: $base_dir"
    log_info "BART type: $BART_TYPE (type number: $BART_TYPE_NUMBER)"

    # Look for the specific type directory first (nested structure)
    local type_dir="$base_dir/$BART_TYPE_NUMBER"

    if [ -d "$type_dir" ]; then
        log_info "Found type directory: $type_dir"
        # Process the single type directory
        if process_bart_type_directory "$type_dir" "$BART_TYPE_NUMBER" "$BART_TYPE"; then
            local file_count=$(find "$type_dir" -maxdepth 1 -type f | wc -l)
            log_success "Successfully processed $file_count files for type $BART_TYPE"
            return 0
        else
            log_error "Failed to process files for type $BART_TYPE"
            return 1
        fi
    fi

    # If no type directory found, check if this is a flat structure (files directly in directory)
    local file_count=$(find "$base_dir" -maxdepth 1 -type f | wc -l)
    if [ $file_count -gt 0 ]; then
        log_info "Found flat file structure with $file_count files"
        # Process files directly in the directory
        if process_bart_type_directory "$base_dir" "$BART_TYPE_NUMBER" "$BART_TYPE"; then
            log_success "Successfully processed $file_count files for type $BART_TYPE"
            return 0
        else
            log_error "Failed to process files for type $BART_TYPE"
            return 1
        fi
    fi

    # No files or directories found
    log_error "No files found in $base_dir"
    log_error "Expected either:"
    log_error "  - Nested structure: $base_dir/$BART_TYPE_NUMBER/"
    log_error "  - Flat structure: files directly in $base_dir/"
    return 1
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            BART_TYPE="$2"
            shift 2
            ;;
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
            TARGET_DIRS+=("$1")
            shift
            ;;
    esac
done

if [ -z "$BART_TYPE" ]; then
    log_error "BART type is required"
    usage
    exit 1
fi

if [ ${#TARGET_DIRS[@]} -eq 0 ]; then
    log_error "No directory path provided"
    usage
    exit 1
fi

# Validate BART type
BART_TYPE_NUMBER=$(get_bart_type_number "$BART_TYPE")
if [ -z "$BART_TYPE_NUMBER" ]; then
    log_error "Invalid BART type: $BART_TYPE"
    log_error "Valid types: buddy_icon_small, buddy_icon, status_str, arrive_sound, rich_text, superbuddy_icon, radio_station, buddy_icon_big, status_str_tod, current_av_track, depart_sound, im_chrome, im_sound, im_chrome_xml, im_chrome_immers, emoticon_set, encr_cert_chain, sign_cert_chain, gateway_cert"
    exit 1
fi

main() {
    log_info "BART Import Script for Retro AIM Server"
    log_info "========================================"
    log_info "Target paths: ${TARGET_DIRS[*]}"
    log_info "BART type: $BART_TYPE (type number: $BART_TYPE_NUMBER)"

    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY RUN MODE - No files will be uploaded"
    fi

    check_prerequisites

    if [ "$DRY_RUN" = false ]; then
        test_api
    fi

    local total_errors=0
    local total_dirs=0

    # Process each file/directory
    for target_path in "${TARGET_DIRS[@]}"; do
        total_dirs=$((total_dirs + 1))

        if process_directory "$target_path"; then
            # Success is logged by process_directory
            :
        else
            total_errors=$((total_errors + 1))
        fi
    done

    # Summary
    log_info "Import completed!"
    log_info "Total files processed: $total_dirs"
    log_info "Total errors: $total_errors"

    if [ $total_errors -eq 0 ]; then
        log_success "All operations completed successfully"
        exit 0
    else
        log_error "Some operations failed"
        exit 1
    fi
}

main
