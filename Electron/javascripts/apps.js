(function() {
    const fields = [
        'name',
        'version',
        'publisher',
        'installdate',
    ];

    // Initailized script
    window.initAppsPage = async function() {
        Init(fields)
    }

    function SendMessageFetchData(){
        let nameClient = GetNameClient()
        const msg = {
            type: "-a",
            data: nameClient
        };
        window.electronAPI.sendGrpcMessage(msg);
    }

    function Init(authKey,fields){


        const clientBtn = document.getElementById('client-btn');
        const fetchBtn = document.getElementById('fetch-btn');

        clientBtn.addEventListener('click', () => { SendMessageFetchClient(); });
        fetchBtn.addEventListener('click', () => { SendMessageFetchData(); });

        window.electronAPI.onGrpcData((msg) => {
            ReceiveMessage(msg,fields);
        });
    }

})();
