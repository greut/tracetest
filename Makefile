
server-generate:
	openapi-generator-cli generate -i api/openapi.yaml -g go-server -o server/

init-submodule:
	git submodule init
	git submodule update

PROTOC_VER=0.9.0
PROTOC_IMAGE=otel/build-protobuf:$(PROTOC_VER)
PROTOC=docker run --rm -u ${shell id -u} -v "${PWD}:${PWD}" -w ${PWD} ${PROTOC_IMAGE} --proto_path=${PWD}

#OTEL_DOCKER_PROTOBUF ?= otel/build-protobuf:0.9.0
#PROTOC := docker run --rm -u ${shell id -u} -v${PWD}:${PWD} -w${PWD} ${OTEL_DOCKER_PROTOBUF} --proto_path=${PWD}

PROTO_INCLUDES := \
	-Ijaeger-idl/proto \
	-I/usr/include/github.com/gogo/protobuf \
	-Iopentelemetry-proto \
	-Iopentelemetry-proto/opentelemetry/proto

PROTO_GOGO_MAPPINGS := $(shell echo \
		Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/types, \
		Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types, \
		Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types, \
		Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types, \
		Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api, \
		Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api, \
		Mopentelemetry/proto/trace/v1/trace.proto=go.opentelemetry.io/proto/otlp/trace/v1, \
		Mtrace/v1/trace.proto=go.opentelemetry.io/proto/otlp/trace/v1, \
		Mmodel.proto=github.com/jaegertracing/jaeger/model \
	| sed 's/ //g')

PROTO_GEN_GO_DIR ?= server/go/internal/proto-gen-go

PROTOC_WITH_GRPC := $(PROTOC) \
		$(PROTO_INCLUDES) \
		--gogo_out=plugins=grpc,$(PROTO_GOGO_MAPPINGS):$(PWD)/${PROTO_GEN_GO_DIR}

PROTOC_INTERNAL := $(PROTOC) \
		$(PROTO_INCLUDES)

SWAGGER_VER=0.12.0
SWAGGER_IMAGE=quay.io/goswagger/swagger:$(SWAGGER_VER)
SWAGGER=docker run --rm -u ${shell id -u} -v "${PWD}:/go/src/${PROJECT_ROOT}" -w /go/src/${PROJECT_ROOT} $(SWAGGER_IMAGE)

proto:
	rm -rf ./$(PROTO_GEN_GO_DIR)
	mkdir -p ${PROTO_GEN_GO_DIR}
	mkdir -p swagger

	# API v3
	$(PROTOC_WITH_GRPC) \
		jaeger-idl/proto/api_v3/query_service.proto

	$(PROTOC_INTERNAL) \
<<<<<<< Updated upstream
		--swagger_out=disable_default_errors=true,json_names_for_fields=true,logtostderr=true:./swagger \
=======
		--openapiv2_out=disable_default_errors=true,openapi_naming_strategy=legacy,logtostderr=true:./swagger \
>>>>>>> Stashed changes
		jaeger-idl/proto/api_v3/query_service.proto

	$(PROTOC_WITH_GRPC) \
		tempo-idl/tempo.proto
	cp tempo-idl/prealloc.go server/go/internal/proto-gen-go/tempo-idl/

swagger: proto
	cp swagger/api_v3/query_service.swagger.json api/trace.json
	cat api/trace.json | yq e -P - > api/trace.yaml
	rm api/trace.json
