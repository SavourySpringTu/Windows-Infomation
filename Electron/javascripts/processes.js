(function() {
    const fields = [
    ];

    // Initailized script
    window.initProcessesPage = async function() {
        Init()
    }

    function SendMessageFetchData(){
        let nameClient = GetNameClient()
        const path = document.getElementById("path").value
        const msg = {
            type: "-p",
            parameter: path,
            data: nameClient
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
