/**
 * Bombardment Landing Page - Interactions & Animations
 */

// 1. Scroll reveal fallback (browsers without CSS animation-timeline)
function initScrollReveals() {
  if (typeof CSS !== 'undefined' && CSS.supports && CSS.supports('animation-timeline: view()')) return;

  const observer = new IntersectionObserver(function (entries) {
    entries.forEach(function (entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.1, rootMargin: '0px 0px -60px 0px' });

  document.querySelectorAll('.reveal').forEach(function (el) {
    observer.observe(el);
  });

  // Safety fallback: force-reveal after 3s if observer never fires
  setTimeout(function () {
    document.querySelectorAll('.reveal:not(.visible)').forEach(function (el) {
      el.classList.add('visible');
    });
  }, 3000);
}

// 2. Floating nav scroll behavior (RAF-throttled)
function initNavScroll() {
  const nav = document.querySelector('.nav');
  if (!nav) return;

  let ticking = false;
  window.addEventListener('scroll', function () {
    if (!ticking) {
      requestAnimationFrame(function () {
        nav.classList.toggle('scrolled', window.scrollY > 100);
        ticking = false;
      });
      ticking = true;
    }
  }, { passive: true });
}

// 3. Mobile nav toggle
function initMobileNav() {
  const toggle = document.querySelector('.nav-toggle');
  const nav = document.querySelector('.nav');
  if (!toggle || !nav) return;

  toggle.addEventListener('click', function () {
    const open = nav.classList.toggle('nav-open');
    toggle.setAttribute('aria-expanded', String(open));
    if (open) {
      const firstLink = nav.querySelector('.nav-links a');
      if (firstLink) firstLink.focus();
    }
  });

  nav.querySelectorAll('.nav-links a').forEach(function (link) {
    link.addEventListener('click', function () {
      nav.classList.remove('nav-open');
      toggle.setAttribute('aria-expanded', 'false');
    });
  });

  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && nav.classList.contains('nav-open')) {
      nav.classList.remove('nav-open');
      toggle.setAttribute('aria-expanded', 'false');
      toggle.focus();
    }
  });
}

// 4. Bento tile cursor tracking (attach/detach on enter/leave)
function initBentoGlow() {
  document.querySelectorAll('.bento-tile').forEach(function (tile) {
    let rafId = 0;

    function onMove(e) {
      cancelAnimationFrame(rafId);
      rafId = requestAnimationFrame(function () {
        const rect = tile.getBoundingClientRect();
        tile.style.setProperty('--mouse-x', (e.clientX - rect.left) + 'px');
        tile.style.setProperty('--mouse-y', (e.clientY - rect.top) + 'px');
      });
    }

    tile.addEventListener('pointerenter', function () {
      tile.addEventListener('pointermove', onMove);
    });

    tile.addEventListener('pointerleave', function () {
      tile.removeEventListener('pointermove', onMove);
      cancelAnimationFrame(rafId);
    });
  });
}

// 5. Generic tab controller (shared by segment control and screenshot tabs)
function initTabs(opts) {
  document.querySelectorAll(opts.tablistSelector).forEach(function (tablist) {
    const tabs = Array.from(tablist.querySelectorAll('[role="tab"]'));

    function activateTab(btn) {
      const targetId = btn.getAttribute(opts.dataAttr);
      const container = opts.getContainer(btn, tablist);
      if (!container) return;

      tabs.forEach(function (t) {
        const isActive = t === btn;
        t.classList.toggle('active', isActive);
        t.setAttribute('aria-selected', String(isActive));
        t.setAttribute('tabindex', isActive ? '0' : '-1');
      });

      container.querySelectorAll(opts.panelSelector).forEach(function (panel) {
        if (opts.useHidden) {
          panel.hidden = panel.id !== targetId;
        } else {
          panel.classList.toggle('active', panel.id === targetId);
        }
      });

      btn.focus();
    }

    tabs.forEach(function (btn, i) {
      btn.setAttribute('tabindex', i === 0 ? '0' : '-1');
      btn.addEventListener('click', function () { activateTab(btn); });
    });

    tablist.addEventListener('keydown', function (e) {
      const idx = tabs.indexOf(document.activeElement);
      if (idx === -1) return;

      let next = -1;
      if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
        next = (idx + 1) % tabs.length;
      } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
        next = (idx - 1 + tabs.length) % tabs.length;
      } else if (e.key === 'Home') {
        next = 0;
      } else if (e.key === 'End') {
        next = tabs.length - 1;
      }

      if (next !== -1) {
        e.preventDefault();
        activateTab(tabs[next]);
      }
    });
  });
}

function initSegmentControl() {
  initTabs({
    tablistSelector: '[role="tablist"].segment-control',
    dataAttr: 'data-tab',
    panelSelector: '.demo-panel',
    useHidden: true,
    getContainer: function (btn) { return btn.closest('section'); }
  });
}

function initScreenshotTabs() {
  initTabs({
    tablistSelector: '.screenshot-tabs[role="tablist"]',
    dataAttr: 'data-screenshot',
    panelSelector: '.screenshot-content',
    useHidden: false,
    getContainer: function (btn, tablist) { return tablist.closest('.demo-panel') || tablist.parentElement; }
  });
}

// 6. Copy to clipboard
function initCopyButtons() {
  document.querySelectorAll('.copy-btn').forEach(function (btn) {
    btn.addEventListener('click', function () {
      const block = btn.closest('.cli-block, .install-cmd');
      if (!block) return;

      const code = block.querySelector('code');
      if (!code) return;

      navigator.clipboard.writeText(code.textContent).then(function () {
        btn.classList.add('copied');
        btn.setAttribute('aria-label', 'Copied');
        setTimeout(function () {
          btn.classList.remove('copied');
          btn.setAttribute('aria-label', 'Copy to clipboard');
        }, 2000);
      }).catch(function () {
        // Fallback: select text for manual copy
        const range = document.createRange();
        range.selectNodeContents(code);
        const sel = window.getSelection();
        sel.removeAllRanges();
        sel.addRange(range);
      });
    });
  });
}

// 7. Pipeline particle animation
function initPipelineParticles() {
  const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
  if (mq.matches) return;

  document.querySelectorAll('.pipeline-step-connector').forEach(function (conn) {
    for (let i = 0; i < 2; i++) {
      const particle = document.createElement('span');
      particle.className = 'data-particle';
      conn.appendChild(particle);
    }
  });
}

// 8. Pause off-screen infinite animations to save CPU/battery
function initAnimationGating() {
  const sections = document.querySelectorAll('.pipeline-detail, .scope');
  if (!sections.length) return;

  const observer = new IntersectionObserver(function (entries) {
    entries.forEach(function (entry) {
      entry.target.classList.toggle('in-view', entry.isIntersecting);
    });
  }, { threshold: 0 });

  sections.forEach(function (s) { observer.observe(s); });
}

// Init all on DOMContentLoaded
document.addEventListener('DOMContentLoaded', function () {
  document.body.classList.add('js-ready');
  initScrollReveals();
  initNavScroll();
  initMobileNav();
  initBentoGlow();
  initSegmentControl();
  initScreenshotTabs();
  initCopyButtons();
  initPipelineParticles();
  initAnimationGating();
});
