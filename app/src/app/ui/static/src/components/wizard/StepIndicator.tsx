import { useWizard } from '../../context/WizardContext';
import { Icon } from '../Icon';

export function StepIndicator() {
  const { steps, currentStep } = useWizard();

  return (
    <ol class="steps" aria-label="Job creation progress">
      {steps.map((step, i) => {
        const stepNum = i + 1;
        const isActive = stepNum === currentStep;
        const isCompleted = stepNum < currentStep;

        const classes = [
          'step',
          isActive && 'active',
          isCompleted && 'completed',
        ]
          .filter(Boolean)
          .join(' ');

        return (
          <li
            key={step.label}
            class={classes}
            aria-current={isActive ? 'step' : undefined}
          >
            <div class="step-trigger">
              <div class="step-circle">
                {isCompleted ? (
                  <Icon name="check" size="sm" />
                ) : (
                  stepNum
                )}
              </div>
              <span class="step-label">{step.label}</span>
            </div>
          </li>
        );
      })}
    </ol>
  );
}
