VERSION ?= 5.0.0
DOCKERUSERNAME ?= chinmaysawant
# Shared HS256 secret between auth-ms and the backend (override in real deploys).
AUTH_JWT_SECRET ?= dev-insecure-secret-change-me

docker:
	docker build -f dockerfolder/Dockerfile --build-arg VERSION=$(VERSION) -t gopdfsuit:$(VERSION) .
	docker run -d -p 8080:8080 gopdfsuit:$(VERSION)

dockertag:
	docker tag gopdfsuit:$(VERSION) $(DOCKERUSERNAME)/gopdfsuit:$(VERSION)
	docker tag gopdfsuit:$(VERSION) $(DOCKERUSERNAME)/gopdfsuit:latest
	docker login
	docker push $(DOCKERUSERNAME)/gopdfsuit:$(VERSION)
	docker push $(DOCKERUSERNAME)/gopdfsuit:latest

pull:
	docker pull $(DOCKERUSERNAME)/gopdfsuit:$(VERSION)
	docker run -d -p 8080:8080 $(DOCKERUSERNAME)/gopdfsuit:$(VERSION)

build: test-unit
	mkdir -p bin
	go build -o bin/app ./cmd/gopdfsuit

# ============================================================
# Tests (alineados con new_tests/test.md).
#
#   make test-unit          Los 10 unitarios del doc (6 go-be + 4 fe-react):
#                             - 6 go-be   (TC 01,04,06,08,09,10) -> TestFlowCases
#                             - 4 fe-react (TC 02,03,05,07)      -> vitest
#
#   make test-integration   Los 5 de integración:
#                             - TC 04 (E2E Go)     -> TestE2ETemplatePDFFlow
#                             - TC 03 (Playwright) -> merge.fail.spec.js
#                             - TC 05 (Playwright) -> auth.spec.js
#                           (TC 01 y TC 02 del doc no están implementados aún
#                            como specs dedicados.)
#
#   make test               ambos.
# ============================================================

test-unit:
	# 6 unitarios go-be (TC 01, 04, 06, 08, 09, 10).
	go test -count=1 -v -run "TestFlowCases" ./new_tests/backend/
	# 4 unitarios fe-react (TC 02, 03, 05, 07).
	# --ignore-scripts evita el bloqueo de pnpm 11 con esbuild (los tests no
	# necesitan el binario nativo de esbuild).
	cd frontend && corepack pnpm install --ignore-scripts && corepack pnpm run test:isolated

test-integration:
	# Integración Go encadenada (sobre IntegrationSuite real con httptest):
	#   TC 01 -> TestMergePreservesUserOrder (orden de páginas en /merge)
	#   TC 02 -> TestGenerateTemplatePDFThenFillWithXFDF (template-pdf + fill encadenados)
	go test -count=1 -v -run "TestIntegrationSuite/TestMergePreservesUserOrder|TestIntegrationSuite/TestGenerateTemplatePDFThenFillWithXFDF" ./test/
	# Integración Go E2E:
	#   TC 04 -> TestE2ETemplatePDFFlow (servidor real, redirect, SPA, /generate/template-pdf)
	go test -count=1 -v -run TestE2ETemplatePDFFlow ./new_tests/backend/
	# Warm-up para que los `go run` de Playwright no tarden en el primer boot.
	go build -o /dev/null ./cmd/gopdfsuit ./auth-ms
	# Integración Playwright:
	#   TC 03 -> merge.fail.spec.js (FE↔BE merge con PDFs corruptos)
	#   TC 05 -> auth.spec.js       (FE → auth-ms → BE)
	cd new_tests/integration && corepack pnpm install --ignore-scripts && corepack pnpm exec playwright install chromium && corepack pnpm run test:all

# Todo lo nuestro.
test: test-unit test-integration

# Alias.
e2e: test-integration

# auth-ms in a container, SQLite persisted in a named volume (http://localhost:9090).
auth-up:
	AUTH_JWT_SECRET=$(AUTH_JWT_SECRET) docker compose up -d --build auth-ms

auth-down:
	docker compose down

auth-logs:
	docker compose logs -f auth-ms

clean:
	rm -rf bin/

# Launch the whole stack locally with the new auth ENABLED, for manual testing:
#   auth-ms (:9090) + backend (:8080) + frontend dev server (:3000)
# Open http://localhost:3000/gopdfsuit/ — it will require login. Ctrl-C stops all.
run:
	@echo "Starting auth-ms :9090, backend :8080, frontend :3000 (auth ON). Open http://localhost:3000/gopdfsuit/ — Ctrl-C to stop."
	@trap 'kill 0' INT TERM EXIT; \
	AUTH_PORT=9090 AUTH_DB_PATH=auth.db AUTH_JWT_SECRET=$(AUTH_JWT_SECRET) go run ./auth-ms & \
	AUTH_ENABLED=true AUTH_JWT_SECRET=$(AUTH_JWT_SECRET) go run ./cmd/gopdfsuit & \
	(cd frontend && corepack pnpm install --ignore-scripts && VITE_AUTH_ENABLED=true VITE_AUTH_URL=http://localhost:9090 VITE_API_URL=http://localhost:8080 corepack pnpm run dev) & \
	wait

fmt:
	go fmt ./...

vet:
	go vet ./...

mod:
	go mod tidy

lint:
	golangci-lint run -E revive,gocritic,gocyclo,goconst ./...
	cd frontend && corepack pnpm run lint
	cd .. 

gdocker: test-unit
	cd frontend && corepack pnpm install --ignore-scripts && corepack pnpm run build && cd ..
	docker rm -f gopdfsuit
	docker build -t gopdfsuit . 

gdocker-run:
	docker run --rm -p 8080:8080 -d --name gopdfsuit gopdfsuit

gdocker-push:
	export VITE_IS_CLOUD_RUN=true;\
	export VITE_ENVIRONMENT=cloudrun;\
	gcloud builds submit --tag us-east1-docker.pkg.dev/gopdfsuit/gopdfsuit/gopdfsuit-app .	
	gcloud run deploy gopdfsuit-service \
    --image us-east1-docker.pkg.dev/gopdfsuit/gopdfsuit/gopdfsuit-app \
    --region us-east1 \
    --platform managed \
    --allow-unauthenticated \
    --max-instances 1 \
    --concurrency 80 \
    --cpu 1 \
    --memory 512Mi \
	--env-vars-file .env

gengine-deploy: test-unit
	export VITE_IS_CLOUD_RUN=true;\
	export VITE_ENVIRONMENT=cloudrun;\
	export DISABLE_PROFILING=true;\
	cd frontend && corepack pnpm install --ignore-scripts && corepack pnpm run build && cd ..
	gcloud app deploy

.PHONY: build test test-unit test-integration auth-up auth-down auth-logs clean run fmt vet mod lint e2e

# go tool pprof -http=:8081 "http://localhost:8080/debug/pprof/profile?seconds=30"
# go tool pprof -http=:8081 "http://localhost:8080/debug/pprof/heap"