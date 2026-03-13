import { useThemeContext } from '../context/ThemeContext';
import { Icon } from './Icon';

export function ThemeToggle() {
  const { theme, toggle } = useThemeContext();
  const isLight = theme === 'light';

  return (
    <button
      type="button"
      class="theme-toggle"
      onClick={toggle}
      aria-label={isLight ? 'Switch to dark mode' : 'Switch to light mode'}
    >
      {isLight ? <Icon name="moon" size="sm" /> : <Icon name="sun" size="sm" />}
    </button>
  );
}
