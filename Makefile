.PHONY: install install-claude-code

# Default target
all: install

# Install makasero-web-backend and build frontend
install:
	@echo "Installing makasero-web-backend..."
	go install ./cmd/makasero-web-backend
	
	@echo "Building web frontend..."
	cd web && npm install --legacy-peer-deps && npm run build
	
	@echo "Creating frontend directory if it doesn't exist..."
	mkdir -p ~/.makasero/web-frontend
	
	@echo "Copying frontend build to ~/.makasero/web-frontend..."
	mkdir -p ~/.makasero/web-frontend/_next
	cp -r web/.next/static ~/.makasero/web-frontend/_next/
	cp -r web/.next/server/app/*.html ~/.makasero/web-frontend/ 2>/dev/null || true
	cp -r web/public/* ~/.makasero/web-frontend/ 2>/dev/null || true
	
	@echo "Installation complete!"

# Install Claude Code CLI
install-claude-code:
	@echo "Installing Claude Code CLI..."
	npm install -g @anthropic-ai/claude-code
	@echo "Claude Code CLI installation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf web/.next
	rm -rf ~/.makasero/web-frontend/*
	
	@echo "Clean complete!"
