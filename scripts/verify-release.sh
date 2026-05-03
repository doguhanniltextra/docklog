#!/bin/bash
# verify-release.sh
# Advanced Stability & Portability Audit for Docklog (Linux/WSL)

set -e

# Colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "\n${CYAN}--- [Docklog Stability Audit: Linux/WSL] ---${NC}"

# 1. Resolve Environment
echo -n "[1/5] Checking environment..."
GOPATH_VAL=$(go env GOPATH)
if [ -z "$GOPATH_VAL" ]; then
    echo -e "${RED} FAILED${NC}"
    exit 1
fi
echo -e "${GREEN} DONE${NC}"

# 2. Build Audit (Static Linking)
echo -n "[2/5] Building static binary (CGO_ENABLED=0)..."
# We build a temporary binary for the portability test
if CGO_ENABLED=0 go build -trimpath -o ./dist/docklog-audit-linux-amd64 .; then
    echo -e "${GREEN} DONE${NC}"
else
    echo -e "${RED} FAILED${NC}"
    exit 1
fi

# 3. Cross-Compile Verification
echo -n "[3/5] Verifying multi-architecture builds..."
# Testing if we can build for ARM and Windows without errors
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o /dev/null .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /dev/null .
echo -e "${GREEN} DONE${NC}"

# 4. Portability Test (Foreign OS Simulation)
echo -e "${YELLOW}[4/5] Running Portability Test (Alpine Linux)...${NC}"
# This runs the compiled binary inside a minimal Alpine container.
# If it runs there, it is truly statically linked and portable.
if docker run --rm \
    -v $(pwd)/dist/docklog-audit-linux-amd64:/usr/local/bin/docklog \
    alpine:latest docklog version; then
    echo -e "${GREEN}    SUCCESS:${NC} Binary is portable and runs on Alpine."
else
    echo -e "${RED}    FAILED:${NC} Binary is not portable or has library dependencies."
    exit 1
fi

# 5. Final Installation & Local Verify
echo -n "[5/5] Performing final local installation..."
pkill docklog || true
go install .
echo -e "${GREEN} DONE${NC}"

echo -e "\n${CYAN}Installed Version:${NC}"
$(go env GOPATH)/bin/docklog version

echo -e "\n${GREEN}✔ RELEASE STABILITY AUDIT PASSED${NC}"
echo -e "This version is safe to publish.\n"
