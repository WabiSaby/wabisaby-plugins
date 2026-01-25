module github.com/wabisaby/wabisaby-plugins/wabisaby-network-storage

go 1.24.0

require (
	github.com/google/uuid v1.6.0
	github.com/wabisaby/wabisaby v0.0.0
	github.com/wabisaby/wabisaby-plugin-sdk v0.0.0
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.7
)

replace github.com/wabisaby/wabisaby => ../../../WabiSaby-Go
replace github.com/wabisaby/wabisaby-plugin-sdk => ../../../WabiSaby-Plugin-SDK
