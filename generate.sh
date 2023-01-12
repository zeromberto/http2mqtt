#!/usr/bin/env bash
set -e
shopt -s globstar
cd `dirname $0`
source .bingo/variables.env

rm -f **/*.pb.go **/*.pb.gw.go swagger/*.json auth/*auth.proto


for FILE in ${APIS_PATH:-apis/proto}/**/*.proto; do
    ./0install.sh run --version 2.11.3 https://apps.0install.net/protobuf/protoc-gen-grpc-gateway.xml \
        -I ${APIS_PATH:-apis/proto} \
        -I apis/vendor/api-common-protos \
        -I apis/vendor/protoc-gen-validate \
		    --plugin=protoc-gen-go=$PROTOC_GEN_GO \
        --go_out=plugins=grpc,paths=source_relative:. \
        --grpc-gateway_out=logtostderr=true,paths=source_relative:. \
        $FILE
done

pushd _hack
go run generate.go --template gateway --out internal/registrations --apis ${APIS_PATH:-apis/proto}
go run generate.go --template auth_proto --out auth --apis ${APIS_PATH:-apis/proto}
go run generate.go --template index_html --out swagger-ui-dist --apis ${APIS_PATH:-apis/proto}
popd

for AUTH_PROTO in auth/*.auth.proto; do

    FILENAME=${AUTH_PROTO#*/}
    DOMAIN_AND_VERSION=${FILENAME%.auth.proto}

    DIR=$(echo ${DOMAIN_AND_VERSION} | sed -e 's/\./\//g' )
    PRODUCT=${DOMAIN_AND_VERSION%.*}
    DOMAIN=${DOMAIN_AND_VERSION%.*}
    VERSION=${DOMAIN_AND_VERSION#*.}

    ./0install.sh run --version 3.20.1 https://apps.0install.net/protobuf/protoc.xml \
        -I ${APIS_PATH:-apis/proto} \
        -I apis/vendor/api-common-protos \
        -I apis/vendor/protoc-gen-validate \
        -I auth \
        --plugin=protoc-gen-openapiv2=$PROTOC_GEN_OPENAPIV2 \
        --openapiv2_out=fqn_for_openapi_name=true,include_package_in_tags=true,disable_default_errors=true,allow_merge=true,json_names_for_fields=false,merge_file_name=${DOMAIN_AND_VERSION//./_}:swagger/ \
        ${APIS_PATH:-apis/proto}/$DIR/*.proto $AUTH_PROTO
done

# Download working version of jq
curl -L -s -o ./jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64
chmod +x ./jq

# Postprocess swagger specs
for FILE in swagger/*.json; do
    cp -T $FILE tmp.json
    cat tmp.json | \
    jq 'del(.tags)' | \
    jq 'del( .. | objects | select(.name=="filter.field_mask.paths"))' | \
    jq 'del( .. | objects | select(.["$ref"]=="#/definitions/google.protobuf.FieldMask"))' | \
    jq 'del( .. | objects | .["google.protobuf.FieldMask"])' | \
    jq 'del(.paths[][].parameters[] | select(.name == "filter.field_mask"))' | \
    jq 'del(.paths[][].responses[].schema.properties.error)' | \
    jq 'del(.definitions[].title)' | \
    jq 'del(.paths[][].parameters[].default)' | \
    jq '(.paths[][].parameters[] | select(.description != null) | .description) |=  sub("\\s{1}(?=-{1}.*UNSPECIFIED).*?(?<=\\\n)"; ""; "gm")' | \
    jq '(.paths[][].parameters[] | select(.enum != null) | .enum) |=  map(select(. | contains("UNSPECIFIED") | not))' > $FILE
    # Fix required fields in filters
    # TODO: This is a temporary fix until we have a better solution, like better OpenAPI 3 Generator
    # We will revisit the issue in the near future.
    if [[ $FILE == *".swagger.json" ]]; then
        cat $FILE > tmp.json
        cat tmp.json | ./jq '(.paths[][].parameters[] | select(.name | startswith("filter.")).required) = false' > $FILE
    fi
    rm -f tmp.json
done

# Verify swagger specs
COUNT=$(wc -l swagger/*.json | tail -1 | tr -d -c 0-9)
if [ ${COUNT} -eq 0 ]; then
  echo "swagger specs are corrupt"
  exit 1
fi
