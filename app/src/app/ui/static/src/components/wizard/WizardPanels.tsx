import type { ComponentChildren } from 'preact';
import { toChildArray } from 'preact';
import { useWizard } from '../../context/WizardContext';

interface WizardPanelsProps {
  children: ComponentChildren[];
}

/**
 * Renders children as step panels, showing only the active one.
 */
export function WizardPanels({ children }: WizardPanelsProps) {
  const { currentStep } = useWizard();
  const panels = toChildArray(children);

  return (
    <>
      {panels.map((panel, i) => {
        const stepNum = i + 1;
        const isActive = stepNum === currentStep;

        return (
          <div
            key={stepNum}
            id={`step-${stepNum}`}
            class={`step-panel${isActive ? ' active' : ''}`}
            data-step={stepNum}
            role="tabpanel"
            aria-hidden={!isActive}
          >
            {panel}
          </div>
        );
      })}
    </>
  );
}
