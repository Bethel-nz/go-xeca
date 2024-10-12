# directories
FRONTEND_DIR=./web
BACKEND_DIR=.

# build frontend
build-frontend:
	@echo "Building React frontend..."
	cd $(FRONTEND_DIR) && bun run build

# build backend
build-backend:
	@echo "Building Go backend..."
	cd $(BACKEND_DIR) && go build -o app
	
# Clean up build files
clean:
	@echo "Cleaning up build files..."
	cd $(FRONTEND_DIR) && rm -rf build
	cd $(BACKEND_DIR) && rm -rf app

# run both frontend and backend in development
run-dev:
	@echo "Running development environment..."
	cd $(FRONTEND_DIR) && bun run dev & cd $(BACKEND_DIR) && go run .

# full build
build: clean build-frontend build-backend
	@echo "Build completed."

# deploy backend
deploy:
	@echo "Deploying Go backend..."
	cd $(BACKEND_DIR) && ./app

.PHONY: build-frontend build-backend clean run-dev build deploy
