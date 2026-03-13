import { render } from 'preact';

function App() {
  return (
    <div class="min-h-screen bg-bg-base text-text-primary flex items-center justify-center">
      <div class="text-center">
        <h1 class="text-2xl font-semibold text-accent">Bombardment</h1>
        <p class="mt-2 text-text-secondary">Build pipeline initialized. UI components coming soon.</p>
      </div>
    </div>
  );
}

const root = document.getElementById('app');
if (root) {
  render(<App />, root);
}
