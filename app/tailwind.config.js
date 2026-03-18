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
        glow: '0 0 20px rgba(129, 140, 248, 0.15)',
        focus: '0 0 0 2px var(--bg-base), 0 0 0 4px var(--accent)',
      },
      backdropBlur: {
        glass: 'var(--glass-blur)',
      },
      transitionDuration: {
        instant: '100ms',
        fast: '150ms',
        normal: '250ms',
        slow: '400ms',
      },
      fontSize: {
        xs: ['11px', { lineHeight: '1.5', letterSpacing: '0.01em' }],
        sm: ['13px', { lineHeight: '1.5', letterSpacing: '0' }],
        base: ['14px', { lineHeight: '1.6', letterSpacing: '0' }],
        lg: ['16px', { lineHeight: '1.5', letterSpacing: '0' }],
        xl: ['18px', { lineHeight: '1.4', letterSpacing: '-0.01em' }],
        '2xl': ['24px', { lineHeight: '1.3', letterSpacing: '-0.02em' }],
        '3xl': ['30px', { lineHeight: '1.2', letterSpacing: '-0.025em' }],
      },
      spacing: {
        0.5: '2px',
      },
    },
  },
  plugins: [],
};
