/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/app/ui/**/*.html',
    './src/app/ui/static/src/**/*.{ts,tsx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        'bg-base': 'var(--bg-base)',
        'bg-surface': 'var(--bg-surface)',
        'bg-elevated': 'var(--bg-elevated)',
        'bg-overlay': 'var(--bg-overlay)',
        'bg-inset': 'var(--bg-inset)',
        'text-primary': 'var(--text-primary)',
        'text-secondary': 'var(--text-secondary)',
        'text-tertiary': 'var(--text-tertiary)',
        'text-inverse': 'var(--text-inverse)',
        accent: {
          DEFAULT: 'var(--accent)',
          hover: 'var(--accent-hover)',
          active: 'var(--accent-active)',
          muted: 'var(--accent-muted)',
          subtle: 'var(--accent-subtle)',
          border: 'var(--accent-border)',
        },
        success: {
          DEFAULT: 'var(--status-success)',
          muted: 'var(--status-success-muted)',
          text: 'var(--status-success-text)',
        },
        warning: {
          DEFAULT: 'var(--status-warning)',
          muted: 'var(--status-warning-muted)',
          text: 'var(--status-warning-text)',
        },
        error: {
          DEFAULT: 'var(--status-error)',
          muted: 'var(--status-error-muted)',
          text: 'var(--status-error-text)',
        },
        info: {
          DEFAULT: 'var(--status-info)',
          muted: 'var(--status-info-muted)',
          text: 'var(--status-info-text)',
        },
        glass: {
          bg: 'var(--glass-bg)',
          border: 'var(--glass-border)',
        },
      },
      borderColor: {
        DEFAULT: 'var(--border-default)',
        hover: 'var(--border-hover)',
        focus: 'var(--border-focus)',
      },
      fontFamily: {
        sans: ['var(--font-sans)'],
        mono: ['var(--font-mono)'],
      },
      borderRadius: {
        sm: '4px',
        DEFAULT: '8px',
        md: '8px',
        lg: '12px',
        xl: '16px',
      },
      boxShadow: {
        sm: '0 1px 2px rgba(0, 0, 0, 0.3)',
        DEFAULT: '0 4px 6px -1px rgba(0, 0, 0, 0.4)',
        lg: '0 10px 15px -3px rgba(0, 0, 0, 0.5)',
        glow: '0 0 20px rgba(6, 182, 212, 0.15)',
      },
      backdropBlur: {
        glass: 'var(--glass-blur)',
      },
    },
  },
  plugins: [],
};
