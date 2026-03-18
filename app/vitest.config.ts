import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'happy-dom',
    include: ['src/app/ui/static/src/**/*.test.{ts,tsx}'],
    globals: true,
  },
  resolve: {
    alias: {
      'react': 'preact/compat',
      'react-dom': 'preact/compat',
      'react/jsx-runtime': 'preact/jsx-runtime',
    },
  },
  esbuild: {
    jsx: 'automatic',
    jsxImportSource: 'preact',
  },
});
