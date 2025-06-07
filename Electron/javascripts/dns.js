(function() {
    const fields = [
        'name',
        'version',
        'publisher',
        'installdate',
    ];

    // Initailized script
    window.initDnsPage = async function() {
        Init()
    }

    function SendMessageFetchData(){
        let nameClient = GetNameClient()
        const msg = {
            type: "-d",

            data: nameClient
        };
        window.electronAPI.sendGrpcMessage(msg);
    }

    function Init(authKey){

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
