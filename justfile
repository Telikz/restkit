# Run go mod tidy in all modules
tidy:
    go mod tidy
    (cd adapters/chi && go mod tidy)
    (cd validation/playground && go mod tidy)
    (cd examples/basic && go mod tidy)
    (cd examples/websocket && go mod tidy)
    (cd adapters/chi/example/chi && go mod tidy)
    (cd adapters/chi/example/chi_mount && go mod tidy)

# Release tags only
release version arguments="":
    VERSION="{{version}}"; VERSION=${VERSION#v}; \
    git tag {{arguments}} -a "v$VERSION" -m "Release v$VERSION" && \
    git tag {{arguments}} -a "adapters/chi/v$VERSION" -m "Release adapters/chi v$VERSION" && \
    git tag {{arguments}} -a "adapters/echo/v$VERSION" -m "Release adapters/echo v$VERSION" && \
    git tag {{arguments}} -a "adapters/gin/v$VERSION" -m "Release adapters/gin v$VERSION" && \
    git tag {{arguments}} -a "adapters/stdlib/v$VERSION" -m "Release adapters/stdlib v$VERSION" && \
    git tag {{arguments}} -a "validators/playground/v$VERSION" -m "Release validators/playground v$VERSION" && \
    git tag {{arguments}} -a "serializers/yaml/v$VERSION" -m "Release serializers/yaml v$VERSION" && \
    git tag {{arguments}} -a "extra/http3/v$VERSION" -m "Release extra/http3 v$VERSION" && \
    git tag {{arguments}} -a "extra/websocket/v$VERSION" -m "Release extra/websocket v$VERSION" && \
    git tag {{arguments}} -a "extra/grpc/v$VERSION" -m "Release extra/grpc v$VERSION" && \
    git push {{arguments}} origin "v$VERSION" "adapters/chi/v$VERSION" "adapters/echo/v$VERSION" "adapters/gin/v$VERSION" "adapters/stdlib/v$VERSION" "validators/playground/v$VERSION" "serializers/yaml/v$VERSION" "extra/http3/v$VERSION" "extra/websocket/v$VERSION" && \
    echo "Released v$VERSION"

fmt:
    go fmt ./... && gofumpt -l -w . && golines -w -m 100 .