#!/bin/sh
# Patches openapi/openapi.yaml servers with APP_URL from the root .env before starting Swagger UI.
set -eu

APP_URL="${APP_URL:-http://localhost:8000}"
APP_URL="${APP_URL%/}"

HTML_DIR="/usr/share/nginx/html"
SRC_DIR="/openapi-src"

mkdir -p "${HTML_DIR}/openapi"
cp "${SRC_DIR}/swagger-initializer.js" "${HTML_DIR}/swagger-initializer.js"

if [ "$APP_URL" != "http://localhost:8000" ]; then
	awk -v app_url="$APP_URL" '
		/^servers:/ {
			print
			print "    - description: API Gateway"
			print "      url: " app_url
			in_servers = 1
			next
		}
		in_servers && /^[^ \t]/ {
			in_servers = 0
		}
		in_servers {
			next
		}
		{ print }
	' "${SRC_DIR}/openapi.yaml" > "${HTML_DIR}/openapi/openapi.yaml"
else
	cp "${SRC_DIR}/openapi.yaml" "${HTML_DIR}/openapi/openapi.yaml"
fi

exec /docker-entrypoint.sh nginx -g "daemon off;"
