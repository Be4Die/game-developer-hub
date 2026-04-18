module github.com/Be4Die/game-developer-hub/gateway

go 1.25.3

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0
	github.com/kelseyhightower/envconfig v1.4.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/Be4Die/game-developer-hub/protos v0.0.0
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
)

replace github.com/Be4Die/game-developer-hub/protos => ../protos
