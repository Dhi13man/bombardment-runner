import { useRef, useCallback } from 'preact/hooks';
import { useWizard } from '../../context/WizardContext';
import { useToast } from '../composites/Toast';
import { Button } from '../primitives';
import { Icon } from '../Icon';

export function WizardNav() {
  const { isFirstStep, isLastStep, stepValid, currentStep, next, back } =
    useWizard();
  const { showToast } = useToast();
  const nextRef = useRef<HTMLButtonElement>(null);
  const submitRef = useRef<HTMLButtonElement>(null);

  const isCurrentValid = stepValid[currentStep];

  const handleNext = useCallback(() => {
    if (!isCurrentValid) {
      // Visual shake + accessible announcement
      const el = isLastStep ? submitRef.current : nextRef.current;
      if (el) {
        el.classList.add('shake');
        setTimeout(() => el.classList.remove('shake'), 400);
      }
      showToast('warning', 'Please complete required fields before proceeding');
      return;
    }
    next();
  }, [isCurrentValid, isLastStep, next, showToast]);

  return (
    <div class="wizard-nav">
      {!isFirstStep ? (
        <Button variant="secondary" onClick={back} title="Go back (Alt+Left)">
          <Icon name="arrow-left" size="sm" />
          Back
        </Button>
      ) : (
        <div />
      )}

      <div class="wizard-nav-end">
        {!isLastStep ? (
          <button
            ref={nextRef}
            type="button"
            class="btn btn-primary"
            disabled={!isCurrentValid}
            onClick={handleNext}
            title="Next step (Alt+Enter)"
          >
            Next
            <Icon name="arrow-right" size="sm" />
          </button>
        ) : (
          <button
            ref={submitRef}
            type="button"
            class="btn btn-primary"
            disabled={!isCurrentValid}
            onClick={handleNext}
          >
            <Icon name="rocket" size="sm" />
            Run Bombardment
          </button>
        )}
      </div>
    </div>
  );
}
