import { createContext } from 'preact';
import { useState, useCallback, useContext, useMemo, useEffect } from 'preact/hooks';
import type { ComponentChildren } from 'preact';

/** All routable views in the application. */
export type ViewName = 'create-job' | 'job-history' | 'job-progress';

interface ViewParams {
  jobId?: string;
}

interface RouterContextValue {
  activeView: ViewName;
  params: ViewParams;
  navigateTo: (view: ViewName, params?: ViewParams) => void;
}

function viewToHash(view: ViewName, params?: ViewParams): string {
  switch (view) {
    case 'create-job': return '#/create';
    case 'job-history': return '#/history';
    case 'job-progress': return `#/jobs/${params?.jobId ?? ''}`;
  }
}

function hashToRoute(hash: string): { view: ViewName; params: ViewParams } {
  if (hash.startsWith('#/jobs/') && hash.length > 7) {
    return { view: 'job-progress', params: { jobId: hash.slice(7) } };
  }
  if (hash === '#/history') return { view: 'job-history', params: {} };
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
  const [initialRoute] = useState(() => hashToRoute(window.location.hash));
  const [activeView, setActiveView] = useState<ViewName>(initialRoute.view);
  const [params, setParams] = useState<ViewParams>(initialRoute.params);

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
