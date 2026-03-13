import type { JSX, ComponentChildren } from 'preact';
import { Icon, type IconName } from '../Icon';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'destructive';
export type ButtonSize = 'sm' | 'md' | 'lg';

interface ButtonProps extends Omit<JSX.HTMLAttributes<HTMLButtonElement>, 'size' | 'icon'> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: IconName;
  iconPosition?: 'left' | 'right';
  children?: ComponentChildren;
}

const VARIANT_CLASS: Record<ButtonVariant, string> = {
  primary: 'btn-primary',
  secondary: 'btn-secondary',
  ghost: 'btn-ghost',
  destructive: 'btn-destructive',
};

const SIZE_CLASS: Record<ButtonSize, string> = {
  sm: 'btn-sm',
  md: '',
  lg: 'btn-lg',
};

export function Button({
  variant = 'primary',
  size = 'md',
  icon,
  iconPosition = 'left',
  class: className,
  children,
  ...rest
}: ButtonProps) {
  const classes = [
    'btn',
    VARIANT_CLASS[variant],
    SIZE_CLASS[size],
    className,
  ].filter(Boolean).join(' ');

  const iconSize = size === 'sm' ? 'sm' : 'sm';

  return (
    <button class={classes} {...rest}>
      {icon && iconPosition === 'left' && <Icon name={icon} size={iconSize} />}
      {children}
      {icon && iconPosition === 'right' && <Icon name={icon} size={iconSize} />}
    </button>
  );
}
