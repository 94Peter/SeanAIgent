IMAGE_TAG := latest
IMAGE_NAME := seanaigent

# Run templ generation in watch mode
templ:
	templ generate --watch --proxy="http://localhost:8082" --open-browser=false

# # Run air for Go hot reload
server:
	air \
	--build.cmd "go build -o tmp/main ./main.go" \
	--build.bin "tmp/main" \
	--build.args_bin "serve console"
	--build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# Watch Tailwind CSS changes
tailwind:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --watch

# Start development server with all watchers
dev:
	make -j3 tailwind templ server





build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .


multi-build:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t 94peter/$(IMAGE_NAME):latest \
		--push .

wrk-test:
	wrk -t8 -c8 -d30s --latency http://localhost:8080/training/booking


wire:
	go generate ./...