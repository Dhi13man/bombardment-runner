import type { ComponentChildren } from 'preact';
import { useEffect } from 'preact/hooks';
import { WizardProvider, useWizard } from '../../context/WizardContext';
import { PageHeader } from '../PageHeader';
import { StepIndicator } from './StepIndicator';
import { WizardNav } from './WizardNav';
import { WizardPanels } from './WizardPanels';

interface WizardProps {
  /** Slot content for each step panel (indexed 0..3). */
  children: ComponentChildren[];
  onSubmit?: () => void;
}

function WizardKeyboardShortcuts() {
  const { next, back, goTo } = useWizard();

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || tag === 'BUTTON') return;

      if (e.altKey && e.key === 'Enter') { e.preventDefault(); next(); }
      if (e.altKey && e.key === 'ArrowLeft') { e.preventDefault(); back(); }
      if (['1', '2', '3', '4'].includes(e.key) && e.altKey && e.shiftKey) {
        e.preventDefault();
        goTo(Number(e.key));
      }
    }
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [next, back, goTo]);

  return null;
}

/**
 * Top-level Wizard component.
 * Wraps WizardProvider, StepIndicator, step panels, and navigation.
 */
export function Wizard({ children, onSubmit }: WizardProps) {
  return (
    <WizardProvider onSubmit={onSubmit}>
      <WizardKeyboardShortcuts />
      <PageHeader
        title="Create Bombardment Job"
        description="Configure your data migration pipeline in 4 steps"
      />
      <StepIndicator />
      <div class="card" style={{ marginTop: '24px' }}>
        <WizardPanels>{children}</WizardPanels>
        <WizardNav />
      </div>
    </WizardProvider>
  );
}
