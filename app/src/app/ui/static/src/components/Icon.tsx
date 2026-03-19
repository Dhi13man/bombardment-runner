import type { JSX } from 'preact';

/**
 * All available icon names in the Lucide SVG sprite.
 * Must stay in sync with scripts/build-sprite.js ICON_LIST.
 */
export type IconName =
  | 'rocket'
  | 'file-input'
  | 'sliders-horizontal'
  | 'target'
  | 'check-circle-2'
  | 'arrow-left'
  | 'arrow-right'
  | 'code'
  | 'link'
  | 'tags'
  | 'clock'
  | 'heart-pulse'
  | 'shield'
  | 'file-code'
  | 'hourglass'
  | 'timer'
  | 'plug'
  | 'scale'
  | 'upload'
  | 'plus'
  | 'trash-2'
  | 'refresh-cw'
  | 'alert-triangle'
  | 'info'
  | 'folder'
  | 'layers'
  | 'settings'
  | 'list-checks'
  | 'history'
  | 'inbox'
  | 'x'
  | 'chevron-down'
  | 'sun'
  | 'moon'
  | 'menu'
  | 'panel-left'
  | 'check'
  | 'search'
  | 'database';

export type IconSize = 'sm' | 'md' | 'lg' | 'xl';

const SIZE_CLASSES: Record<IconSize, string> = {
  sm: 'w-4 h-4',   // 16px
  md: 'w-5 h-5',   // 20px
  lg: 'w-6 h-6',   // 24px
  xl: 'w-8 h-8',   // 32px
};

interface IconProps {
  name: IconName;
  size?: IconSize;
  class?: string;
  /** For icon-only buttons: provides accessible label */
  'aria-label'?: string;
}

/**
 * Renders an icon from the Lucide SVG sprite.
 *
 * Decorative icons (default) get `aria-hidden="true"`.
 * Interactive icons with `aria-label` are announced by screen readers.
 */
export function Icon({ name, size = 'md', class: className, ...rest }: IconProps): JSX.Element {
  const sizeClass = SIZE_CLASSES[size];
  const classes = [sizeClass, 'flex-shrink-0', className].filter(Boolean).join(' ');
  const isDecorative = !rest['aria-label'];

  return (
    <svg class={classes} aria-hidden={isDecorative ? 'true' : undefined} {...rest}>
      <use href={`/static/icons/sprite.svg#icon-${name}`} />
    </svg>
  );
}
