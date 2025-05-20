param (
    [string]$action = $args[0]
)
if($action -eq "proto"){
    Remove-Item .\proto\agent_client.pb.go -ErrorAction SilentlyContinue
    Remove-Item .\proto\agent_client_grpc.pb.go -ErrorAction SilentlyContinue
    protoc --proto_path=proto --go_out=proto --go-grpc_out=proto proto/agent_client.proto
}elseif ($action -eq "buildx64"){
    $ENV:GOARCH = "amd64"
    go build -o main_client_x64.exe main_client.go
    go build -o main_server_x64.exe main_server.go
    go build -o main_x64.exe main.go
}elseif ($action -eq "buildx86"){
    $ENV:GOARCH = "386"
    go build -o main_client_x86.exe main_client.go
    go build -o main_sever_x86.exe main_server.go
    go build -o main_x86.exe main.go
    $ENV:GOARCH = "amd64"
}