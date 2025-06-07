const { app, BrowserWindow, globalShortcut, Menu ,ipcMain} = require('electron');
const path = require('path');
const {ConnectServer,biDirectionalStream} = require('./grpc.js')
const {CommandMessage} = require('./proto/agent_pb');
const ffi = require('ffi-napi');
const ref = require('ref-napi');
const struct = require('ref-struct-di')(ref);


let mainWindow;
let AUTHENTICATION_KEY = null;

const voidPtr = ref.refType(ref.types.void);
const stringPtr = ref.types.CString;
const SEE_MASK_NOCLOSEPROCESS = 0x00000040;


const SHELLEXECUTEINFO = struct({
    cbSize: 'ulong',
    fMask: 'ulong',
    hwnd: voidPtr,
    lpVerb: stringPtr,
    lpFile: stringPtr,
    lpParameters: stringPtr,
    lpDirectory: stringPtr,
    nShow: 'int',
    hInstApp: voidPtr,
    lpIDList: voidPtr,
    lpClass: stringPtr,
    hkeyClass: voidPtr,
    dwHotKey: 'ulong',
    hIcon: voidPtr,
    hProcess: voidPtr,
})

function createWindow() {
    mainWindow = new BrowserWindow({
        width: 1100,
        height: 1000,
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'),
            contextIsolation: true,
            nodeIntegration: false,
            webviewTag: true,
            sandbox: false,
        }
    });

    mainWindow.loadFile('./views/index.html');

    globalShortcut.register('Control+Shift+F12', () => {
        if (mainWindow) {
            if (mainWindow.webContents.isDevToolsOpened()) {
                mainWindow.webContents.closeDevTools();
            } else {
                mainWindow.webContents.openDevTools();
            }
        }
    });
}

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') app.quit();
});

app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
    else if (mainWindow) {
        mainWindow.webContents.openDevTools();
    }
});

app.disableHardwareAcceleration();

app.on('will-quit', () => {
    globalShortcut.unregisterAll();
});

// ========================== Init Window,Grpc, Stream  ============================

app.whenReady().then(async () => {
    ShellRunServer()

    Menu.setApplicationMenu(null);
    createWindow();
    // init client and stream

    let client = false

    while (client === false) {
        client = await ConnectServer();
        if (client === false) {
            await new Promise(resolve => setTimeout(resolve, 3000));
        }
    }

    const stream = biDirectionalStream({
        onData: (data) => {
            if (data.getType() === "electron") {
                if (data.getAuth() !== "") {
                    AUTHENTICATION_KEY = data.getAuth();
                    console.log(AUTHENTICATION_KEY)
                } else {
                    console.error("Authentication fail!")
                    stream.cancel();
                }
            }
            const msg = DecodeMessage(data)
            mainWindow.webContents.send('grpc-data', msg);
        },
        onEnd: () => {
            console.log("Stream close!")
            mainWindow.webContents.send('grpc-error', "Server close!")
        },
        onError: (err) => {
            mainWindow.webContents.send('grpc-error', err.message || err.toString());
        }
    }, client)

    // send msg authentication to server
    const authMsg = new CommandMessage();
    authMsg.setType("electron")
    stream.write(authMsg)

    // listen renderer send message to stream
    ipcMain.on('grpc-send-message', (event, msg) => {
        if (stream) {
            const commandMessage = new CommandMessage()
            commandMessage.setType(msg.type)
            commandMessage.setAuth(AUTHENTICATION_KEY)
            commandMessage.setParameter(msg.parameter)
            commandMessage.setData(msg.data)
            stream.write(commandMessage); // send message to server
        } else {
            console.error("Stream error!");
        }
    });
});

function DecodeMessage(data){
    return {
        auth: data.getAuth(),
        id: data.getId(),
        type: data.getType(),
        parameter: data.getParameter(),
        error: data.getError(),
        data:data.getData(),
    };
}

function ShellRunServer(){
    const SHELLEXECUTEINFOPtr = ref.refType(SHELLEXECUTEINFO);
    const execInfo = new SHELLEXECUTEINFO();
    execInfo.cbSize = SHELLEXECUTEINFO.size;
    execInfo.fMask = SEE_MASK_NOCLOSEPROCESS;
    execInfo.hwnd = ref.NULL;
    execInfo.lpVerb = Buffer.from('open\0', 'ucs2');
    execInfo.lpFile = Buffer.from(path.resolve(__dirname, '../AgentServer/main_server_x64.exe') + '\0', 'ucs2');
    execInfo.lpParameters = ref.NULL;
    execInfo.lpDirectory = ref.NULL;
    execInfo.nShow = 5; //SW_SHOW
    execInfo.hInstApp = ref.NULL;
    execInfo.lpIDList = ref.NULL;
    execInfo.lpClass = ref.NULL;
    execInfo.hkeyClass = ref.NULL;
    execInfo.dwHotKey = 0;
    execInfo.hIcon = ref.NULL;
    execInfo.hProcess = ref.NULL;
    const shell32 = ffi.Library('shell32', {
        ShellExecuteExW: ['bool', [SHELLEXECUTEINFOPtr]],
    });
    const success = shell32.ShellExecuteExW(execInfo.ref());
}