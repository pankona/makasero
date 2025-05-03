.PHONY: install

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
	cp -r web/.next/static ~/.makasero/web-frontend/
	cp -r web/public/* ~/.makasero/web-frontend/ 2>/dev/null || true
	
	@echo "Installation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf web/.next
	rm -rf ~/.makasero/web-frontend/*
	
	@echo "Clean complete!"
