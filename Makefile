.PHONY: all clean

# Default target
all: s3peep

# Build s3peep binary - uses local Go if available, otherwise Podman
s3peep:
	@if command -v go >/dev/null 2>&1; then \
		echo "Building with local Go compiler..."; \
		go build -o s3peep ./cmd/s3peep; \
	else \
		echo "Go not found. Building with Podman..."; \
		podman run --rm -v "$$PWD:/app:Z" -w /app docker.io/golang:1.24 \
			sh -c "go mod tidy && go build -o s3peep ./cmd/s3peep"; \
	fi

# Clean build artifacts
clean:
	rm -f s3peep
