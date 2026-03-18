import { useWizard } from '../../context/WizardContext';
import { Icon } from '../Icon';

export function StepIndicator() {
  const { steps, currentStep, goTo } = useWizard();

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

        const circle = (
          <div class="step-circle">
            {isCompleted ? (
              <Icon name="check" size="sm" />
            ) : (
              <Icon name={step.icon} size="sm" />
            )}
          </div>
        );

        return (
          <li
            key={step.label}
            class={classes}
            aria-current={isActive ? 'step' : undefined}
          >
            {isCompleted ? (
              <button
                type="button"
                class="step-trigger"
                onClick={() => goTo(stepNum)}
                aria-label={`Go back to ${step.label}`}
              >
                {circle}
                <span class="step-label">{step.label}</span>
              </button>
            ) : (
              <div class="step-trigger">
                {circle}
                <span class="step-label">{step.label}</span>
              </div>
            )}
          </li>
        );
      })}
    </ol>
  );
}
