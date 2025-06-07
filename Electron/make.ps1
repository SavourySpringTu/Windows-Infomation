param(
    [string]$action = $args[0]
)

if($action -eq "proto"){
    Remove-Item .\proto\agent_client.pb.go -ErrorAction SilentlyContinue
    Remove-Item .\proto\agent_client_grpc.pb.go -ErrorAction SilentlyContinue
    protoc --go_out=. --go-grpc_out=. proto\agent_server.proto
}