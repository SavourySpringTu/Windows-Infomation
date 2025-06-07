#   Agent
A Golang-based Windows client that uses WinAPI to collect system information and communicates with a central server via gRPC. The project also includes a special GUI client that connects to the server for remote control.

## Features

- Collect system information (os, hardware, registry, dns, connections, network, BIOS)
- List installed apps
- List process
- List information files (hash)
- Multi client connect to server
- GUI control server

## Tech Stack

- Golang
- C/C++
- Windows API
- gRPC
- NodeJS
- Electron
- Python
- PowerShell

## Installation & Usage
- You should read tab details before run code
- 
- git clone https://github.com/SavourySpringTu/Agent
- .\make b
- cd Electron
- npm install
- cd ..
- .\make r e
- .\make r c

## Details
- You should build (at least build main_server.go in AgentServer) before run Electron
- You can read file make.ps1 for more information (run, build, gen file proto)
- For Electron, if you want "npm install", you need to pay attention to versions python 3.10, ffi-napi 4.0.3 ,node-gyp 9.4.0 and install build tools (microsoft). 
- If you want build Electron you should edit code. In Electron/node-modules/deps/libffi/libffi.gyp. Replace line 76 to:
```
'action': [
  '../../../deps/libffi/preprocess_asm.cmd',
    'include',
    'config/<(OS)/<(target_arch)',
    '<(RULE_INPUT_PATH)',
    '<(INTERMEDIATE_DIR)/<(RULE_INPUT_ROOT).asm',
  ],
```
  
- You should run client by Administrator

## License

MIT License. See [LICENSE](./LICENSE) for details.
