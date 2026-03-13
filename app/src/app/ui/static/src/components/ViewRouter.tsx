import { useRouter } from '../context/RouterContext';
import { JobFormProvider } from '../context/JobFormContext';
import { PageHeader } from './PageHeader';
import { Wizard } from './wizard';
import { SourceStep } from './steps/SourceStep';
import { TransformStep } from './steps/TransformStep';

/**
 * Renders the active view based on router state.
 * Step content will be replaced by D10-D13.
 */
export function ViewRouter() {
  const { activeView } = useRouter();

  switch (activeView) {
    case 'create-job':
      return (
        <div id="view-create-job">
          <JobFormProvider>
            <Wizard onSubmit={() => { /* D11 will wire submission */ }}>
              {/* Step 1: Source (D08) */}
              <SourceStep />

              {/* Step 2: Transform (D09) */}
              <TransformStep />

              {/* Step 3: Target (D10) */}
              <div>
                <p class="text-text-secondary text-sm">
                  Target configuration will be built in D10.
                </p>
              </div>

              {/* Step 4: Review (D11) */}
              <div>
                <p class="text-text-secondary text-sm">
                  Review &amp; submit will be built in D11.
                </p>
              </div>
            </Wizard>
          </JobFormProvider>
        </div>
      );

    case 'job-history':
      return (
        <div id="view-job-history">
          <PageHeader
            title="Job History"
            description="View all bombardment jobs and their statuses"
          />
          <div class="card-flat">
            <p class="text-text-secondary text-sm">
              Job history table will be built in D13.
            </p>
          </div>
        </div>
      );

    case 'job-progress':
      return (
        <div id="view-job-progress">
          <PageHeader
            title="Job Progress"
            description="Track the status of a running bombardment job"
          />
          <div class="card-flat">
            <p class="text-text-secondary text-sm">
              Job progress view will be built in D12.
            </p>
          </div>
        </div>
      );

    default:
      return null;
  }
}
