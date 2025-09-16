VERSION ?= 1.0.0
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

build:
	go build -o bin/app .

test:
	go test ./...

clean:
	rm -rf bin/

run:
	cd frontend && npm run build && cd ..
	go run cmd/gopdfsuit/main.go

fmt:
	go fmt ./...

vet:
	go vet ./...

mod:
	go mod tidy

.PHONY: build test clean run fmt vet mod

