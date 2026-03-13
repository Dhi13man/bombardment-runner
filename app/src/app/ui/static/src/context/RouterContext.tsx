import { createContext } from 'preact';
import { useState, useCallback, useContext } from 'preact/hooks';
import type { ComponentChildren } from 'preact';

/** All routable views in the application. */
export type ViewName = 'create-job' | 'job-history' | 'job-progress';

interface RouterContextValue {
  /** Currently active view. */
  activeView: ViewName;
  /** Navigate to a named view. No-op if the name is invalid. */
  navigateTo: (view: ViewName) => void;
}

const RouterContext = createContext<RouterContextValue>({
  activeView: 'create-job',
  navigateTo: () => {},
});

export function RouterProvider({ children }: { children: ComponentChildren }) {
  const [activeView, setActiveView] = useState<ViewName>('create-job');

  const navigateTo = useCallback((view: ViewName) => {
    setActiveView(view);
  }, []);

  return (
    <RouterContext.Provider value={{ activeView, navigateTo }}>
      {children}
    </RouterContext.Provider>
  );
}

export function useRouter(): RouterContextValue {
  return useContext(RouterContext);
}
