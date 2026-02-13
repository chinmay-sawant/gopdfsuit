VERSION ?= 2.0.0
DOCKERUSERNAME ?= chinmaysawant

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

build: test-integration
	mkdir -p bin
	go build -o bin/app ./cmd/gopdfsuit

test:
	go test ./...
	python3 -m pytest bindings/python/tests

test-integration:
	go test -count=1 -v ./test

clean:
	rm -rf bin/

run: test-integration lint
	export VITE_IS_CLOUD_RUN=false;\
	export VITE_ENVIRONMENT=local;\
	export VITE_API_URL=http://localhost:8080;\
	cd frontend && npm run build && cd ..
	go run cmd/gopdfsuit/main.go

fmt:
	go fmt ./...

vet:
	go vet ./...

mod:
	go mod tidy

lint:
	golangci-lint run ./...
	cd frontend && npm run lint
	cd .. 

gdocker: test-integration
	cd frontend && npm run build && cd ..
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

gengine-deploy: test-integration
	export VITE_IS_CLOUD_RUN=true;\
	export VITE_ENVIRONMENT=cloudrun;\
	export DISABLE_PROFILING=true;\
	cd frontend && npm run build && cd ..
	gcloud app deploy

.PHONY: build test clean run fmt vet mod lint

# go tool pprof -http=:8081 "http://localhost:8080/debug/pprof/profile?seconds=30"
# go tool pprof -http=:8081 "http://localhost:8080/debug/pprof/heap"