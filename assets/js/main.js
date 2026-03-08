// Global Application Logic

// HTMX CSRF Configuration
document.body.addEventListener('htmx:configRequest', (evt) => {
    const csrfInput = document.querySelector('input[name="_csrf"]');
    if (csrfInput) {
        evt.detail.headers['X-CSRF-Token'] = csrfInput.value;
    }
});

// Toast System
function showToast({ title, description, variant = 'default', duration = 5000 }) {
    const trigger = document.getElementById('toast-trigger');
    if (!trigger) {
        console.error('Global toast trigger element (#toast-trigger) not found!');
        return;
    }
    const params = new URLSearchParams({
        title: title || '',
        description: description || '',
        variant: variant,
        duration: duration,
    });
    const url = `/components/toast?${params.toString()}`;
    trigger.setAttribute('hx-get', url);
    htmx.process(trigger);
    trigger.click();
}

document.body.addEventListener('showToast', function(evt) {
    if (evt.detail) {
        showToast({
            title: decodeBase64Utf8(evt.detail.title), 
            description: decodeBase64Utf8(evt.detail.description), 
            variant: evt.detail.variant,
        });
    }
});

function decodeBase64Utf8(base64String) {
    try {
        const binaryString = atob(base64String);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
            bytes[i] = binaryString.charCodeAt(i);
        }
        const decoder = new TextDecoder('utf-8');
        return decoder.decode(bytes);
    } catch (e) {
        console.error("Base64 decode or UTF-8 parse failed:", e);
        return null; 
    }
}
