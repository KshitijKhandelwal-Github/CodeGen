mkdir -p scripts
cat > scripts/build-release.sh << 'EOF'
#!/bin/bash

# Build binaries for all platforms
VERSION=${1:-v0.1.0}

echo "Building CodeGen Pro $VERSION for all platforms..."
echo ""

# Clean
rm -rf dist
mkdir -p dist

# Build for each platform
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION" -o dist/codegen-macos-amd64 ./cmd/cli
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=$VERSION" -o dist/codegen-macos-arm64 ./cmd/cli
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION" -o dist/codegen-linux-amd64 ./cmd/cli
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=$VERSION" -o dist/codegen-linux-arm64 ./cmd/cli
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$VERSION" -o dist/codegen-windows-amd64.exe ./cmd/cli

echo "✓ Built binaries:"
ls -lh dist/
echo ""

# Create checksums
cd dist
shasum -a 256 * > checksums.txt
cd ..

echo "✓ Created checksums"
echo ""
echo "Release files ready in dist/"
EOF

chmod +x scripts/build-release.sh