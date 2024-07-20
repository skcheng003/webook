# .PHONY: docker
# docker:
#	@rm webook || true
#	@echo "Clean up."
#	@go mod tidy
#	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
#	@docker rmi -f senkie/webook:v0.0.1
#	@docker build -t senkie/webook:v0.0.1 .
#	@echo "Build image finish."

.PHONY: mock
mock:
	@echo "Generate mock files."
	@mockgen -source=internal/service/user.go -package=svcmocks -destination=internal/service/mocks/user.mock.gen.go
	@mockgen -source=internal/service/code.go -package=svcmocks -destination=internal/service/mocks/code.mock.gen.go
	@go mod tidy