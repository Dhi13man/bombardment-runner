import { render } from 'preact';
import { ThemeProvider } from './context/ThemeContext';
import { RouterProvider } from './context/RouterContext';
import { ToastProvider } from './components/composites/Toast';
import { JobFormProvider } from './context/JobFormContext';
import { AppShell } from './components/AppShell';

function App() {
  return (
    <ThemeProvider>
      <RouterProvider>
        <ToastProvider>
          <JobFormProvider>
            <AppShell />
          </JobFormProvider>
        </ToastProvider>
      </RouterProvider>
    </ThemeProvider>
  );
}

const root = document.getElementById('app');
if (root) {
  render(<App />, root);
}
