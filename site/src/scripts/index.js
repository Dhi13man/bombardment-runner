/**
 * You Got This - Main JavaScript
 * Handles all interactive elements on the landing page
 */

// Constants
const ROTATION_INTERVAL = 2500; // ms between auto-rotations
const ANIMATION_DURATION = 500; // ms for fade animations
const PARALLAX_FACTOR = 0.2; // parallax movement factor
const SCROLL_THRESHOLD = 600; // max scroll for parallax effect

/**
 * Screenshot tab carousel functionality
 */
class ScreenshotCarousel {
    constructor() {
        this.tabs = document.querySelectorAll('.screenshot-tab');
        this.contents = document.querySelectorAll('.screenshot-content');
        this.currentIndex = 0;
        this.interval = null;
        this.section = document.querySelector('.screenshots');
        this.isTransitioning = false;

        // Preload all images to prevent lag during transitions
        this.preloadImages();
        this.init();
    }

    init() {
        // Set up tab click events
        this.tabs.forEach((tab, index) => {
            tab.addEventListener('click', () => {
                if (!this.isTransitioning) {
                    this.switchToTab(index);
                }
            });
        });

        // Pause rotation on user interaction
        if (this.section) {
            this.section.addEventListener('mouseenter', () => this.stopRotation());
            this.section.addEventListener('mouseleave', () => this.startRotation());
            this.section.addEventListener('touchstart', () => this.stopRotation(), { passive: true });
            this.section.addEventListener('touchend', () => this.startRotation());
        }
    }

    preloadImages() {
        // Find all screenshot images and preload them
        const images = document.querySelectorAll('.screenshot-box img');
        images.forEach(img => {
            const src = img.getAttribute('src');
            if (src) {
                const preloadImg = new Image();
                preloadImg.src = src;
            }
        });
    }

    switchToTab(index) {
        if (index < 0 || index >= this.tabs.length || this.isTransitioning) return;

        this.isTransitioning = true;

        // Remove active class from all tabs
        this.tabs.forEach(tab => tab.classList.remove('active'));

        // Add active class to selected tab immediately
        this.tabs[index].classList.add('active');

        // Create a fade-out effect for the current content
        const currentContent = document.querySelector('.screenshot-content.active');
        if (currentContent) {
            currentContent.style.animation = 'none'; // Reset animation
            void currentContent.offsetWidth; // Trigger reflow
            currentContent.style.animation = 'fadeZoomOut 0.3s ease forwards';

            setTimeout(() => {
                // Remove active class from all contents
                this.contents.forEach(content => {
                    content.classList.remove('active');
                    content.style.animation = ''; // Reset animation
                });

                // Add active class to selected content with fresh animation
                this.contents[index].classList.add('active');
                this.contents[index].style.animation = 'fadeZoomIn 0.6s ease';

                this.currentIndex = index;
                this.isTransitioning = false;
            }, 300); // Match the fadeOut duration
        } else {
            this.contents.forEach(content => content.classList.remove('active'));
            this.contents[index].classList.add('active');
            this.currentIndex = index;
            this.isTransitioning = false;
        }

        // Reset the timer when a user manually clicks
        this.resetRotation();
    }

    startRotation() {
        if (this.interval) return;

        this.interval = setInterval(() => {
            if (!this.isTransitioning) {
                const nextIndex = (this.currentIndex + 1) % this.tabs.length;
                this.switchToTab(nextIndex);
            }
        }, ROTATION_INTERVAL);
    }

    stopRotation() {
        if (this.interval) {
            clearInterval(this.interval);
            this.interval = null;
        }
    }

    resetRotation() {
        this.stopRotation();
        this.startRotation();
    }
}

/**
 * Browser tab switcher
 */
class BrowserTabSwitcher {
    constructor() {
        this.tabs = document.querySelectorAll('.browser-tab');
        this.contents = document.querySelectorAll('.browser-content');

        this.init();
    }

    init() {
        this.tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                // Get the browser ID from data attribute
                const browserId = tab.getAttribute('data-browser');
                this.switchTab(browserId);
            });
        });
    }

    switchTab(browserId) {
        // Remove active class from all tabs
        this.tabs.forEach(tab => tab.classList.remove('active'));

        // Add active class to clicked tab
        const activeTab = document.querySelector(`.browser-tab[data-browser="${browserId}"]`);
        if (activeTab) activeTab.classList.add('active');

        // Hide all content
        this.contents.forEach(content => content.classList.remove('active'));

        // Show corresponding content
        const activeContent = document.getElementById(browserId);
        if (activeContent) activeContent.classList.add('active');
    }
}

/**
 * Parallax scroll effect
 */
class ParallaxEffect {
    constructor() {
        this.header = document.querySelector('header');
        this.headerContent = document.querySelector('.header-content');

        if (this.header && this.headerContent) {
            this.init();
        }
    }

    init() {
        window.addEventListener('scroll', () => {
            this.updateParallax();
        });

        // Initial update
        this.updateParallax();
    }

    updateParallax() {
        const scrollPosition = window.scrollY;

        if (scrollPosition < SCROLL_THRESHOLD) {
            // Move the header content up slightly as user scrolls down
            this.headerContent.style.transform = `translateY(${scrollPosition * PARALLAX_FACTOR}px)`;

            // Create depth with background position shift
            this.header.style.backgroundPosition = `0px ${scrollPosition * 0.5}px`;
        }
    }
}

/**
 * Initialize all components when the DOM is fully loaded
 */
document.addEventListener('DOMContentLoaded', () => {
    // Initialize screenshot carousel
    const carousel = new ScreenshotCarousel();
    carousel.startRotation();

    // Initialize browser tab switcher
    const browserTabs = new BrowserTabSwitcher();

    // Initialize parallax effect
    const parallax = new ParallaxEffect();

    // Add accessibility improvements
    improveAccessibility();
});

/**
 * Improve accessibility for interactive elements
 */
function improveAccessibility() {
    // Make tabs keyboard navigable
    document.querySelectorAll('.screenshot-tab, .browser-tab').forEach(tab => {
        tab.setAttribute('role', 'tab');
        tab.setAttribute('tabindex', '0');

        // Handle keyboard interaction
        tab.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                tab.click();
            }
        });
    });

    // Improve tab panel accessibility
    document.querySelectorAll('.screenshot-content, .browser-content').forEach(content => {
        content.setAttribute('role', 'tabpanel');
    });
}