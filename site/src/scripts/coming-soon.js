/**
 * Coming Soon Page - Countdown Timer and Calendar Events
 */

// Constants
const LAUNCH_DATE = new Date('2025-04-23T09:00:00').getTime();
const LAUNCH_END_TIME = '2025-04-23T10:00:00';
const EVENT_TITLE = 'You Got This Extension Launch';
const EVENT_DESCRIPTION = 'The You Got This productivity browser extension is launching! Download it from our website: https://bombardment.work';
const EVENT_LOCATION = 'https://bombardment.work';
const UPDATE_INTERVAL = 1000; // Update countdown every second

const updateTime = () => {
    const now = new Date().getTime();
    const timeRemaining = LAUNCH_DATE - now;

    if (timeRemaining < 0) {
        clearInterval(countdown);
        displayCountdownValues(0, 0, 0, 0);
        return;
    }

    // Calculate time components
    const days = Math.floor(timeRemaining / (1000 * 60 * 60 * 24));
    const hours = Math.floor((timeRemaining % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((timeRemaining % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((timeRemaining % (1000 * 60)) / 1000);

    displayCountdownValues(days, hours, minutes, seconds);
}

/**
 * Initialize the countdown timer
 */
function initCountdown() {
    const countdown = setInterval(updateTime, UPDATE_INTERVAL);
    updateTime(); // Initial call to display countdown immediately
}

/**
 * Display countdown values in the DOM
 */
function displayCountdownValues(days, hours, minutes, seconds) {
    document.getElementById('days').innerHTML = days.toString().padStart(2, '0');
    document.getElementById('hours').innerHTML = hours.toString().padStart(2, '0');
    document.getElementById('minutes').innerHTML = minutes.toString().padStart(2, '0');
    document.getElementById('seconds').innerHTML = seconds.toString().padStart(2, '0');
}

/**
 * Add Google Calendar integration
 */
function initGoogleCalendar() {
    const googleCalendarBtn = document.getElementById('google-calendar');
    if (!googleCalendarBtn) return;

    googleCalendarBtn.addEventListener('click', function (e) {
        e.preventDefault();

        // Format dates for Google Calendar (YYYYMMDDTHHMMSS format)
        const startDate = formatDateForGoogleCalendar(LAUNCH_DATE);
        const endDate = formatDateForGoogleCalendar(LAUNCH_END_TIME);

        const googleCalendarUrl = 'https://www.google.com/calendar/render?' +
            'action=TEMPLATE' +
            '&text=' + encodeURIComponent(EVENT_TITLE) +
            '&details=' + encodeURIComponent(EVENT_DESCRIPTION) +
            '&location=' + encodeURIComponent(EVENT_LOCATION) +
            '&dates=' + startDate + '/' + endDate;

        window.open(googleCalendarUrl, '_blank');
    });
}

/**
 * Format date for Google Calendar URL
 */
function formatDateForGoogleCalendar(dateString) {
    const date = new Date(dateString);
    return date.getFullYear() +
        ('0' + (date.getMonth() + 1)).slice(-2) +
        ('0' + date.getDate()).slice(-2) +
        'T' +
        ('0' + date.getHours()).slice(-2) +
        ('0' + date.getMinutes()).slice(-2) +
        ('0' + date.getSeconds()).slice(-2);
}

/**
 * Initialize ICS file download functionality
 */
function initIcsDownload() {
    const icsDownloadBtn = document.getElementById('ics-download');
    if (!icsDownloadBtn) return;

    icsDownloadBtn.addEventListener('click', function (e) {
        e.preventDefault();

        // Create ICS calendar file content
        const icsContent = generateIcsContent();

        // Create and trigger download
        downloadIcsFile(icsContent);
    });
}

/**
 * Generate ICS file content
 */
function generateIcsContent() {
    // Format dates for ICS file (YYYYMMDDTHHMMSS format, no separators)
    const startDate = formatDateForIcs(LAUNCH_DATE);
    const endDate = formatDateForIcs(LAUNCH_END_TIME);

    return 'BEGIN:VCALENDAR\n' +
        'VERSION:2.0\n' +
        'PRODID:-//You Got This//Extension Launch//EN\n' +
        'CALSCALE:GREGORIAN\n' +
        'BEGIN:VEVENT\n' +
        `DTSTART:${startDate}\n` +
        `DTEND:${endDate}\n` +
        `SUMMARY:${EVENT_TITLE}\n` +
        `DESCRIPTION:${EVENT_DESCRIPTION}\n` +
        `LOCATION:${EVENT_LOCATION}\n` +
        'STATUS:CONFIRMED\n' +
        'SEQUENCE:0\n' +
        'BEGIN:VALARM\n' +
        'TRIGGER:-PT24H\n' +
        'ACTION:DISPLAY\n' +
        'DESCRIPTION:Reminder\n' +
        'END:VALARM\n' +
        'END:VEVENT\n' +
        'END:VCALENDAR';
}

/**
 * Format date for ICS file
 */
function formatDateForIcs(dateString) {
    const date = new Date(dateString);
    return date.getFullYear() +
        ('0' + (date.getMonth() + 1)).slice(-2) +
        ('0' + date.getDate()).slice(-2) +
        'T' +
        ('0' + date.getHours()).slice(-2) +
        ('0' + date.getMinutes()).slice(-2) +
        ('0' + date.getSeconds()).slice(-2);
}

/**
 * Download ICS file
 */
function downloadIcsFile(icsContent) {
    const element = document.createElement('a');
    element.setAttribute('href', 'data:text/calendar;charset=utf-8,' + encodeURIComponent(icsContent));
    element.setAttribute('download', 'you_got_this_launch.ics');

    element.style.display = 'none';
    document.body.appendChild(element);

    element.click();

    document.body.removeChild(element);
}

/**
 * Initialize email notification form
 */
function initNotificationForm() {
    const form = document.querySelector('.email-form');
    if (!form) return;

    form.addEventListener('submit', (e) => {
        e.preventDefault();

        // In a real implementation, you'd send this data to your backend
        const email = form.querySelector('.email-input').value;

        // Show success message
        showFormSuccessMessage(form);
    });
}

/**
 * Show success message after form submission
 */
function showFormSuccessMessage(form) {
    form.innerHTML = '<p style="text-align: center; font-weight: 500; color: var(--color-success);">Thanks for signing up! We\'ll notify you when we launch.</p>';
}

/**
 * Initialize all coming soon page functionality
 */
function initComingSoonPage() {
    initCountdown();
    initGoogleCalendar();
    initIcsDownload();
    initNotificationForm();
}

// Initialize when DOM is fully loaded
document.addEventListener('DOMContentLoaded', initComingSoonPage);