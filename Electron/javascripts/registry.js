(function() {
    const fields = [
    ];

    // Initailized script
    window.initRegistryPage = async function() {
        Init()
    }

    function SendMessageFetchData(){
        let nameClient = GetNameClient()
        const path = document.getElementById("path").value
        const msg = {
            type: "-r",
            data: nameClient,
            parameter: path
        };
        window.electronAPI.sendGrpcMessage(msg);
    }

    function Init(){
        const clientBtn = document.getElementById('client-btn');
        const fetchBtn = document.getElementById('fetch-btn');

        clientBtn.addEventListener('click', () => { SendMessageFetchClient(); });
        fetchBtn.addEventListener('click', () => { SendMessageFetchData(); });

        // Register listener to receive data from main
        window.electronAPI.onGrpcData((msg) => {
            ReceiveMessage(msg,fields);
        });
    }

})();
