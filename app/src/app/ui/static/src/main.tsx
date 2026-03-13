import { render } from 'preact';
import { ThemeProvider } from './context/ThemeContext';
import { RouterProvider } from './context/RouterContext';
import { ToastProvider } from './components/composites/Toast';
import { AppShell } from './components/AppShell';

function App() {
  return (
    <ThemeProvider>
      <RouterProvider>
        <ToastProvider>
          <AppShell />
        </ToastProvider>
      </RouterProvider>
    </ThemeProvider>
  );
}

const root = document.getElementById('app');
if (root) {
  render(<App />, root);
}
