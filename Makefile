swagger:
	swag fmt && \
	cd ./cmd/main && \
	swag init --pd && \
	cd ../../

.PHONY: swagger