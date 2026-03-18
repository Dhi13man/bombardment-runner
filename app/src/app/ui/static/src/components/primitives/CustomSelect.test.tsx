import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@testing-library/preact';
import { CustomSelect } from './CustomSelect';

const OPTIONS = [
  { value: 'csv', label: 'CSV' },
  { value: 'json', label: 'JSON' },
  { value: 'xml', label: 'XML', disabled: true },
];

describe('CustomSelect', () => {
  it('renders with placeholder when no value selected', () => {
    // Arrange & Act
    const { getByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={() => {}} placeholder="Pick one" />,
    );

    // Assert
    const trigger = getByRole('combobox');
    expect(trigger.textContent).toContain('Pick one');
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
  });

  it('renders selected option label', () => {
    // Arrange & Act
    const { getByRole } = render(
      <CustomSelect options={OPTIONS} value="json" onChange={() => {}} />,
    );

    // Assert
    expect(getByRole('combobox').textContent).toContain('JSON');
  });

  it('opens dropdown on click and shows listbox', () => {
    // Arrange
    const { getByRole, queryByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={() => {}} />,
    );

    // Act
    fireEvent.click(getByRole('combobox'));

    // Assert
    expect(getByRole('combobox').getAttribute('aria-expanded')).toBe('true');
    expect(queryByRole('listbox')).toBeTruthy();
  });

  it('calls onChange when option is clicked', () => {
    // Arrange
    const mockOnChange = vi.fn();
    const { getByRole, getAllByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={mockOnChange} />,
    );

    // Act
    fireEvent.click(getByRole('combobox')); // open
    const opts = getAllByRole('option');
    fireEvent.click(opts[1]); // click "JSON"

    // Assert
    expect(mockOnChange).toHaveBeenCalledWith('json');
  });

  it('does not call onChange for disabled options', () => {
    // Arrange
    const mockOnChange = vi.fn();
    const { getByRole, getAllByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={mockOnChange} />,
    );

    // Act
    fireEvent.click(getByRole('combobox'));
    const opts = getAllByRole('option');
    fireEvent.click(opts[2]); // click "XML" (disabled)

    // Assert
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('navigates with ArrowDown/ArrowUp keys', () => {
    // Arrange
    const { getByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={() => {}} />,
    );
    const trigger = getByRole('combobox');

    // Act - open with ArrowDown
    fireEvent.keyDown(trigger, { key: 'ArrowDown' });

    // Assert - should be open
    expect(trigger.getAttribute('aria-expanded')).toBe('true');
    // activedescendant should be set
    expect(trigger.getAttribute('aria-activedescendant')).toBeTruthy();
  });

  it('closes on Escape key', () => {
    // Arrange
    const { getByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={() => {}} />,
    );
    const trigger = getByRole('combobox');

    // Act
    fireEvent.click(trigger); // open
    expect(trigger.getAttribute('aria-expanded')).toBe('true');
    fireEvent.keyDown(trigger, { key: 'Escape' });

    // Assert
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
  });

  it('selects on Enter key', () => {
    // Arrange
    const mockOnChange = vi.fn();
    const { getByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={mockOnChange} />,
    );
    const trigger = getByRole('combobox');

    // Act
    fireEvent.keyDown(trigger, { key: 'ArrowDown' }); // open, highlight first
    fireEvent.keyDown(trigger, { key: 'Enter' }); // select

    // Assert
    expect(mockOnChange).toHaveBeenCalledWith('csv');
  });

  it('skips disabled options during arrow navigation', () => {
    // Arrange
    const mockOnChange = vi.fn();
    const { getByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={mockOnChange} />,
    );
    const trigger = getByRole('combobox');

    // Act - open, go to JSON (index 1), then ArrowDown should skip XML (disabled)
    fireEvent.keyDown(trigger, { key: 'ArrowDown' }); // open at csv (0)
    fireEvent.keyDown(trigger, { key: 'ArrowDown' }); // json (1)
    fireEvent.keyDown(trigger, { key: 'ArrowDown' }); // should stay at json (xml disabled)
    fireEvent.keyDown(trigger, { key: 'Enter' });

    // Assert
    expect(mockOnChange).toHaveBeenCalledWith('json');
  });

  it('generates stable IDs even without id prop', () => {
    // Arrange & Act
    const { getByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={() => {}} />,
    );
    const trigger = getByRole('combobox');

    // Assert - should have a generated ID and aria-controls
    expect(trigger.id).toBeTruthy();
    expect(trigger.getAttribute('aria-controls')).toBeTruthy();
  });

  it('shows error state with aria-invalid', () => {
    // Arrange & Act
    const { getByRole, getByText } = render(
      <CustomSelect
        options={OPTIONS}
        value=""
        onChange={() => {}}
        error="Required field"
      />,
    );

    // Assert
    expect(getByRole('combobox').getAttribute('aria-invalid')).toBe('true');
    expect(getByText('Required field')).toBeTruthy();
  });

  it('supports type-ahead character search', () => {
    // Arrange
    const mockOnChange = vi.fn();
    const { getByRole } = render(
      <CustomSelect options={OPTIONS} value="" onChange={mockOnChange} />,
    );
    const trigger = getByRole('combobox');

    // Act - open then type 'j' to jump to JSON
    fireEvent.keyDown(trigger, { key: 'ArrowDown' }); // open
    fireEvent.keyDown(trigger, { key: 'j' }); // type-ahead to JSON
    fireEvent.keyDown(trigger, { key: 'Enter' });

    // Assert
    expect(mockOnChange).toHaveBeenCalledWith('json');
  });

  it('renders label when provided', () => {
    // Arrange & Act
    const { getByText } = render(
      <CustomSelect
        options={OPTIONS}
        value=""
        onChange={() => {}}
        label="Format"
      />,
    );

    // Assert
    expect(getByText('Format')).toBeTruthy();
  });
});
