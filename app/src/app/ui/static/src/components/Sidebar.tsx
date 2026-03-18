import { useRef, useEffect } from 'preact/hooks';
import { useRouter, type ViewName } from '../context/RouterContext';
import { ThemeToggle } from './ThemeToggle';
import { Icon, type IconName } from './Icon';

interface NavItem {
  id: ViewName;
  label: string;
  icon: IconName;
}

const NAV_ITEMS: NavItem[] = [
  { id: 'create-job', label: 'Create Job', icon: 'plus' },
  { id: 'job-history', label: 'Job History', icon: 'history' },
];

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { activeView, navigateTo } = useRouter();
  const sidebarRef = useRef<HTMLElement>(null);

  // Focus first nav link when mobile sidebar opens + trap focus inside
  useEffect(() => {
    if (!isOpen || !sidebarRef.current) return;

    const firstLink = sidebarRef.current.querySelector<HTMLButtonElement>('.sidebar-link');
    firstLink?.focus();

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') { onClose(); return; }
      if (e.key !== 'Tab' || !sidebarRef.current) return;
      const focusable = sidebarRef.current.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
      );
      if (focusable.length === 0) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  function handleNav(view: ViewName) {
    navigateTo(view);
    onClose();
  }

  const sidebarClass = ['sidebar', isOpen && 'open'].filter(Boolean).join(' ');

  return (
    <>
      <nav
        ref={sidebarRef}
        id="sidebar"
        class={sidebarClass}
        aria-label="Main navigation"
      >
        <div class="sidebar-header">
          <Icon name="rocket" class="sidebar-logo" />
          <span class="sidebar-title">Bombardment</span>
        </div>

        <div class="sidebar-nav">
          <span class="sidebar-section-label" role="heading" aria-level={2}>Jobs</span>
          {NAV_ITEMS.map((item) => {
            const isActive = activeView === item.id;
            return (
              <button
                key={item.id}
                type="button"
                class={`sidebar-link${isActive ? ' active' : ''}`}
                aria-current={isActive ? 'page' : undefined}
                onClick={() => handleNav(item.id)}
              >
                <Icon name={item.icon} size="sm" class="sidebar-link-icon" />
                <span>{item.label}</span>
                <kbd class="sidebar-shortcut">Alt+{NAV_ITEMS.indexOf(item) + 1}</kbd>
              </button>
            );
          })}
        </div>

        <div class="sidebar-footer">
          <div class="flex items-center justify-between">
            <ThemeToggle />
            <kbd class="sidebar-shortcut" title="Open command palette">Cmd+K</kbd>
          </div>
        </div>
      </nav>

      {/* Mobile overlay */}
      {isOpen && (
        <div
          class="sidebar-overlay"
          aria-hidden="true"
          onClick={onClose}
        />
      )}
    </>
  );
}
