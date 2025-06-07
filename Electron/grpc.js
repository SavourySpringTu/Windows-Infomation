const grpc = require('@grpc/grpc-js');
const AgentService = require('./proto/agent_grpc_pb');

function ConnectServer() {
    const client = new AgentService.AgentServiceClient(
        'localhost:59051',
        grpc.credentials.createInsecure()
    );

    return new Promise((resolve) => {
        client.waitForReady(Date.now() + 3000, (err) => {
            if (err) {
                resolve(false);
            } else {
                resolve(client);
            }
        });
    });
}
function biDirectionalStream({ onData, onEnd, onError },client) {
    const stream = client.streamMessage();
    stream.on('data', onData);
    stream.on('end', onEnd);
    stream.on('error', onError);
    return stream;
}


module.exports = {
    ConnectServer,
    biDirectionalStream,
};

