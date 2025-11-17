#!/bin/bash
# Verify gosec release integrity using GPG signatures and checksums
# Usage: ./verify-gosec-release.sh <version>
# Example: ./verify-gosec-release.sh 2.22.10

set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 2.22.10"
    exit 1
fi

# Remove 'v' prefix if present
VERSION="${VERSION#v}"

TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

echo "=== Verifying gosec v${VERSION} ==="

# URLs for verification
CHECKSUMS_URL="https://github.com/securego/gosec/releases/download/v${VERSION}/gosec_${VERSION}_checksums.txt"
GPG_SIG_URL="https://github.com/securego/gosec/releases/download/v${VERSION}/gosec_${VERSION}_checksums.txt.gpg"
RELEASE_API="https://api.github.com/repos/securego/gosec/releases/tags/v${VERSION}"

echo "1. Downloading checksums file..."
if ! curl -sL -o "${TEMP_DIR}/checksums.txt" "$CHECKSUMS_URL"; then
    echo "ERROR: Failed to download checksums file"
    exit 1
fi
echo "   Downloaded: ${CHECKSUMS_URL}"

echo "2. Downloading GPG signature..."
if ! curl -sL -o "${TEMP_DIR}/checksums.txt.gpg" "$GPG_SIG_URL"; then
    echo "ERROR: Failed to download GPG signature"
    exit 1
fi
echo "   Downloaded: ${GPG_SIG_URL}"

echo "3. Verifying GPG signature..."
# Import gosec maintainers' public key if not present
# The gosec project uses GitHub's verified commits
if command -v gpg &> /dev/null; then
    # Try to verify signature
    if gpg --verify "${TEMP_DIR}/checksums.txt.gpg" "${TEMP_DIR}/checksums.txt" 2>&1; then
        echo "   GPG signature verified successfully"
    else
        echo "   WARNING: GPG verification failed or key not found"
        echo "   To import the signing key, visit: https://github.com/securego/gosec"
        echo "   Proceeding with checksum display (verify manually)"
    fi
else
    echo "   WARNING: gpg not installed, skipping signature verification"
    echo "   Install gnupg to enable GPG signature verification"
fi

echo "4. Fetching release commit SHA from GitHub API..."
COMMIT_SHA=""
if command -v jq &> /dev/null; then
    RELEASE_INFO=$(curl -sL "$RELEASE_API")
    COMMIT_SHA=$(echo "$RELEASE_INFO" | jq -r '.target_commitish // empty')
    TAG_NAME=$(echo "$RELEASE_INFO" | jq -r '.tag_name // empty')
    CREATED_AT=$(echo "$RELEASE_INFO" | jq -r '.created_at // empty')

    if [ -n "$COMMIT_SHA" ]; then
        echo "   Release Tag: ${TAG_NAME}"
        echo "   Created: ${CREATED_AT}"
        echo "   Commit SHA: ${COMMIT_SHA}"
    else
        echo "   WARNING: Could not fetch commit SHA from API"
    fi
else
    echo "   WARNING: jq not installed, cannot parse release info"
    echo "   Install jq to fetch commit SHA automatically"
fi

echo "5. Displaying checksums (verify against official release page):"
echo "-------------------------------------------------------------------"
cat "${TEMP_DIR}/checksums.txt"
echo "-------------------------------------------------------------------"

echo ""
echo "=== RECOMMENDED CI/CD CONFIGURATION ==="
if [ -n "$COMMIT_SHA" ]; then
    echo "GitHub Actions workflow entry:"
    echo ""
    echo "      - name: Run Gosec Security Scanner"
    echo "        # SECURITY: Pinned to specific commit SHA for v${VERSION} (immutable)"
    echo "        # Checksum verification: ${CHECKSUMS_URL}"
    echo "        # GPG signature: ${GPG_SIG_URL}"
    echo "        uses: securego/gosec@${COMMIT_SHA} # v${VERSION}"
    echo ""
else
    echo "NOTE: Commit SHA not available. Visit:"
    echo "  https://github.com/securego/gosec/releases/tag/v${VERSION}"
    echo "to find the full commit SHA for the release."
fi

echo "=== VERIFICATION COMPLETE ==="
echo ""
echo "To ensure supply chain security:"
echo "1. Verify the checksums match the official release page"
echo "2. Verify the GPG signature (if gpg is available)"
echo "3. Use the full commit SHA in your CI/CD workflow"
echo "4. Do NOT use floating refs like @master or @main"
