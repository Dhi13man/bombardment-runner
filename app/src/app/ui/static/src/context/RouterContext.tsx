import { createContext } from 'preact';
import { useState, useCallback, useContext, useMemo, useEffect } from 'preact/hooks';
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

const ROUTE_MAP: Record<string, { view: ViewName; extractParams?: (hash: string) => ViewParams }> = {
  '#/create': { view: 'create-job' },
  '#/history': { view: 'job-history' },
  '#/jobs/': {
    view: 'job-progress',
    extractParams: (hash) => ({ jobId: hash.replace('#/jobs/', '') }),
  },
};

function viewToHash(view: ViewName, params?: ViewParams): string {
  switch (view) {
    case 'create-job': return '#/create';
    case 'job-history': return '#/history';
    case 'job-progress': return `#/jobs/${params?.jobId ?? ''}`;
  }
}

function hashToRoute(hash: string): { view: ViewName; params: ViewParams } {
  if (hash.startsWith('#/jobs/') && hash.length > 7) {
    return { view: 'job-progress', params: { jobId: hash.replace('#/jobs/', '') } };
  }
  for (const [pattern, route] of Object.entries(ROUTE_MAP)) {
    if (hash === pattern) {
      return { view: route.view, params: route.extractParams?.(hash) ?? {} };
    }
  }
  return { view: 'create-job', params: {} };
}

function viewTitle(view: ViewName, params?: ViewParams): string {
  switch (view) {
    case 'create-job': return 'Create Job - Bombardment';
    case 'job-history': return 'Job History - Bombardment';
    case 'job-progress': {
      const id = params?.jobId;
      return id ? `Job #${id.substring(0, 6)} - Bombardment` : 'Job Progress - Bombardment';
    }
  }
}

const RouterContext = createContext<RouterContextValue>({
  activeView: 'create-job',
  params: {},
  navigateTo: () => {},
});

export function RouterProvider({ children }: { children: ComponentChildren }) {
  const [activeView, setActiveView] = useState<ViewName>(() => hashToRoute(window.location.hash).view);
  const [params, setParams] = useState<ViewParams>(() => hashToRoute(window.location.hash).params);

  const navigateTo = useCallback((view: ViewName, newParams?: ViewParams) => {
    const hash = viewToHash(view, newParams);
    window.history.pushState(null, '', hash);
    setActiveView(view);
    setParams(newParams ?? {});
    document.title = viewTitle(view, newParams);
  }, []);

  useEffect(() => {
    function handlePopState() {
      const route = hashToRoute(window.location.hash);
      setActiveView(route.view);
      setParams(route.params);
      document.title = viewTitle(route.view, route.params);
    }
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  // Set initial title on mount only
  useEffect(() => {
    document.title = viewTitle(activeView, params);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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
