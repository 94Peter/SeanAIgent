// HTMX CSRF Configuration
document.body.addEventListener('htmx:configRequest', (evt) => {
    const csrfInput = document.querySelector('input[name="_csrf"]');
    if (csrfInput) {
        evt.detail.headers['X-CSRF-Token'] = csrfInput.value;
    }
});

function checkinManager(initialData, sessionId, isStarted) {
    return {
        attendance: initialData,
        original: JSON.parse(JSON.stringify(initialData)),
        sessionId: sessionId,
        isStarted: isStarted,
        showAddModal: false,
        submitting: false,
        
        get hasChanges() {
            return JSON.stringify(this.attendance) !== JSON.stringify(this.original);
        },

        get stats() {
            const vals = Object.values(this.attendance);
            return {
                total: vals.length,
                attended: vals.filter(v => v === 'CheckedIn').length,
                leave: vals.filter(v => v === 'Leave').length,
                absent: vals.filter(v => v === 'Absent').length
            };
        },

        setStatus(id, status) {
            if (!this.isStarted) return;
            // Toggle logic: if clicking the same status, revert to Pending
            if (this.attendance[id] === status) {
                this.attendance[id] = 'Pending';
            } else {
                this.attendance[id] = status;
            }
        },

        async submitBatch() {
            if (this.submitting || !this.isStarted) return;
            this.submitting = true;
            
            const updates = Object.entries(this.attendance)
                .filter(([id, status]) => status !== this.original[id])
                .map(([id, status]) => ({ bookingId: id, status: status }));

            try {
                const response = await fetch('/v2/admin/checkin/batch-update', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': document.querySelector('input[name="_csrf"]').value
                    },
                    body: JSON.stringify({
                        sessionId: this.sessionId,
                        updates: updates
                    })
                });

                if (response.ok) {
                    this.original = JSON.parse(JSON.stringify(this.attendance));
                    window.location.reload(); // Refresh to get fresh data
                } else {
                    const data = await response.json();
                    alert(data.message || '儲存失敗，請檢查時間限制');
                }
            } catch (e) {
                alert('儲存出錯');
            } finally {
                this.submitting = false;
            }
        }
    }
}
