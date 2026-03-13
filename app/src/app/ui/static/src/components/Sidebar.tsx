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

  // Focus first nav link when mobile sidebar opens
  useEffect(() => {
    if (isOpen) {
      const firstLink = sidebarRef.current?.querySelector<HTMLButtonElement>('.sidebar-link');
      firstLink?.focus();
    }
  }, [isOpen]);

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
          <span class="sidebar-section-label">Jobs</span>
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
                <Icon name={item.icon} size="sm" />
                <span>{item.label}</span>
              </button>
            );
          })}
        </div>

        <div class="sidebar-footer">
          <ThemeToggle />
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
