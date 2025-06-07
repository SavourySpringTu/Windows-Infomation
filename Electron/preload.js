const { contextBridge, ipcRenderer } = require('electron');
const YAML = require('yaml');

let grpcCallback = null; // save callback

contextBridge.exposeInMainWorld('electronAPI', {

    sendGrpcMessage: (msg) => ipcRenderer.send('grpc-send-message', msg),

    // register a listener to receive data from main
    onGrpcData: (callback) => {
        if (!grpcCallback) {
            grpcCallback = (event, data) => callback(data);
            ipcRenderer.on('grpc-data', grpcCallback);
        }
    },

    parse: (str) => YAML.parse(str),
    stringify: (obj) => YAML.stringify(obj),
});
