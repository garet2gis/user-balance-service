swagger:
	swag fmt && \
	cd ./cmd/main && \
	swag init --pd && \
	cd ../../

build-test:
	docker build -f ./scripts/test.Dockerfile -t go-postgres-test:local .

run-test:
	docker run -e POSTGRES_HOST_AUTH_METHOD=trust -v ${PWD}/cover.out:/testdir/cover.out -e GIT_URL='' go-postgres-test:local

test: build-test run-test

see-cover:
	go tool cover -html=cover.out

.PHONY: swagger build-test run-test see-cover test