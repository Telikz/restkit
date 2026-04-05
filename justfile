# Run go mod tidy in all modules
tidy:
    go mod tidy
    (cd adapters/chi && go mod tidy)
    (cd validation/playground && go mod tidy)
    (cd examples/basic && go mod tidy)
    (cd adapters/chi/example/chi && go mod tidy)
    (cd adapters/chi/example/chi_mount && go mod tidy)

# Release tags only
release version:
    VERSION="{{version}}"; VERSION=${VERSION#v}; \
    git tag -a "v$VERSION" -m "Release v$VERSION" && \
    git tag -a "adapters/chi/v$VERSION" -m "Release adapters/chi v$VERSION" && \
    git tag -a "validators/playground/v$VERSION" -m "Release validators/playground v$VERSION" && \
    git tag -a "serializers/yaml/v$VERSION" -m "Release serializers/yaml v$VERSION" && \
    git push origin "v$VERSION" "adapters/chi/v$VERSION" "validators/playground/v$VERSION" "serializers/yaml/v$VERSION" && \
    echo "Released v$VERSION"

fmt:
    go fmt ./... && gofumpt -l -w . && golines -w -m 100 .