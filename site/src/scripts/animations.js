/**
 * Bombardment Website Animation Scripts
 * Contains all animation-related scripts and interactive functionality
 */

// Animation constants for typing effect
const ANIMATION = {
    TYPING: {
        SPEED: 40,           // Speed for typing characters (ms)
        DELETE_SPEED: 20,    // Speed for deleting characters (ms)
        PAUSE_DURATION: 1200, // How long to pause at the end of a phrase (ms) 
        INITIAL_DELAY: 300   // Delay before starting the animation (ms)
    }
};

// Subtitle typing animation
const subtitles = [
    "No more random last-minute scripts",
    "Switch to no-code, performant, data migrations",
    "Transform your data migration process",
    "Fast, lightweight and scalable data processing",
    "Process data in batches concurrently",
    "Client-side load balancing built-in",
    "Modular design for easy extension",
    "Multiple channels for target systems",
    "Golang powered: for maximum performance"
];

// State variables for typing animation
let currentSubtitleIndex = 0;
let currentCharIndex = 0;
let isDeleting = false;
let typingSpeed = ANIMATION.TYPING.SPEED;

function typeSubtitle() {
    const subtitleElement = document.querySelector('.typing-animation');
    if (!subtitleElement) return;

    const currentText = subtitles[currentSubtitleIndex];
    
    if (isDeleting) {
        // Deleting text
        subtitleElement.textContent = currentText.substring(0, currentCharIndex - 1);
        currentCharIndex--;
        typingSpeed = ANIMATION.TYPING.DELETE_SPEED;
    } else {
        // Typing text
        subtitleElement.textContent = currentText.substring(0, currentCharIndex + 1);
        currentCharIndex++;
        typingSpeed = ANIMATION.TYPING.SPEED;
    }
    
    // Check if completed typing the current subtitle
    if (!isDeleting && currentCharIndex === currentText.length) {
        // Pause at the end of typing before starting to delete
        isDeleting = true;
        typingSpeed = ANIMATION.TYPING.PAUSE_DURATION;
    } else if (isDeleting && currentCharIndex === 1) {
        // Move to the next subtitle
        isDeleting = false;
        currentSubtitleIndex = (currentSubtitleIndex + 1) % subtitles.length;
    }
    
    setTimeout(typeSubtitle, typingSpeed);
}

// Copy to clipboard functionality
function copyToClipboard(elementId) {
    const element = document.getElementById(elementId);
    const text = element.textContent;

    navigator.clipboard.writeText(text).then(() => {
        const button = element.parentElement.querySelector('.copy-button');
        const originalIcon = button.innerHTML;
        button.innerHTML = '<i class="fa-solid fa-check"></i> Copied!';

        setTimeout(() => {
            button.innerHTML = originalIcon;
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy text: ', err);
        alert('Failed to copy text. Please try again.');
    });
}

// Setup tab switching functionality
function setupTabSwitching() {
    document.querySelectorAll('.code-tab').forEach(tab => {
        tab.addEventListener('click', () => {
            const tabId = tab.getAttribute('data-tab');

            // Deactivate all tabs
            document.querySelectorAll('.code-tab').forEach(t => {
                t.classList.remove('active');
            });
            document.querySelectorAll('.code-content').forEach(c => {
                c.classList.remove('active');
            });

            // Activate the clicked tab
            tab.classList.add('active');
            document.getElementById(tabId).classList.add('active');

            // Rehighlight the code
            Prism.highlightAll();
        });
    });
}

// Setup fade-in animations using Intersection Observer
function setupFadeInAnimations() {
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('visible');
                observer.unobserve(entry.target);
            }
        });
    }, {
        root: null,
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    });

    document.querySelectorAll('.fade-in').forEach(element => {
        observer.observe(element);
    });
}

// Initialize all functionality when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Start the typing animation with a small delay
    setTimeout(typeSubtitle, 500);
    
    // Setup tab switching if tabs exist
    setupTabSwitching();
    
    // Setup fade-in animations
    setupFadeInAnimations();
    
    // Initialize Prism.js syntax highlighting
    if (typeof Prism !== 'undefined') {
        Prism.highlightAll();
    }
});
