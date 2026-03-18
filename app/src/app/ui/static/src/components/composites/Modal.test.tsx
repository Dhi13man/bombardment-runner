import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, act } from '@testing-library/preact';
import { Modal, ConfirmModal } from './Modal';

// happy-dom has limited <dialog> support; stub showModal/close
beforeEach(() => {
  HTMLDialogElement.prototype.showModal = vi.fn(function (this: HTMLDialogElement) {
    this.setAttribute('open', '');
  });
  HTMLDialogElement.prototype.close = vi.fn(function (this: HTMLDialogElement) {
    this.removeAttribute('open');
  });
});

describe('Modal', () => {
  it('renders nothing when closed', () => {
    // Arrange & Act
    const { container } = render(
      <Modal open={false} onClose={() => {}} title="Test">
        <p>Content</p>
      </Modal>,
    );

    // Assert
    expect(container.querySelector('dialog')).toBeNull();
  });

  it('renders dialog when open', () => {
    // Arrange & Act
    const { getByText } = render(
      <Modal open={true} onClose={() => {}} title="Test Modal">
        <p>Modal content</p>
      </Modal>,
    );

    // Assert
    expect(getByText('Test Modal')).toBeTruthy();
    expect(getByText('Modal content')).toBeTruthy();
  });

  it('calls showModal when opened', () => {
    // Arrange & Act
    render(
      <Modal open={true} onClose={() => {}} title="Test">
        <p>Content</p>
      </Modal>,
    );

    // Assert
    expect(HTMLDialogElement.prototype.showModal).toHaveBeenCalled();
  });

  it('has aria-labelledby pointing to title', () => {
    // Arrange & Act
    const { container, getByText } = render(
      <Modal open={true} onClose={() => {}} title="Labeled Modal">
        <p>Content</p>
      </Modal>,
    );

    // Assert
    const dialog = container.querySelector('dialog');
    const titleId = dialog?.getAttribute('aria-labelledby');
    expect(titleId).toBeTruthy();
    const titleEl = getByText('Labeled Modal');
    expect(titleEl.id).toBe(titleId);
  });

  it('calls onClose when close button clicked', () => {
    // Arrange
    const mockClose = vi.fn();
    const { getByLabelText } = render(
      <Modal open={true} onClose={mockClose} title="Test">
        <p>Content</p>
      </Modal>,
    );

    // Act
    fireEvent.click(getByLabelText('Close'));

    // Assert
    expect(mockClose).toHaveBeenCalledOnce();
  });

  it('calls onClose when backdrop (dialog itself) clicked', () => {
    // Arrange
    const mockClose = vi.fn();
    const { container } = render(
      <Modal open={true} onClose={mockClose} title="Test">
        <p>Content</p>
      </Modal>,
    );
    const dialog = container.querySelector('dialog')!;

    // Act - click on the dialog element itself (backdrop)
    fireEvent.click(dialog);

    // Assert
    expect(mockClose).toHaveBeenCalledOnce();
  });

  it('does not call onClose when panel content clicked', () => {
    // Arrange
    const mockClose = vi.fn();
    const { getByText } = render(
      <Modal open={true} onClose={mockClose} title="Test">
        <p>Inner content</p>
      </Modal>,
    );

    // Act - click on inner content (not the dialog itself)
    fireEvent.click(getByText('Inner content'));

    // Assert
    expect(mockClose).not.toHaveBeenCalled();
  });

  it('renders actions footer when provided', () => {
    // Arrange & Act
    const { getByText } = render(
      <Modal
        open={true}
        onClose={() => {}}
        title="Test"
        actions={<button type="button">Save</button>}
      >
        <p>Content</p>
      </Modal>,
    );

    // Assert
    expect(getByText('Save')).toBeTruthy();
  });

  it('restores focus to trigger element on close', () => {
    // Arrange
    const triggerButton = document.createElement('button');
    triggerButton.textContent = 'Trigger';
    document.body.appendChild(triggerButton);
    triggerButton.focus();
    const mockClose = vi.fn();

    const { rerender } = render(
      <Modal open={true} onClose={mockClose} title="Focus Test">
        <p>Content</p>
      </Modal>,
    );

    // Act - close the modal
    rerender(
      <Modal open={false} onClose={mockClose} title="Focus Test">
        <p>Content</p>
      </Modal>,
    );

    // Assert
    expect(document.activeElement).toBe(triggerButton);

    // Cleanup
    document.body.removeChild(triggerButton);
  });
});

describe('ConfirmModal', () => {
  it('renders message and confirm/cancel buttons when open', () => {
    // Arrange & Act
    const { getByText } = render(
      <ConfirmModal
        open={true}
        onClose={() => {}}
        onConfirm={() => {}}
        title="Confirm Delete"
        message="Are you sure?"
        confirmLabel="Delete"
      />,
    );

    // Assert
    expect(getByText('Confirm Delete')).toBeTruthy();
    expect(getByText('Are you sure?')).toBeTruthy();
    expect(getByText('Delete')).toBeTruthy();
    expect(getByText('Cancel')).toBeTruthy();
  });

  it('calls onConfirm then onClose on confirm button click', async () => {
    // Arrange
    const mockConfirm = vi.fn().mockResolvedValue(undefined);
    const mockClose = vi.fn();
    const { getByText } = render(
      <ConfirmModal
        open={true}
        onClose={mockClose}
        onConfirm={mockConfirm}
        title="Test"
        message="Sure?"
      />,
    );

    // Act
    await act(async () => {
      fireEvent.click(getByText('Confirm'));
    });

    // Assert
    expect(mockConfirm).toHaveBeenCalledOnce();
    expect(mockClose).toHaveBeenCalledOnce();
  });

  it('calls onClose on cancel button click', () => {
    // Arrange
    const mockClose = vi.fn();
    const { getByText } = render(
      <ConfirmModal
        open={true}
        onClose={mockClose}
        onConfirm={() => {}}
        title="Test"
        message="Sure?"
      />,
    );

    // Act
    fireEvent.click(getByText('Cancel'));

    // Assert
    expect(mockClose).toHaveBeenCalledOnce();
  });

  it('shows loading state during async confirm', async () => {
    // Arrange
    let resolveConfirm: () => void;
    const confirmPromise = new Promise<void>((r) => { resolveConfirm = r; });
    const { getByText } = render(
      <ConfirmModal
        open={true}
        onClose={() => {}}
        onConfirm={() => confirmPromise}
        title="Test"
        message="Sure?"
      />,
    );

    // Act - click confirm, don't resolve yet
    await act(async () => {
      fireEvent.click(getByText('Confirm'));
    });

    // Assert - should show loading text
    expect(getByText('Processing...')).toBeTruthy();

    // Cleanup - resolve to prevent hanging
    await act(async () => { resolveConfirm!(); });
  });

  it('applies destructive variant class to confirm button', () => {
    // Arrange & Act
    const { getAllByText } = render(
      <ConfirmModal
        open={true}
        onClose={() => {}}
        onConfirm={() => {}}
        title="Delete Item"
        message="Sure?"
        variant="destructive"
        confirmLabel="Delete"
      />,
    );

    // Assert - find the button (not the h2 title)
    const deleteBtn = getAllByText('Delete').find((el) => el.tagName === 'BUTTON');
    expect(deleteBtn).toBeTruthy();
    expect(deleteBtn!.className).toContain('btn-destructive');
  });
});
