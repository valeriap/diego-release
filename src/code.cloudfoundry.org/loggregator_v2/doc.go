package loggregator_v2

//go:generate bash -c "protoc ../loggregator-api/v2/*.proto --go_out=plugins=grpc:. --proto_path=../loggregator-api/v2"
