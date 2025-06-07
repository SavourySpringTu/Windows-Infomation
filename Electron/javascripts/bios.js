(function() {
    const fields = [
    'systemmanufacturer',
    'uuid',
    'systemmodel',
    'processor',
    'baseboardmanufacturer',
    'baseboardmodel',
    'diskdriveserial',
    'baseboardserial'
];

    // Initailized script
    window.initBiosPage = async function() {
        Init()
    }

    function SendMessageFetchData(){
        let nameClient = GetNameClient()
        const msg = {
            type: "-b",
            data: nameClient
        };
        window.electronAPI.sendGrpcMessage(msg);
    }

    function Init(){

        const clientBtn = document.getElementById('client-btn');
        const fetchBtn = document.getElementById('fetch-btn');

        clientBtn.addEventListener('click', () => { SendMessageFetchClient(); });
        fetchBtn.addEventListener('click', () => { SendMessageFetchData(); });

        window.electronAPI.onGrpcData((msg) => {
            ReceiveMessage(msg,fields);
        });
    }
})();
