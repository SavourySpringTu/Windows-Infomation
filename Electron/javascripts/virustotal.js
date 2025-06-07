(function() {
    API_KEY_VIRUSTOTAL = "eb1c2840b49435b10cadda14f7c3c03f5918e81babde477de89a55da9f542b64"
    URL_VIRUSTOTAL = "https://www.virustotal.com/api/v3/files/"
    const fields = [
    ];

    // Initailized script
    window.initVirustotalPage = async function() {
        const authKey = await window.electronAPI.getAuthKey();
        Init(authKey)
    }

    function SendMessageFetchData(authKey){
        let nameClient = GetNameClient()
        const path = document.getElementById("path").value
        const msg = {
            type: "-f",
            auth: authKey,
            parameter: path,
            data: nameClient
        };
        window.electronAPI.sendGrpcMessage(msg);
    }

    function Init(authKey){
        const clientBtn = document.getElementById('client-btn');
        const fetchBtn = document.getElementById('fetch-btn');

        clientBtn.addEventListener('click', () => { SendMessageFetchClient(authKey); });
        fetchBtn.addEventListener('click', () => { SendMessageFetchData(authKey); });

        // Register listener to receive data from main
        window.electronAPI.onGrpcData((msg) => {
            AnalyzeFile(msg);
        });
    }

    async function AnalyzeFile(msg) {
        sha256 = msg.data[0].sha_256
        console.log()
        const url = URL_VIRUSTOTAL+sha256
        responseVirusTotal = window.electronAPI.axiosGet(url,{
            headers:{
                'x-apikey': API_KEY_VIRUSTOTAL
            }
        });
        const data = responseVirusTotal.data;
        console.log(data)
    }
})();


