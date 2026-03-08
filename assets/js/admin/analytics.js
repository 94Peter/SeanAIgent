// Business Analytics Chart Initialization

document.addEventListener('DOMContentLoaded', function() {
    const canvas = document.getElementById('historicalChart');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    const stats = JSON.parse(canvas.getAttribute('data-stats') || '[]');
    
    const labels = stats.map(s => s.month);
    const bookedData = stats.map(s => s.bookedCount);
    const attendedData = stats.map(s => s.attendedCount);

    if (typeof Chart === 'undefined') {
        console.error('Chart.js is not loaded');
        return;
    }

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: '預約數',
                    data: bookedData,
                    borderColor: '#60A5FA',
                    backgroundColor: 'rgba(96, 165, 250, 0.1)',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true
                },
                {
                    label: '簽到數',
                    data: attendedData,
                    borderColor: '#34D399',
                    backgroundColor: 'rgba(52, 211, 153, 0.1)',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false,
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: '#27272A'
                    },
                    ticks: {
                        color: '#8E8E93'
                    }
                },
                x: {
                    grid: {
                        display: false
                    },
                    ticks: {
                        color: '#8E8E93'
                    }
                }
            },
            plugins: {
                legend: {
                    position: 'top',
                    labels: {
                        color: '#FFFFFF',
                        boxWidth: 12,
                        font: { size: 12 }
                    }
                }
            }
        }
    });
});
