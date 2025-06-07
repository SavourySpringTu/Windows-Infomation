param(
    [string]$action = $args[0],
    [string]$param = $args[1]
)

if($action -eq "proto"){
    Remove-Item .AgentClient/proto/agent.pb.go -ErrorAction SilentlyContinue
    Remove-Item .AgentClient/proto/agent_grpc.pb.go -ErrorAction SilentlyContinue
    Remove-Item .AgentServer/proto/agent.pb.go -ErrorAction SilentlyContinue
    Remove-Item .AgentServer/proto/agent_grpc.pb.go -ErrorAction SilentlyContinue

    protoc --proto_path=./shared-protos --go_out=./AgentClient --go-grpc_out=./AgentClient agent.proto
    protoc --proto_path=./shared-protos --go_out=./AgentServer --go-grpc_out=./AgentServer agent.proto
    Push-Location ./Electron
    try{
        Remove-Item ./proto/agent.pb.go -ErrorAction SilentlyContinue
        Remove-Item ./proto/agent_grpc.pb.go -ErrorAction SilentlyContinue
        npm run gen:proto
    }finally{
        Pop-Location
    }
}
elseif ($action -eq "b") {
    Push-Location ./AgentServer
    try{
        $ENV:GOARCH = "386"
        go build -o main_server_x86.exe main_server.go
        $ENV:GOARCH = "amd64"
        go build -o main_server_x64.exe main_server.go
    }finally{
        Pop-Location
    }

    Push-Location ./AgentClient
    try{
        $ENV:GOARCH = "386"
        go build -o main_client_x86.exe main_client.go
        $ENV:GOARCH = "amd64"
        go build -o main_client_x64.exe main_client.go
    }finally{
        Pop-Location
    }
    Push-Location ./Electron
    try{
        npm run build
    }finally{
        Pop-Location
    }
}
elseif ($action -eq "r"){
    if ($param -eq "s")
    {
        Push-Location ./AgentServer
        try{
            go run main_server.go
        }finally{
            Pop-Location
        }
    }
    elseif ($param -eq "c"){
        Push-Location ./AgentClient
        try{
            go run main_client.go
        }finally{
            Pop-Location
        }
    }
    elseif($param -eq "e"){
        try{
            Push-Location ./Electron
            try{
                npm start
            }finally{
                Pop-Location
            }
        }finally{
            Pop-Location
        }
    }
}
elseif($action -eq "rb"){
    if($param -eq "e"){
        Push-Location ./Electron/dist/win-unpacked
        try{
            .\electron.exe
        }finally{
            Pop-Location
        }
    }
    elseif($param -eq "s"){
        Push-Location ./AgentServer
        try{
            .\main_server_x64.exe
        }finally{
            Pop-Location
        }
    }
    elseif($param -eq "c"){
        Push-Location ./AgentClient
        try{
            .\main_client_x64.exe
        }finally{
            Pop-Location
        }
    }
}