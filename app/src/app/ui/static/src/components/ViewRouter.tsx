import { useRouter } from '../context/RouterContext';
import { PageHeader } from './PageHeader';
import { CreateJobView } from './views/CreateJobView';

/**
 * Renders the active view based on router state.
 */
export function ViewRouter() {
  const { activeView } = useRouter();

  switch (activeView) {
    case 'create-job':
      return <CreateJobView />;

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
