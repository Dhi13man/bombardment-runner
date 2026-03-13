import { createContext } from 'preact';
import { useState, useCallback, useContext, useMemo } from 'preact/hooks';
import type { ComponentChildren } from 'preact';

/** All routable views in the application. */
export type ViewName = 'create-job' | 'job-history' | 'job-progress';

/** Optional parameters passed between views. */
export interface ViewParams {
  jobId?: string;
}

interface RouterContextValue {
  /** Currently active view. */
  activeView: ViewName;
  /** Parameters for the current view. */
  params: ViewParams;
  /** Navigate to a named view with optional params. */
  navigateTo: (view: ViewName, params?: ViewParams) => void;
}

const RouterContext = createContext<RouterContextValue>({
  activeView: 'create-job',
  params: {},
  navigateTo: () => {},
});

export function RouterProvider({ children }: { children: ComponentChildren }) {
  const [activeView, setActiveView] = useState<ViewName>('create-job');
  const [params, setParams] = useState<ViewParams>({});

  const navigateTo = useCallback((view: ViewName, newParams?: ViewParams) => {
    setActiveView(view);
    setParams(newParams ?? {});
  }, []);

  const value = useMemo(
    () => ({ activeView, params, navigateTo }),
    [activeView, params, navigateTo],
  );

  return (
    <RouterContext.Provider value={value}>
      {children}
    </RouterContext.Provider>
  );
}

export function useRouter(): RouterContextValue {
  return useContext(RouterContext);
}
