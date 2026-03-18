import type { ComponentChildren } from 'preact';
import { toChildArray } from 'preact';
import { useEffect, useRef } from 'preact/hooks';
import { useWizard } from '../../context/WizardContext';

interface WizardPanelsProps {
  children: ComponentChildren[];
}

/**
 * Renders children as step panels, showing only the active one.
 * Manages focus to the first interactive element when the active step changes.
 */
export function WizardPanels({ children }: WizardPanelsProps) {
  const { currentStep } = useWizard();
  const panels = toChildArray(children);
  const prevStep = useRef(currentStep);

  // Focus management: move focus to first interactive element on step change
  useEffect(() => {
    if (prevStep.current !== currentStep) {
      prevStep.current = currentStep;
      requestAnimationFrame(() => {
        const panel = document.getElementById(`step-${currentStep}`);
        if (!panel) return;
        const focusable = panel.querySelector<HTMLElement>(
          'input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), button:not([disabled]), [tabindex]:not([tabindex="-1"])',
        );
        focusable?.focus();
      });
    }
  }, [currentStep]);

  return (
    <div class="step-panels-container">
      {panels.map((panel, i) => {
        const stepNum = i + 1;
        const isActive = stepNum === currentStep;

        return (
          <div
            key={stepNum}
            id={`step-${stepNum}`}
            class={`step-panel${isActive ? ' active' : ''}`}
            data-step={stepNum}
            role="region"
            aria-label={`Step ${stepNum}`}
            aria-hidden={!isActive ? 'true' : undefined}
          >
            {panel}
          </div>
        );
      })}
    </div>
  );
}
