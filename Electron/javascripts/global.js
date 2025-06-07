function SendMessageFetchClient(){
    const msg = {
        type: "loadclient",
    };
    window.electronAPI.sendGrpcMessage(msg);
}

function ReceiveMessage(msg,fields){
    console.log("Received message:", msg);
    if (!msg.error || msg.error === "") {
        if (msg.type === "loadclient") {
            setSelectClient(msg.data);
        } else {
            setInputInfo(msg,fields);
        }
    } else {
        showError(msg.error)
    }
}


function setSelectClient(listClient) {
    const clientMap = window.electronAPI.parse(listClient);
    if (Object.keys(clientMap).length === 0) {
        showError("No clients connected!")
        return;
    }
    CleanAllInput();

    const select = document.getElementById('client-select');
    select.innerHTML = '';

    Object.entries(clientMap).forEach(([key, value]) => {
        const option = document.createElement('option');
        option.value = key;
        option.textContent = value;
        select.appendChild(option);
    });
}

function setInputInfo(msg,fields){
    const data = window.electronAPI.parse(msg.data);
    CleanAllInput();
    if (!data) {
        console.warn("Parsed data is null or undefined");
        return;
    }

    if (msg.type === "-a") {
        setInputForApps(data)
    }else if(msg.type==="-c"){
        setInputConnections(data)
    }else if(msg.type === "-d"){
        setInputForDns(data)
    }else if(msg.type === "-f"){
        setInputForFiles(data)
    }else if(msg.type === "-k"){
        setInputForKernels(data)
    }
    else if (msg.type==="-r"){
        setInputForRegistry(data)
    }
    else if(msg.type==="-p"){
        setInputProcesses(data)
    }
    else if(msg.type ==="-s"){
        setInputSystem(data)
    }else if (msg.type ==="-i"){
        setInputHardWare(data)
    }
    else if (msg.type ==="-b"){
        setInputBIOS(data)
    }
}

function setInputHardWare(data){
    document.getElementById('namecpu').value = data["namecpu"]
    document.getElementById('ram').value = data["ram"]
    document.getElementById('sizedisk').value = data["sizedisk"]
}

function setInputSystem(data){
    document.getElementById('namepc').value = data["namepc"]
    document.getElementById('nameos').value = data["nameos"]
    document.getElementById('hostname').value = data["hostname"]
    document.getElementById('timeup').value = data["timeup"]
}

function setInputBIOS(data){
    document.getElementById('systemmanufacturer').value = data["systemmanufacturer"]
    document.getElementById('systemmodel').value = data["systemmodel"]
    document.getElementById('uuid').value = data["uuid"]
    document.getElementById('processor').value = data["processor"]
    document.getElementById('baseboardmanufacturer').value = data["baseboardmanufacturer"]
    document.getElementById('baseboardserial').value = data["baseboardserial"]
    document.getElementById('baseboardmodel').value = data["baseboardmodel"]
    document.getElementById('diskdriveserial').value = data["diskdriveserial"]
}

function setInputConnections(data) {
    const tcpTbody = document.querySelector('#tcp-table tbody');
    const udpTbody = document.querySelector('#udp-table tbody');
    tcpTbody.innerHTML = '';
    udpTbody.innerHTML = '';

    // TCP
    Object.values(data.connectiontcpinfo || {}).forEach(records => {
        records.forEach(conn => {
            const row = document.createElement('tr');
            row.innerHTML = `
        <td>${conn.pid}</td>
        <td>${conn.localaddress}</td>
        <td>${conn.remoteaddress}</td>
        <td>${conn.localport}</td>
        <td>${conn.remoteport}</td>
        <td>${conn.state}</td>
      `;
            tcpTbody.appendChild(row);
        });
    });

    // UDP
    Object.values(data.connectionudpinfo || {}).forEach(records => {
        records.forEach(conn => {
            const row = document.createElement('tr');
            row.innerHTML = `
        <td>${conn.pid}</td>
        <td>${conn.localaddr}</td>
        <td>${conn.remoteaddr || '-'}</td>
        <td>${conn.localport}</td>
        <td>${conn.remoteport}</td>
      `;
            udpTbody.appendChild(row);
        });
    });
}


function setInputForDns(data) {
    const tbody = document.querySelector('#table-dns tbody');
    tbody.innerHTML = '';

    Object.entries(data||{}).forEach(([domain, records]) => {
        const row = document.createElement('tr');

        const recordRows = records.map(record => `
            <div><div>${record.type}</div> ${record.recordname}: ${record.record}</div>
        `).join('');

        row.innerHTML = `
            <td>${domain}</td>
            <td>${records.length ? recordRows : '<em>Empty record</em>'}</td>
        `;
        tbody.appendChild(row);
    });
}

function setInputForApps(data){
    const table = document.getElementById("table-apps")
    const tbody = table.querySelector("tbody");
    tbody.innerHTML = '';

    Object.values(data || {}).forEach(entry => {
        const row = document.createElement('tr');
        row.innerHTML = `
                <td>${entry.name}</td>
                <td>${entry.version}</td>
                <td>${entry.publisher}</td>
                <td>${entry.installdate}</td>
              `;
        table.appendChild(row);
    });
}

function setInputForFiles(data){
    const tbody = document.querySelector('#table-apps tbody');
    tbody.innerHTML = '';

    Object.values(data || {}).forEach(entry => {
        const row = document.createElement('tr');
        row.innerHTML = `
                <td>
                  <div><strong>Name:</strong> ${entry.name}</div>
                  <div><strong>Date Created:</strong> ${entry.datecreated}</div>
                  <div><strong>Date Modified:</strong> ${entry.datemodified}</div>
                  <div><strong>Size:</strong> ${entry.size}</div>
                  <div><strong>MD5:</strong> ${entry.md5}</div>
                  <div><strong>SHA 1:</strong> ${entry.sha1}</div>
                  <div><strong>SHA 256:</strong> ${entry.sha_256}</div>
                </td>
              `;
        tbody.appendChild(row);
    });
}

function setInputForRegistry(data){
    const tbody = document.querySelector('#table-registry tbody');
    tbody.innerHTML = '';

    Object.values(data || {}).forEach(entry => {
        const row = document.createElement('tr');
        row.innerHTML = `
                <td>
                  <div><strong>Name:</strong> ${entry.name}</div>
                  <div><strong>Value:</strong> ${entry.data}</div>
                </td>
              `;
        tbody.appendChild(row);
    });
}

function setInputForKernels(data){
    const tbody = document.querySelector('#table-kernels tbody');
    tbody.innerHTML = '';

    Object.values(data || {}).forEach(entry => {
        const row = document.createElement('tr');
        row.innerHTML = `
                <td>
                  <div><strong>Name:</strong> ${entry.name}</div>
                  <div><strong>Path:</strong> ${entry.path}</div>
                  <div><strong>SHA 256:</strong> ${entry.sha256}</div>
                  <div><strong>Startup Mode:</strong> ${entry.startupmode}</div>
                  <div><strong>State:</strong> ${entry.state}</div>
                </td>
              `;
        tbody.appendChild(row);
    });
}

function setInputProcesses(processes) {
    const container = document.getElementById('process-container');
    container.innerHTML = ''; // clear cũ nếu cần

    Object.values(processes || {}).forEach(entry => {
        const div = document.createElement('div');
        div.classList.add('process-box');
        div.style.border = '1px solid #ccc';
        div.style.padding = '10px';
        div.style.marginBottom = '20px';

        div.innerHTML = `
            <h3>${entry.name} (PID: ${entry.pid})</h3>
            <p><strong>Command Line:</strong> ${entry.commandline}</p>
            <p><strong>Parent PID:</strong> ${entry.parentpid}</p>
            <p><strong>Runtime:</strong> ${entry.runtime}</p>

            <details>
                <summary><strong>Token</strong></summary>
                <ul>
                    <li><strong>User:</strong> ${entry.token?.user || 'N/A'}</li>
                    <li><strong>SID:</strong> ${entry.token?.sid || 'N/A'}</li>
                    <li><strong>Session:</strong> ${entry.token?.session}</li>
                    <li><strong>Logon Session:</strong> ${entry.token?.logonsession}</li>
                    <li><strong>Virtualized:</strong> ${entry.token?.virtualized}</li>
                    <li><strong>Protected:</strong> ${entry.token?.protected}</li>
                </ul>
            </details>

            <details>
                <summary><strong>Groups (${entry.token?.groups?.length || 0})</strong></summary>
                <ul>
                    ${entry.token?.groups?.map(g => `<li>${g.name} (${g.sid})</li>`).join('') || ''}
                </ul>
            </details>

            <details>
                <summary><strong>Privileges (${entry.token?.privileges?.length || 0})</strong></summary>
                <ul>
                    ${entry.token?.privileges?.map(p => `<li>${p.name} (LUID: ${p.luid})</li>`).join('') || ''}
                </ul>
            </details>

            <details>
                <summary><strong>Modules (${entry.module?.length || 0})</strong></summary>
                <ul>
                    ${entry.module?.map(m => `<li>${m.name}</li>`).join('') || ''}
                </ul>
            </details>

            <details>
                <summary><strong>TCP Connections (${entry.connectiontcp?.length || 0})</strong></summary>
                ${entry.connectiontcp?.length ? `
                <table border="1" cellpadding="5">
                    <tr><th>Local Address</th><th>Remote Address</th><th>Local Port</th><th>Remote Port</th><th>State</th></tr>
                    ${entry.connectiontcp.map(c => `
                        <tr>
                            <td>${c.localaddress}</td>
                            <td>${c.remoteaddress}</td>
                            <td>${c.localport}</td>
                            <td>${c.remoteport}</td>
                            <td>${c.state}</td>
                        </tr>
                    `).join('')}
                </table>` : '<p>No TCP connections.</p>'}
            </details>

            <details>
                <summary><strong>UDP Connections (${entry.connectionudp?.length || 0})</strong></summary>
                ${entry.connectionudp?.length ? `
                <table>
                    <tr><th>Local Address</th><th>Remote Address</th><th>Local Port</th><th>Remote Port</th></tr>
                    ${entry.connectionudp.map(c => `
                        <tr>
                            <td>${c.localaddress}</td>
                            <td>${c.remoteaddress}</td>
                            <td>${c.localport}</td>
                            <td>${c.remoteport}</td>
                        </tr>
                    `).join('')}
                </table>` : '<p>No UDP connections.</p>'}
            </details>
        `;

        container.appendChild(div);
    });
}



// ==============================================================================================
function CleanAllInput() {
    const inputs = document.querySelectorAll('input');
    inputs.forEach(input => {
        input.value = '';
    });
}

function GetNameClient(){
    const client = document.getElementById('client-select');
    const nameClient = client.options[client.selectedIndex]?.text || "";
    if (nameClient === "") {
        client.value = "Please choose client!" ;
        return ""
    }
    return nameClient
}

function showError(err) {
    const alertBox = document.getElementById("error-alert");
    alertBox.textContent = err;
    alertBox.classList.add("show");
    alertBox.classList.remove("hide");
    console.log("loi ne", err)
    setTimeout(() => {
        alertBox.classList.add("hide");
        setTimeout(() => {
            alertBox.classList.remove("show");
        }, 500);
    }, 3000);
}