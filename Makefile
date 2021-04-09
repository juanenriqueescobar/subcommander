
sqs:
	docker run -p 9324:9324 -v "$(PWD)/internal/testing/elasticmq:/etc/elasticmq" s12v/elasticmq

build:
	cd internal/di && wire
	go build -o subcommander cmd/main.go
	mv subcommander subcommander-v0.0.9
	sha256sum subcommander-v0.0.9

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

test-race:
	go test -race -timeout 10s -count=1 ./...

# integration_test_clean:
# 	@docker stop subcommander-localsqs > /dev/null 2>&1 || true
# 	@docker rm subcommander-localsqs > /dev/null 2>&1 || true

# integration_test_sqs:
# 	@docker run -d -p 9324:9324 -v "$(PWD)/internal/testing/elasticmq:/etc/elasticmq" --name subcommander-localsqs s12v/elasticmq

# integration_test: integration_test_clean | integration_test_sqs	
# 	@sleep 5
# 	go test -coverprofile cp.out ./...
# 	$(MAKE) integration_test_clean
