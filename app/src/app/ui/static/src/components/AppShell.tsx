import { useRef } from 'preact/hooks';
import { useSidebar } from '../hooks/useSidebar';
import { Sidebar } from './Sidebar';
import { ViewRouter } from './ViewRouter';
import { Icon } from './Icon';

/**
 * Top-level application shell.
 * Provides the sidebar + main content layout with mobile support.
 */
export function AppShell() {
  const { isOpen, open, close } = useSidebar();
  const toggleRef = useRef<HTMLButtonElement>(null);

  function handleClose() {
    close();
    // Return focus to hamburger trigger per accessibility spec
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
    </>
  );
}
