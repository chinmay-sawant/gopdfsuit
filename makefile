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

build:
	mkdir -p bin
	go build -o bin/app ./cmd/gopdfsuit

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

gdocker:
	docker rm -f gopdfsuit
	docker build -t gopdfsuit . 

gdocker-run:
	docker run --rm -p 8080:8080 -d --name gopdfsuit gopdfsuit
	
.PHONY: build test clean run fmt vet mod

