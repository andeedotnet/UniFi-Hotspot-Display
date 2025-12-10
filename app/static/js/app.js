async function loadVouchers() {
    try {
        const res = await fetch('/json/vouchers');
        const codes = await res.json();
        const container = document.getElementById('vouchers');
        if (!container) return;
        container.innerHTML = codes.map(c => `<div class="list-group-item font-monospace border-0 bg-transparent text-white fs-5">${formatVoucher(c)}</div>`).join('');
    } catch (e) {
        const container = document.getElementById('vouchers');
        if (container) container.textContent = 'Error loading vouchers';
        console.error(e);
    }
}

function formatVoucher(code) {

    return code.slice(0, 5) + '-' + code.slice(5);

}

async function loadWifiData() {
    try {
        const res = await fetch('/json/wifi_data');
        const data = await res.json();
        document.getElementById('wifi_ssid').textContent = data.ssid;
        document.getElementById('wifi_password').textContent = data.password;
    } catch (e) {
        const container = document.getElementById('wifi_ssid');
        if (container) container.textContent = 'Error loading WiFi info';
        console.error(e);
    }
}

window.addEventListener('DOMContentLoaded', () => {
    loadVouchers();
    loadWifiData();
    setInterval(loadVouchers, 5000);
});