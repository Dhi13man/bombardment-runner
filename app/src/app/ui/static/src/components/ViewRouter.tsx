import { useRouter } from '../context/RouterContext';
import { PageHeader } from './PageHeader';
import { Wizard } from './wizard';

/**
 * Renders the active view based on router state.
 * Placeholder step content will be replaced by D08-D13.
 */
export function ViewRouter() {
  const { activeView } = useRouter();

  switch (activeView) {
    case 'create-job':
      return (
        <div id="view-create-job">
          <Wizard onSubmit={() => { /* D11 will wire this */ }}>
            {/* Step 1: Source (D08) */}
            <div>
              <p class="text-text-secondary text-sm">
                Source configuration will be built in D08.
              </p>
            </div>

            {/* Step 2: Transform (D09) */}
            <div>
              <p class="text-text-secondary text-sm">
                Transform configuration will be built in D09.
              </p>
            </div>

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
