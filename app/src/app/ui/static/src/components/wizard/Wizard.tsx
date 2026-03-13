import type { ComponentChildren } from 'preact';
import { WizardProvider } from '../../context/WizardContext';
import { PageHeader } from '../PageHeader';
import { StepIndicator } from './StepIndicator';
import { WizardNav } from './WizardNav';
import { WizardPanels } from './WizardPanels';

interface WizardProps {
  /** Slot content for each step panel (indexed 0..3). */
  children: ComponentChildren[];
  onSubmit?: () => void;
}

/**
 * Top-level Wizard component.
 * Wraps WizardProvider, StepIndicator, step panels, and navigation.
 */
export function Wizard({ children, onSubmit }: WizardProps) {
  return (
    <WizardProvider onSubmit={onSubmit}>
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
