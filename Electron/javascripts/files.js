(function() {
    const fields = [
        'name',
        'datecreated',
        'datemodified',
        'size',
        'md5',
        'sha1',
        'sha_256'
    ];

    // Initailized script
    window.initFilesPage = async function() {
        Init()
    }

    function SendMessageFetchData(authKey){
        let nameClient = GetNameClient()
        const path = document.getElementById("path").value
        const msg = {
            type: "-f",
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
