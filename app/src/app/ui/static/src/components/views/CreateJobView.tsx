import { useState, useCallback } from 'preact/hooks';
import { JobFormProvider, useJobForm } from '../../context/JobFormContext';
import { useRouter } from '../../context/RouterContext';
import { useToast } from '../composites/Toast';
import { Wizard } from '../wizard';
import { SourceStep } from '../steps/SourceStep';
import { TransformStep } from '../steps/TransformStep';
import { TargetStep } from '../steps/TargetStep';
import { ReviewStep } from '../steps/ReviewStep';
import { createJob, ApiError } from '../../api/client';

/**
 * Inner component that has access to JobFormContext.
 * Handles submission via the Wizard's onSubmit callback.
 */
function CreateJobWizard() {
  const { toRequest, reset } = useJobForm();
  const { navigateTo } = useRouter();
  const { showToast } = useToast();
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = useCallback(async () => {
    if (submitting) return;
    setSubmitting(true);
    try {
      const payload = toRequest();
      const job = await createJob(payload);
      showToast('success', `Job ${job.id.substring(0, 8)}... created successfully`);
      reset();
      navigateTo('job-progress', { jobId: job.id });
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Failed to create job';
      showToast('error', message);
    } finally {
      setSubmitting(false);
    }
  }, [submitting, toRequest, showToast, navigateTo, reset]);

  return (
    <Wizard onSubmit={handleSubmit}>
      <SourceStep />
      <TransformStep />
      <TargetStep />
      <ReviewStep />
    </Wizard>
  );
}

/**
 * Complete Create Job view with all required providers.
 */
export function CreateJobView() {
  return (
    <div id="view-create-job">
      <JobFormProvider>
        <CreateJobWizard />
      </JobFormProvider>
    </div>
  );
}
