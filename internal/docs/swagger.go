package docs

import (
	"net/http"
)

// ServeSwaggerUI serves the Swagger UI HTML
func ServeSwaggerUI(w http.ResponseWriter, swaggerUIPath string) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Swagger UI</title>
	<meta charset="utf-8"/>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
	<script>
		const ui = SwaggerUIBundle({
			url: "` + swaggerUIPath + `/openapi.json",
			dom_id: '#swagger-ui',
			presets: [
				SwaggerUIBundle.presets.apis,
				SwaggerUIBundle.SwaggerUIStandalonePreset
			],
			layout: "BaseLayout"
		})
	</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
