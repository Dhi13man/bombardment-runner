import { render } from 'preact';
import { RouterProvider } from './context/RouterContext';
import { ToastProvider } from './components/composites/Toast';
import { AppShell } from './components/AppShell';

function App() {
  return (
    <RouterProvider>
      <ToastProvider>
        <AppShell />
      </ToastProvider>
    </RouterProvider>
  );
}

const root = document.getElementById('app');
if (root) {
  render(<App />, root);
}
