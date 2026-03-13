import { render } from 'preact';
import { ThemeProvider } from './context/ThemeContext';
import { RouterProvider } from './context/RouterContext';
import { AppShell } from './components/AppShell';

function App() {
  return (
    <ThemeProvider>
      <RouterProvider>
        <AppShell />
      </RouterProvider>
    </ThemeProvider>
  );
}

const root = document.getElementById('app');
if (root) {
  render(<App />, root);
}
