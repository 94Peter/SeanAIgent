// Booking V2 Calendar and Action Logic (Stable Version)

const fetchApi = async (u, o = {}) => {
    const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
    const headers = { 'Content-Type': 'application/json', ...o.headers };
    
    if (csrfToken && ['POST', 'PUT', 'DELETE', 'PATCH'].includes(o.method?.toUpperCase())) {
        headers['X-CSRF-TOKEN'] = csrfToken;
    }

    const opts = {
        ...o,
        headers,
        credentials: 'same-origin'
    };

    const res = await fetch(u, opts);
    const d = await res.json(); 
    if (!res.ok) throw new Error(d.message || 'API failed'); 
    return d;
};

const getStatusClasses = (s) => {
    const m = {
        'Booked': 'bg-[#60A5FA] text-white border-transparent',
        'Leave': 'bg-[#F59E0B]/10 text-[#F59E0B] border border-[#F59E0B]/30',
        'CheckedIn': 'bg-[#10B981]/10 text-[#10B981] border border-[#10B981]/30',
        'Absent': 'bg-[#EF4444]/10 text-[#EF4444] border border-[#EF4444]/30'
    };
    return m[s] || 'bg-[#3F3F46]/10 text-[#A1A1AA] border border-[#3F3F46]/30';
};

const jsRenderTag = (p, mini) => {
    const s = p.status || p.Status, n = p.name || p.Name, classes = "rounded transition-colors " + getStatusClasses(s);
    if (mini) return ("<span class=\"text-[9px] px-1.5 py-0.5 rounded-[4px] overflow-hidden whitespace-nowrap block " + classes + "\">" + n + "</span>");
    const suffix = s === 'Leave' ? ' (請假)' : (s === 'CheckedIn' ? ' (已簽到)' : (s === 'Absent' ? ' (缺席)' : ''));
    const slotId = p.slot_id || p.SlotID;
    const bookingTime = p.booking_time || p.BookingTime;
    const bookingId = p.booking_id || p.BookingID;
    const escapedName = n.replace(/'/g, "\\'");
    return `<button class="text-xs px-2 py-1 ${classes}" data-status="${s}" data-name="${escapedName}" onclick="handleMyBookingTagClick(this, '${escapedName}', this.dataset.status, '${bookingTime}', '${bookingId}', '${slotId}')">${n}${suffix}</button>`;
};

const jsRenderSlotContent = (s) => {
    const tags = (s.attendees || []).map(p => jsRenderTag(p, true)).join('');
    return ("<div class=\"flex flex-col leading-tight\"><span class=\"text-xs text-[#8E8E93] font-mono\">" + s.time_display + "</span><span class=\"text-xs font-bold text-white break-all leading-tight line-clamp-3 overflow-hidden\">" + s.course_name + "</span></div><div class=\"flex flex-col gap-1 mt-1\">" + tags + "</div>");
};

async function refreshSlot(slotId) {
    if (!slotId) return;
    try {
        const s = await fetchApi("/api/v2/calendar/slots/" + slotId);
        const btn = document.getElementById('slot-' + slotId);
        if (btn) {
            btn.innerHTML = jsRenderSlotContent(s);
            if (s.is_past) {
                btn.className = "w-full rounded-[8px] py-2 px-1.5 flex flex-col gap-1 text-left transition-all bg-[#1C1C1E] opacity-50 cursor-not-allowed";
                btn.onclick = null;
            } else {
                btn.className = "w-full rounded-[8px] py-2 px-1.5 flex flex-col gap-1 text-left transition-all bg-[#1C1C1E] active:scale-95";
                btn.onclick = () => {
                    openBookingPopup(s.id, s.time_display, s.end_time_display, s.course_name, s.booked_count, s.capacity, s.attendees);
                };
            }
            btn.classList.add('ring-2', 'ring-[#FFD700]', 'ring-inset');
            setTimeout(() => {
                btn.classList.remove('ring-2', 'ring-[#FFD700]', 'ring-inset');
            }, 1000);
        }
        const popupSlotId = document.getElementById('popup-slot-id');
        if (popupSlotId && popupSlotId.value === slotId) {
            const countEl = document.getElementById('popup-booked-count');
            if (countEl) countEl.innerText = s.booked_count;
        }
    } catch (e) { console.error("Refresh slot failed:", e); }
}

function refreshStats() {
    if (currentYear && currentMonth) fetchMonthlyStats(currentYear, currentMonth);
}

let currentYear, currentMonth, isLoading = false, isEndTop = false, isEndBottom = false, firstLoadedDate, lastLoadedDate;

const updateHeaderTitle = () => {
    const c = document.getElementById('calendar-container'), t = document.getElementById('current-month-title');
    if (!c || !t) return;
    const weeks = c.querySelectorAll('[data-week-id]'), cTop = c.getBoundingClientRect().top;
    for (let w of weeks) {
        if (w.getBoundingClientRect().bottom > cTop + 100) {
            const d = w.querySelector('[data-date]');
            if (d) {
                const parts = d.dataset.date.split('-'), y = parts[0], m = parts[1], mi = parseInt(m, 10);
                if (y !== currentYear || mi !== currentMonth) { 
                    currentYear = y; currentMonth = mi; t.innerText = y + "年" + mi + "月"; 
                    fetchMonthlyStats(y, mi); 
                }
            }
            break;
        }
    }
};

async function fetchMonthlyStats(y, m) {
    const label = document.getElementById('stats-label'); if (label) label.innerText = y + " 年" + m + "月統計:";
    try {
        const stats = await fetchApi("/api/v2/calendar/stats?year=" + y + "&month=" + m);
        const body = document.getElementById('stats-children-body');
        if (body) body.innerHTML = stats.children.map(c => ("<tr class=\"border-b border-[#27272A]/50 last:border-0\"><td class=\"py-2 text-left\">" + c.name + "</td><td class=\"py-2 text-[#60A5FA]\">" + c.upcoming + "</td><td class=\"py-2 text-[#34D399]\">" + c.completed + "</td><td class=\"py-2 text-[#8E8E93]\">" + c.leave + "</td><td class=\"py-2 " + (c.absent > 0 ? 'text-[#F87171]' : 'text-[#8E8E93]') + "\">" + c.absent + "</td><td class=\"py-2 text-[#8E8E93]\">" + c.avg_week.toFixed(1) + "</td></tr>")).join('');
        const tu = document.getElementById('total-upcoming'); if (tu) tu.innerText = stats.total_upcoming + " 預約";
        const ts = document.getElementById('total-sessions'); if (ts) ts.innerText = stats.total_sessions + " 堂";
        const tl = document.getElementById('total-leave'); if (tl) tl.innerText = stats.total_leave;

        const frequentList = document.getElementById('frequent-names-list');
        if (frequentList && stats.children) {
            frequentList.innerHTML = stats.children.map(c => {
                const escapedName = c.name.replace(/'/g, "\\'");
                return `<button onclick="addDraftTag('${escapedName}')" class="px-3 py-1.5 rounded-full bg-[#27272A] text-white text-sm border border-[#3A3A3C] hover:bg-[#3A3A3C]">${c.name}</button>`;
            }).join('');
        }
    } catch (e) { console.error(e); }
}

const renderWeekRow = (w) => {
    const days = w.days.map(d => {
        const slots = d.slots.map(s => {
            if (s.is_empty) return ("<div class=\"w-full rounded-[8px] py-2 px-1.5 border border-dashed border-[#3A3A3C] flex items-center justify-center text-left min-h-[40px] opacity-50 cursor-default\"><span class=\"text-xs text-[#8E8E93]\">[未排課]</span></div>");
            const state = s.is_past ? 'opacity-50 cursor-not-allowed' : 'active:scale-95';
            const onclick = s.is_past ? '' : ("onclick=\"openBookingPopup('" + s.id + "', '" + s.time_display + "', '" + s.end_time_display + "', '" + s.course_name + "', " + s.booked_count + ", " + s.capacity + ", " + JSON.stringify(s.attendees).replace(/\"/g, '&quot;') + ")\"");
            return ("<button id=\"slot-" + s.id + "\" class=\"w-full rounded-[8px] py-2 px-1.5 flex flex-col gap-1 text-left transition-transform bg-[#1C1C1E] " + state + "\" " + onclick + ">" + jsRenderSlotContent(s) + "</button>");
        }).join('');
        return ("<div class=\"flex flex-col items-center gap-2 min-h-[120px]\" data-date=\"" + d.full_date + "\"><div class=\"flex flex-col items-center justify-center w-10 h-10 rounded-full text-sm font-medium " + (d.is_today ? "bg-white text-black" : "text-[#8E8E93]") + "\"><span class=\"text-xs uppercase leading-none\">" + d.day_of_week + "</span><span class=\"text-lg leading-none font-bold\">" + d.date_display + "</span></div><div class=\"w-full flex flex-col gap-2 px-0\">" + slots + "</div></div>");
    }).join('');
    return ("<div class=\"snap-start pt-4 pb-2 border-b border-[#27272A] min-h-[50vh]\" data-week-id=\"" + w.id + "\"><div class=\"grid grid-cols-7 gap-[2px] px-1\">" + days + "</div></div>");
};

async function loadWeeks(dir) {
    if (isLoading || (dir === 'prev' && isEndTop) || (dir === 'next' && isEndBottom)) return;
    const c = document.getElementById('calendar-container'), ref = dir === 'next' ? lastLoadedDate : firstLoadedDate;
    if (!ref) return; isLoading = true;
    try {
        const data = await fetchApi("/api/v2/calendar/weeks?start_date=" + ref + "&direction=" + dir);
        if (data.weeks && data.weeks.length) {
            const oldH = c.scrollHeight, oldT = c.scrollTop;
            if (dir === 'prev') {
                c.style.scrollBehavior = 'auto';
                c.style.scrollSnapType = 'none';
            }
            const html = data.weeks.map(w => renderWeekRow(w)).join(''), topS = document.getElementById('sentinel-top'), botS = document.getElementById('sentinel-bottom');
            const temp = document.createElement('div'); temp.innerHTML = html; const nodes = Array.from(temp.children);
            if (dir === 'next') { 
                nodes.forEach(n => c.insertBefore(n, botS)); 
                lastLoadedDate = data.weeks[data.weeks.length-1].days[6].full_date; 
            } else { 
                nodes.reverse().forEach(n => c.insertBefore(n, topS.nextSibling)); 
                c.scrollTop = c.scrollHeight - oldH + oldT;
                firstLoadedDate = data.weeks[0].days[0].full_date; 
            }
            if (dir === 'prev') {
                requestAnimationFrame(() => {
                    c.style.scrollBehavior = '';
                    c.style.scrollSnapType = '';
                });
            }
        } else {
            if (dir === 'next') {
                isEndBottom = true;
                showToast({ title: "提示", description: "最多只能查詢半年內", variant: "default" });
            } else {
                isEndTop = true;
                showToast({ title: "提示", description: "最多只能查詢三個月前", variant: "default" });
            }
        }
    } catch (e) { console.error(e); }
    finally { updateHeaderTitle(); setTimeout(() => { isLoading = false; checkSentinels(); }, 500); }
}

const checkSentinels = () => {
    const c = document.getElementById('calendar-container'), cRect = c.getBoundingClientRect();
    const isV = el => { if (!el) return false; const r = el.getBoundingClientRect(); return r.bottom >= cRect.top - 10 && r.top <= cRect.bottom + 10; };
    if (!isLoading) {
        if (!isEndBottom && isV(document.getElementById('sentinel-bottom'))) loadWeeks('next');
        else if (!isEndTop && isV(document.getElementById('sentinel-top'))) loadWeeks('prev');
    }
};

document.addEventListener('DOMContentLoaded', () => {
    const days = document.querySelectorAll('[data-date]');
    if (days.length) { firstLoadedDate = days[0].dataset.date; lastLoadedDate = days[days.length-1].dataset.date; }
    const c = document.getElementById('calendar-container'), today = c.querySelector('.bg-white.text-black');
    if (today) today.parentElement.scrollIntoView({ block: 'center' });
    c.addEventListener('scroll', () => { updateHeaderTitle(); checkSentinels(); });
    updateHeaderTitle();

    const configEl = document.getElementById('liff-config');
    const liffId = configEl ? configEl.getAttribute('data-liff-id') : null;
    if (liffId && typeof liff !== 'undefined') {
        liff.init({ liffId: liffId }).catch(err => console.error("LIFF init failed", err));
    }
});

function handleSmartInputKeydown(e) {
    if (e.key === 'Enter' || e.key === ',') {
        e.preventDefault();
        const val = e.target.value.trim().replace(/,/g, '');
        if (val) {
            addDraftTag(val);
            e.target.value = '';
        }
    } else if (e.key === 'Backspace' && !e.target.value) {
        const tags = document.querySelectorAll('#smart-input-container [data-draft="true"]');
        if (tags.length) {
            tags[tags.length - 1].remove();
        }
    }
}

function openBookingPopup(id, time, endTime, title, bookedCount, capacity, attendees) {
    window.currentIdempotencyKey = self.crypto.randomUUID();
    document.getElementById('popup-course-title').innerText = title; 
    document.getElementById('popup-time-info').innerText = time + ' ~ ' + endTime; 
    document.getElementById('popup-slot-id').value = id;
    document.getElementById('popup-booked-count').innerText = bookedCount;
    document.getElementById('popup-capacity').innerText = capacity;
    const bookedList = document.getElementById('booked-participants-list'), draftC = document.getElementById('smart-input-container'), input = document.getElementById('smart-input');
    bookedList.innerHTML = ''; input.value = ''; Array.from(draftC.children).forEach(c => c !== input && c.remove());
    if (attendees && attendees.length) { document.getElementById('booked-list-wrapper').classList.remove('hidden'); attendees.forEach(p => addBookedTag(p.name || p.Name, p.status || p.Status, p.booking_time || p.BookingTime, p.booking_id || p.BookingID)); }
    else document.getElementById('booked-list-wrapper').classList.add('hidden');
    
    document.getElementById('booking-popup').classList.remove('hidden');
}

function closeBookingPopup() { 
    window.currentIdempotencyKey = null;
    document.getElementById('booking-popup').classList.add('hidden'); 
    
    // Reset views
    document.getElementById('booking-leave-view').classList.add('hidden'); 
    document.getElementById('booking-main-view').classList.remove('hidden'); 
    
    // Hide all inline confirmation panels
    const panels = document.querySelectorAll('[id$="-confirm-panel"]');
    panels.forEach(p => {
        p.classList.add('translate-y-full');
        setTimeout(() => p.classList.add('hidden'), 300);
    });

    const leaveInput = document.getElementById('leaveReason');
    if (leaveInput) leaveInput.value = '';
}
function openMyBookings() { document.getElementById('my-bookings-modal').classList.remove('hidden'); switchMyBookingsTab('upcoming'); }
function closeMyBookings() { document.getElementById('my-bookings-modal').classList.add('hidden'); }

const addBookedTag = (n, s, t, id) => {
    const tag = document.createElement('div'); tag.className = "flex items-center gap-1 px-3 py-1 rounded-full cursor-pointer transition-all " + getStatusClasses(s);
    tag.innerHTML = ("<span>" + n + "</span>"); 
    tag.dataset.status = s;
    tag.onclick = () => handleTagAction(tag, "booking", n, tag.dataset.status, t, id, (ns) => { 
        if (ns === "Remove") tag.remove(); 
        else {
            tag.dataset.status = ns;
            tag.className = "flex items-center gap-1 px-3 py-1 rounded-full cursor-pointer transition-all " + getStatusClasses(ns); 
        }
    });
    document.getElementById('booked-participants-list').appendChild(tag); document.getElementById('booked-list-wrapper').classList.remove('hidden');
};

let currentTab = 'upcoming';
async function switchMyBookingsTab(type) {
    currentTab = type; const u = document.getElementById('tab-upcoming'), h = document.getElementById('tab-history'), l = document.getElementById('my-bookings-list');
    const active = "pb-2 text-[#FFD700] border-b-2 border-[#FFD700] font-medium", inactive = "pb-2 text-[#8E8E93] font-medium hover:text-white transition-colors border-b-2 border-transparent";
    u.className = type === 'upcoming' ? active : inactive; h.className = type === 'history' ? active : inactive;
    l.innerHTML = '<div class="text-center text-[#8E8E93] py-8">載入中...</div>';
    try {
        const data = await fetchApi("/api/v2/my-bookings?type=" + type);
        l.innerHTML = data.items.length ? data.items.map(item => ("<div class=\"bg-[#000000] p-3 rounded-lg border border-[#27272A]\"><div class=\"mb-2\"><div class=\"text-white font-bold\">" + item.date_display + "</div><div class=\"text-xs text-[#8E8E93]\">" + item.title + "</div></div><div class=\"flex flex-wrap gap-2\">" + item.attendees.map(p => jsRenderTag(p, false)).join('') + "</div></div>")).join('') : '<div class="text-center text-[#8E8E93] py-8">無預約紀錄</div>';
    } catch(e) { l.innerHTML = '<div class="text-center text-[#F87171] py-8">載入失敗</div>'; }
}

async function handleMyBookingTagClick(btn, n, s, t, id, slotId) { 
    handleTagAction(btn, "my-booking", n, btn.dataset.status || s, t, id, (ns) => { 
        if (ns === "Remove") btn.remove(); 
        else {
            btn.dataset.status = ns;
            switchMyBookingsTab(currentTab); 
        }
    }, slotId); 
}
async function handleTagAction(btn, prefix, n, s, t, id, callback, slotId) {
    const currentSlotId = slotId || document.getElementById('popup-slot-id').value;
    
    const executeAction = async (url, method, successStatus) => {
        const originalHtml = btn.innerHTML;
        const isButton = btn.tagName === 'BUTTON';
        
        try {
            if (isButton) btn.disabled = true;
            btn.innerHTML = "處理中...";
            btn.classList.add('opacity-50', 'cursor-not-allowed', 'pointer-events-none');
            
            const opts = { method: method, headers: {} };
            if (window.currentIdempotencyKey) {
                opts.headers['Idempotency-Key'] = window.currentIdempotencyKey;
            }

            await fetchApi(url, opts);
            if (successStatus === "Booked") {
                btn.innerHTML = "<span>" + n + "</span>"; 
            } else {
                btn.innerHTML = originalHtml;
            }
            
            callback(successStatus);
            refreshSlot(currentSlotId);
            refreshStats();
        } catch(e) {
            showToast({ title: "錯誤", description: e.message, variant: "destructive" });
            btn.innerHTML = originalHtml;
        } finally {
            if (isButton) btn.disabled = false;
            btn.classList.remove('opacity-50', 'cursor-not-allowed', 'pointer-events-none');
        }
    };

    if (s === "Booked") {
        const diff = (new Date() - new Date(t)) / 36e5;
        if (diff < 24) {
            if (await showInlineConfirm(prefix, "取消預約", "建立未滿 24 小時，確定取消？")) {
                window.currentIdempotencyKey = self.crypto.randomUUID();
                await executeAction("/api/v2/bookings/" + id, 'DELETE', "Remove");
            }
        } else {
            if (prefix === 'my-booking') closeMyBookings();
            openLeaveRequest(id, n, currentSlotId);
        }
    } else if (s === "Leave") {
        if (await showInlineConfirm(prefix, "恢復預約", "取消 " + n + " 的請假？")) {
            window.currentIdempotencyKey = self.crypto.randomUUID();
            await executeAction("/api/v2/bookings/" + id + "/leave", 'DELETE', "Booked");
        }
    }
}

function showInlineConfirm(prefix, title, msg) {
    return new Promise(resolve => {
        const p = document.getElementById(prefix + '-confirm-panel'), t = document.getElementById(prefix + '-confirm-title'), m = document.getElementById(prefix + '-confirm-msg');
        const c = document.getElementById(prefix + '-confirm-cancel'), o = document.getElementById(prefix + '-confirm-ok');
        if (!p) return resolve(false);
        t.innerText = title; m.innerText = msg; p.classList.remove('hidden'); requestAnimationFrame(() => p.classList.remove('translate-y-full'));
        const done = r => { p.classList.add('translate-y-full'); setTimeout(() => p.classList.add('hidden'), 300); resolve(r); };
        c.onclick = () => done(false); o.onclick = () => done(true);
    });
}

function addDraftTag(name) {
    const input = document.getElementById('smart-input'), container = document.getElementById('smart-input-container');
    if ([...container.querySelectorAll('[data-name]')].some(t => t.dataset.name === name)) return;
    const tag = document.createElement('div'); tag.className = "flex items-center gap-1 px-3 py-1 rounded-full bg-[#27272A] text-white text-sm cursor-pointer border border-[#3A3A3C]";
    tag.innerHTML = ("<span>" + name + "</span><span class=\"text-[#8E8E93] text-xs ml-1\">×</span>");
    tag.dataset.name = name; tag.dataset.draft = "true"; tag.onclick = () => tag.remove();
    container.insertBefore(tag, input);
}

let isSubmitting = false;
const submitBooking = async () => {
    if (isSubmitting) return;
    
    const id = document.getElementById('popup-slot-id').value;
    const input = document.getElementById('smart-input');
    const tags = document.querySelectorAll('[data-draft="true"]');
    let names = [...tags].map(t => t.dataset.name);
    
    const pendingName = input.value.trim().replace(/,/g, '');
    if (pendingName && !names.includes(pendingName)) {
        names.push(pendingName);
    }

    if (!names.length) { closeBookingPopup(); return; }

    const submitBtn = document.querySelector('#booking-main-view button[onclick="submitBooking()"]');
    const originalText = submitBtn.innerText;
    
    isSubmitting = true;
    submitBtn.disabled = true;
    submitBtn.innerText = "處理中...";
    submitBtn.classList.add('opacity-50', 'cursor-not-allowed');

    try {
        const opts = { 
            method: 'POST', 
            body: JSON.stringify({ slot_id: id, student_names: names }),
            headers: {} 
        };
        if (window.currentIdempotencyKey) {
            opts.headers['Idempotency-Key'] = window.currentIdempotencyKey;
        }

        const res = await fetchApi('/api/v2/bookings', opts);
        res.new_bookings.forEach(b => addBookedTag(b.name, b.status, b.booking_time, b.booking_id));
        tags.forEach(t => t.remove()); 
        input.value = ''; 
        showToast({ title: "成功", description: "預約成功！", variant: "default" });
        closeBookingPopup(); 
        refreshSlot(id); 
        refreshStats();
    } catch(e) { 
        showToast({ title: "預約失敗", description: e.message, variant: "destructive" });
    } finally {
        isSubmitting = false;
        submitBtn.disabled = false;
        submitBtn.innerText = originalText;
        submitBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    }
};

function openLeaveRequest(id, name, slotId) { 
    window.currentIdempotencyKey = self.crypto.randomUUID();
    document.getElementById('leave-booking-id').value = id; 
    document.getElementById('leave-student-name').innerText = name; 
    if (slotId) document.getElementById('popup-slot-id').value = slotId;
    document.getElementById('booking-main-view').classList.add('hidden'); 
    document.getElementById('booking-leave-view').classList.remove('hidden');
    document.getElementById('booking-popup').classList.remove('hidden');
}
function cancelLeaveRequest() { 
    window.currentIdempotencyKey = null;
    document.getElementById('booking-leave-view').classList.add('hidden'); 
    document.getElementById('booking-main-view').classList.remove('hidden'); 
}

async function submitLeaveRequest(e) { 
    e.preventDefault(); 
    if (isSubmitting) return;

    const slotId = document.getElementById('popup-slot-id').value;
    const submitBtn = e.target.querySelector('button[type="submit"]');
    const originalText = submitBtn.innerText;

    isSubmitting = true;
    submitBtn.disabled = true;
    submitBtn.innerText = "處理中...";
    submitBtn.classList.add('opacity-50', 'cursor-not-allowed');

    try { 
        const opts = { 
            method: 'POST', 
            body: JSON.stringify({ reason: document.getElementById('leaveReason').value }),
            headers: {}
        };
        if (window.currentIdempotencyKey) {
            opts.headers['Idempotency-Key'] = window.currentIdempotencyKey;
        }

        await fetchApi("/api/v2/bookings/" + document.getElementById('leave-booking-id').value + "/leave", opts); 
        showToast({ title: "成功", description: "請假申請已送出", variant: "default" });
        cancelLeaveRequest(); 
        closeBookingPopup(); 
        await refreshSlot(slotId); 
        refreshStats(); 
    } catch(err) { 
        showToast({ title: "申請失敗", description: err.message, variant: "destructive" });
    } finally {
        isSubmitting = false;
        submitBtn.disabled = false;
        submitBtn.innerText = originalText;
        submitBtn.classList.remove('opacity-50', 'cursor-not-allowed');
    }
}

async function shareBookingStatus() {
    const btn = document.getElementById('share-booking-btn');
    const originalContent = btn.innerHTML;
    try {
        if (!liff.isInClient()) {
            showToast({ title: "提示", description: "請在 LINE 中執行", variant: "default" });
            return;
        }
        btn.disabled = true;
        btn.innerHTML = '<span>正在準備...</span>';
        const data = await fetchApi('/booking/summary/line-liff');
        if (data.message) {
            await liff.sendMessages([{ type: 'text', text: data.message }]);
            liff.closeWindow();
        } else {
            throw new Error("無法取得分享訊息");
        }
    } catch (e) {
        console.error(e);
        showToast({ title: "分享失敗", description: e.message || "發生錯誤", variant: "destructive" });
        btn.disabled = false;
        btn.innerHTML = originalContent;
    }
}
