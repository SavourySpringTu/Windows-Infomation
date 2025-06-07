const {app, BrowserWindow,Menu} = require('electron')

function createWindow(){
    const win = new BrowserWindow({
        width: 900,
        height: 700,
        webPreferences: {
            nodeIntegration: true,
            contextIsolation: false,
            webviewTag: true,
        }
    })
    win.loadFile('index.html');
}

app.whenReady().then(()=>{
    Menu.setApplicationMenu(null);
    createWindow();
});

app.on('window-all-closed',()=>{
    if (process.platform !== 'darwin') app.quit();
})

app.on('activate', ()=>{
    if(BrowserWindow.getAllWindows().length ===0)createWindow();
})

app.disableHardwareAcceleration();