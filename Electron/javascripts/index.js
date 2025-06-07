document.addEventListener('DOMContentLoaded', () => {
    const buttons = document.querySelectorAll('.sidebar button');
    const staticTabs = document.querySelectorAll('.static-tab');
    const tabContainer = document.getElementById('tab-container');

    const loadedScripts = new Set(); // Track loaded scripts

    function hideAllContent() {
        staticTabs.forEach(tab => tab.classList.remove('active'));
        tabContainer.innerHTML = '';
        tabContainer.style.display = 'none';
        buttons.forEach(button => button.classList.remove('active'));
    }

    async function loadTab(tabId, targetUrl) {
        hideAllContent();
        const button = document.querySelector(`.sidebar button[data-tab-id="${tabId}"]`);
        if (button) button.classList.add('active');

        if (targetUrl) {
            try {
                const response = await fetch(targetUrl);
                if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
                const html = await response.text();
                tabContainer.innerHTML = html; // update new html
                tabContainer.style.display = 'block';

                const scriptName = targetUrl.split('/').pop().replace('.html', '.js');
                const scriptPath = `../javascripts/${scriptName}`;

                // Load script if not already loaded
                if (!loadedScripts.has(scriptPath)) {
                    await loadScript(scriptPath);
                    loadedScripts.add(scriptPath);
                }

                await new Promise(resolve => setTimeout(resolve, 0));

                const initFunctionName = `init${capitalize(tabId)}Page`;

                if (typeof window[initFunctionName] === 'function') {
                    window[initFunctionName](); // Call init function
                } else {
                    console.warn(`Function ${initFunctionName} not found.`);
                }
            } catch (error) {
                console.error('Error loading content:', error);
                tabContainer.innerHTML = '<p>Error loading content.</p>';
                tabContainer.style.display = 'block';
            }
        } else {
            // Logic for static tab (home)
            const staticTab = document.getElementById(tabId);
            if (staticTab) {
                staticTab.classList.add('active');
            }
        }
    }

    function loadScript(scriptUrl) {
        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            script.src = scriptUrl;
            script.onload = () => {
                resolve();
            };
            script.onerror = () => {
                console.error(`Failed to load script: ${scriptUrl}`);
                reject(new Error(`Failed to load script: ${scriptUrl}`));
            };
            document.body.appendChild(script);
        });
    }

    function capitalize(str) {
        return str.charAt(0).toUpperCase() + str.slice(1);
    }

    buttons.forEach(button => {
        button.addEventListener('click', () => {
            const tabId = button.getAttribute('data-tab-id');
            const targetUrl = button.getAttribute('data-target-url');
            loadTab(tabId, targetUrl);
        });
    });

    // Auto click home when start
    const homeButton = document.querySelector('.sidebar button[data-tab-id="home"]');
    if (homeButton) {
        homeButton.click();
    }
});
