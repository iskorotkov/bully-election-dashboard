version = v1.1.1
image = iskorotkov/bully-election-dashboard
namespace = bully-election-dashboard

.PHONY: ci
ci: build test build-image push-image deploy

.PHONY: build
build:
	go build ./...

.PHONY: run
run:
	go run ./...

.PHONY: test
test:
	go test ./...

.PHONY: test-short
test-short:
	go test ./... -short

.PHONY: build-image
build-image:
	docker build -t $(image):$(version) -f build/bully-election-dashboard.dockerfile .

.PHONY: push-image
push-image:
	docker push $(image):$(version)

.PHONY: deploy
deploy:
	kubectl apply -f deploy/bully-election-dashboard.yml -n $(namespace)

.PHONY: undeploy
undeploy:
	kubectl delete -f deploy/bully-election-dashboard.yml -n $(namespace)
