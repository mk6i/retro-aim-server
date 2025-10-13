#!/bin/bash

# BART Import Script for Retro AIM Server
# This script imports BART (Buddy ART) asset files into Retro AIM Server
# via the management API.
# 
# Compatible with macOS and Linux terminals

set -e

# Ensure we're using bash and have proper error handling
if [ -z "$BASH_VERSION" ]; then
    echo "Error: This script requires bash" >&2
    exit 1
fi

# Default values
API_BASE_URL="http://localhost:8080"
VERBOSE=false
DRY_RUN=false
BART_TYPE=""
TARGET_FILES=()

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 [OPTIONS] -t <type> <file_path> [file_path...]"
    echo ""
    echo "Import BART assets into Retro AIM Server"
    echo ""
    echo "Arguments:"
    echo "  file_path         Path to BART asset file(s) to import"
    echo "                   Files should be named by their hash (hexadecimal)"
    echo "                   Multiple files can be specified for bulk import"
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
    echo "  $0 -t buddy_icon /path/to/bart/abc123def456"
    echo "  $0 --type status_str --verbose --dry-run /path/to/file1 /path/to/file2"
    echo "  $0 -t arrive_sound /path/to/files/*"
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

is_hex_string() {
    local string="$1"
    # Check if string contains only hexadecimal characters (0-9, a-f, A-F)
    if [[ "$string" =~ ^[0-9a-fA-F]+$ ]]; then
        return 0
    else
        return 1
    fi
}

# Normalize path for cross-platform compatibility
normalize_path() {
    local path="$1"
    # Remove trailing slashes and normalize the path
    echo "$path" | sed 's|/*$||'
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
}

test_api() {
    log_info "Testing API connectivity..."

    local response
    local http_code
    if response=$(curl -s -w "%{http_code}" "$API_BASE_URL/bart?type=0" 2>/dev/null); then
        # Extract HTTP code from response (last 3 characters)
        http_code=$(echo "$response" | tail -c 4)
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

        # Extract HTTP code from response (last 3 characters)
        http_code=$(echo "$response" | tail -c 4)
        # Extract response body (all except last 3 characters)
        # Use a more portable approach that works on both macOS and Linux
        response_length=$(echo "$response" | wc -c)
        response_body_length=$((response_length - 4))
        response_body=$(echo "$response" | head -c "$response_body_length")

        case "$http_code" in
            201)
                log_success "Uploaded $hash"
                if [ "$VERBOSE" = true ]; then
                    echo "$response_body"
                fi
                return 0
                ;;
            409)
                log_warning "Asset $hash already exists"
                return 0
                ;;
            400)
                log_error "Bad request for $hash (type: $bart_type)"
                echo "$response_body"
                return 1
                ;;
            413)
                log_error "File too large for $hash (type: $bart_type)"
                return 1
                ;;
            *)
                log_error "Upload failed for $hash (type: $bart_type) - HTTP $http_code"
                echo "$response_body"
                return 1
                ;;
        esac
    else
        log_error "Failed to upload $hash (type: $bart_type)"
        return 1
    fi
}

process_file() {
    local file_path="$1"
    file_path=$(normalize_path "$file_path")

    if [ ! -f "$file_path" ]; then
        log_error "File $file_path does not exist"
        return 1
    fi

    log_info "Processing file: $file_path"
    local filename=$(basename "$file_path")
    
    # Validate that filename is a hexadecimal string
    if ! is_hex_string "$filename"; then
        log_error "Cannot process file '$filename' - filename is not a valid hexadecimal string"
        return 1
    fi
    
    if upload_bart_asset "$file_path" "$BART_TYPE_NUMBER" "$filename"; then
        log_success "Successfully processed file: $file_path"
        return 0
    else
        log_error "Failed to process file: $file_path"
        return 1
    fi
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
            TARGET_FILES+=("$1")
            shift
            ;;
    esac
done

if [ -z "$BART_TYPE" ]; then
    log_error "BART type is required"
    usage
    exit 1
fi

if [ ${#TARGET_FILES[@]} -eq 0 ]; then
    log_error "No file path provided"
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
    log_info "Target files: ${TARGET_FILES[*]}"
    log_info "BART type: $BART_TYPE (type number: $BART_TYPE_NUMBER)"

    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY RUN MODE - No files will be uploaded"
    fi

    check_prerequisites

    if [ "$DRY_RUN" = false ]; then
        test_api
    fi

    local total_errors=0

    # Process each file
    # Ensure array is properly expanded for cross-platform compatibility
    if [ ${#TARGET_FILES[@]} -gt 0 ]; then
        for target_path in "${TARGET_FILES[@]}"; do
            if process_file "$target_path"; then
                # Success is logged by process_file
                :
            else
                total_errors=$((total_errors + 1))
            fi
        done
    fi

    # Summary
    log_info "Import completed!"
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
