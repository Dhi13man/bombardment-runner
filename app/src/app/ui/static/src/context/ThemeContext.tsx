import { createContext } from 'preact';
import { useContext, useMemo } from 'preact/hooks';
import type { ComponentChildren } from 'preact';
import { useTheme, type Theme } from '../hooks/useTheme';

interface ThemeContextValue {
  theme: Theme;
  toggle: () => void;
}

const ThemeContext = createContext<ThemeContextValue>({
  theme: 'dark',
  toggle: () => {},
});

export function ThemeProvider({ children }: { children: ComponentChildren }) {
  const { theme, toggle } = useTheme();
  const value = useMemo(() => ({ theme, toggle }), [theme, toggle]);

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useThemeContext(): ThemeContextValue {
  return useContext(ThemeContext);
}
