import { createContext } from 'preact';
import { useState, useCallback, useContext, useMemo, useRef } from 'preact/hooks';
import type { ComponentChildren } from 'preact';
import type { IconName } from '../components/Icon';

/* ---------- Types ---------- */

export interface StepDefinition {
  label: string;
  icon: IconName;
}

interface WizardContextValue {
  steps: StepDefinition[];
  currentStep: number;
  totalSteps: number;
  /** Per-step validation state. */
  stepValid: Record<number, boolean>;
  /** Navigate to the next step (if current step is valid). Returns false if validation fails. */
  next: () => boolean;
  /** Navigate to the previous step. */
  back: () => void;
  /** Go to a specific step (1-indexed). */
  goTo: (step: number) => void;
  /** Mark a step as valid or invalid. */
  setValid: (step: number, valid: boolean) => void;
  /** Whether we're on the final step. */
  isLastStep: boolean;
  /** Whether we're on the first step. */
  isFirstStep: boolean;
}

/* ---------- Steps ---------- */

const WIZARD_STEPS: StepDefinition[] = [
  { label: 'Source', icon: 'file-input' },
  { label: 'Transform', icon: 'sliders-horizontal' },
  { label: 'Target', icon: 'target' },
  { label: 'Review', icon: 'check-circle-2' },
];

/* ---------- Context ---------- */

const WizardContext = createContext<WizardContextValue | null>(null);

interface WizardProviderProps {
  children: ComponentChildren;
  /** Called when the submit button is clicked on the final step. */
  onSubmit?: () => void;
}

export function WizardProvider({ children, onSubmit }: WizardProviderProps) {
  const [currentStep, setCurrentStep] = useState(1);
  const [stepValid, setStepValid] = useState<Record<number, boolean>>({
    1: false,
    2: false,
    3: false,
    4: false,
  });

  const totalSteps = WIZARD_STEPS.length;

  // Refs for stable callbacks
  const stepValidRef = useRef(stepValid);
  stepValidRef.current = stepValid;
  const currentStepRef = useRef(currentStep);
  currentStepRef.current = currentStep;
  const onSubmitRef = useRef(onSubmit);
  onSubmitRef.current = onSubmit;

  const setValid = useCallback((step: number, valid: boolean) => {
    setStepValid((prev) => ({ ...prev, [step]: valid }));
  }, []);

  const next = useCallback(() => {
    const cs = currentStepRef.current;
    if (!stepValidRef.current[cs]) return false;
    if (cs < totalSteps) {
      setCurrentStep((s) => s + 1);
      return true;
    }
    if (cs === totalSteps && onSubmitRef.current) {
      onSubmitRef.current();
      return true;
    }
    return false;
  }, [totalSteps]);

  const back = useCallback(() => {
    setCurrentStep((s) => (s > 1 ? s - 1 : s));
  }, []);

  const goTo = useCallback(
    (step: number) => {
      if (step >= 1 && step <= totalSteps) {
        setCurrentStep(step);
      }
    },
    [totalSteps],
  );

  const value = useMemo<WizardContextValue>(
    () => ({
      steps: WIZARD_STEPS,
      currentStep,
      totalSteps,
      stepValid,
      next,
      back,
      goTo,
      setValid,
      isLastStep: currentStep === totalSteps,
      isFirstStep: currentStep === 1,
    }),
    [currentStep, totalSteps, stepValid, next, back, goTo, setValid],
  );

  return (
    <WizardContext.Provider value={value}>{children}</WizardContext.Provider>
  );
}

export function useWizard(): WizardContextValue {
  const ctx = useContext(WizardContext);
  if (!ctx) {
    throw new Error('useWizard must be used within a WizardProvider');
  }
  return ctx;
}
