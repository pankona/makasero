# CLAUDE.md

This file contains the minimum required commands that should pass before pushing changes to the repository.

## Go Backend Commands

Run these commands from the repository root:

```bash
# Download Go dependencies
go mod download

# Check code formatting (should show no output)
gofmt -d .

# Run static analysis
go vet ./...

# Run tests
go test -v ./...

# Run staticcheck (requires staticcheck tool)
staticcheck ./...
```

## Web Frontend Commands

Run these commands from the `web/` directory:

```bash
# Install dependencies
npm install --legacy-peer-deps

# Run linting
npm run lint

# Build the project
npm run build
```

## Installation Requirements

- Go 1.24 or later
- Node.js and npm
- staticcheck tool (`go install honnef.co/go/tools/cmd/staticcheck@latest`)

## Notes

- The `gofmt -d .` command should produce no output if formatting is correct
- All tests should pass with `go test -v ./...`
- The frontend should build successfully with `npm run build`
- Linting should pass with `npm run lint`