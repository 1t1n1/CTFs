async function toggleMode() {
    const newMode = document.querySelector('meta[name="config-prefs-mode"]').getAttribute('content') === 'dark' ? 'light' : 'dark';
    try {
        const data = new URLSearchParams();
        data.append('config', 'prefs');
        data.append('key', 'mode');
        data.append('val', newMode);
        await fetch('/set-config', {
            method: 'POST',
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            body: data.toString()
        });
        location.reload();
    } catch (e) {
        // no-op
    }
}