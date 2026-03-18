import type { ComponentChildren } from 'preact';
import { useRef, useEffect, useState } from 'preact/hooks';
import { Icon } from '../Icon';

let modalIdCounter = 0;

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ComponentChildren;
  actions?: ComponentChildren;
  variant?: 'default' | 'destructive';
}

export function Modal({ open, onClose, title, children, actions, variant }: ModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const triggerRef = useRef<Element | null>(null);
  const [titleId] = useState(() => `modal-title-${++modalIdCounter}`);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (open) {
      triggerRef.current = document.activeElement;
      dialog.showModal();
    } else {
      dialog.close();
      if (triggerRef.current instanceof HTMLElement) {
        triggerRef.current.focus();
      }
    }
  }, [open]);

  if (!open) return null;

  return (
    <dialog
      ref={dialogRef}
      class="modal-backdrop"
      aria-labelledby={titleId}
      onClick={(e) => { if (e.target === dialogRef.current) onClose(); }}
      onCancel={onClose}
    >
      <div class={`modal-panel${variant === 'destructive' ? ' modal-destructive' : ''}`}>
        <div class="modal-header">
          <h2 id={titleId}>{title}</h2>
          <button
            type="button"
            class="btn btn-ghost btn-sm"
            aria-label="Close"
            onClick={onClose}
          >
            <Icon name="x" size="sm" />
          </button>
        </div>
        <div class="modal-body">
          {children}
        </div>
        {actions && (
          <div class="modal-footer">
            {actions}
          </div>
        )}
      </div>
    </dialog>
  );
}

interface ConfirmModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void | Promise<void>;
  title: string;
  message: string;
  confirmLabel?: string;
  variant?: 'default' | 'destructive';
}

export function ConfirmModal({
  open,
  onClose,
  onConfirm,
  title,
  message,
  confirmLabel = 'Confirm',
  variant = 'default',
}: ConfirmModalProps) {
  const [loading, setLoading] = useState(false);

  async function handleConfirm() {
    setLoading(true);
    try {
      await onConfirm();
      onClose();
    } catch {
      // Error handling is expected to be done in onConfirm via toast
    } finally {
      setLoading(false);
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={title}
      variant={variant}
      actions={
        <>
          <button type="button" class="btn btn-secondary" onClick={onClose} disabled={loading}>Cancel</button>
          <button
            type="button"
            class={`btn ${variant === 'destructive' ? 'btn-destructive' : 'btn-primary'}`}
            onClick={handleConfirm}
            disabled={loading}
            aria-busy={loading ? 'true' : undefined}
          >
            {loading ? 'Processing...' : confirmLabel}
          </button>
        </>
      }
    >
      <p>{message}</p>
    </Modal>
  );
}
