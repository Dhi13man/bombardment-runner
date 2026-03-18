import { useRef, useState, useEffect } from 'preact/hooks';
import { useSidebar } from '../hooks/useSidebar';
import { useRouter } from '../context/RouterContext';
import { Sidebar } from './Sidebar';
import { ViewRouter } from './ViewRouter';
import { CommandPalette } from './composites/CommandPalette';
import { Icon } from './Icon';

/**
 * Top-level application shell.
 * Provides the sidebar + main content layout with mobile support.
 */
export function AppShell() {
  const { isOpen, open, close } = useSidebar();
  const toggleRef = useRef<HTMLButtonElement>(null);
  const { navigateTo } = useRouter();
  const [paletteOpen, setPaletteOpen] = useState(false);

  useEffect(() => {
    function handleGlobalKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setPaletteOpen((o) => !o);
      }
      if (e.altKey && !e.ctrlKey && !e.metaKey) {
        if (e.key === '1') { e.preventDefault(); navigateTo('create-job'); }
        if (e.key === '2') { e.preventDefault(); navigateTo('job-history'); }
      }
    }
    document.addEventListener('keydown', handleGlobalKeyDown);
    return () => document.removeEventListener('keydown', handleGlobalKeyDown);
  }, [navigateTo]);

  function handleClose() {
    close();
    toggleRef.current?.focus();
  }

  return (
    <>
      {/* Skip-to-content link (visible on Tab) */}
      <a href="#main-content" class="skip-link">
        Skip to content
      </a>

      <div class="app-layout">
        <Sidebar isOpen={isOpen} onClose={handleClose} />

        {/* Mobile header with hamburger */}
        <header class="mobile-header">
          <button
            ref={toggleRef}
            type="button"
            class="btn btn-ghost"
            aria-label="Open navigation"
            aria-expanded={isOpen}
            aria-controls="sidebar"
            onClick={open}
          >
            <Icon name="menu" size="md" />
          </button>
          <span class="text-md font-semibold">Bombardment</span>
        </header>

        {/* Main content area */}
        <main id="main-content" class="main-content">
          <ViewRouter />
        </main>
      </div>
      <CommandPalette isOpen={paletteOpen} onClose={() => setPaletteOpen(false)} />
    </>
  );
}
