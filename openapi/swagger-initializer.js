// Swagger UI initializer — works behind Kong (/) and direct (:8081).
window.onload = function () {
  window.ui = SwaggerUIBundle({
    url: "/openapi/openapi.yaml",
    dom_id: "#swagger-ui",
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "StandaloneLayout",
    persistAuthorization: true,
    tryItOutEnabled: true,
    validatorUrl: null,
  });
};
