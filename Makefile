build-proto:
	protoc -I=./ \
		--go_out=. --go_opt=module=github.com/infraconf/service-api \
		--go-grpc_out=. --go-grpc_opt=module=github.com/infraconf/service-api \
		api/*.proto
