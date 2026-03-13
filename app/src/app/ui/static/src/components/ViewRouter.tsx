import { useRouter } from '../context/RouterContext';
import { PageHeader } from './PageHeader';
import { CreateJobView } from './views/CreateJobView';
import { JobProgressView } from './views/JobProgressView';

/**
 * Renders the active view based on router state.
 */
export function ViewRouter() {
  const { activeView } = useRouter();

  switch (activeView) {
    case 'create-job':
      return <CreateJobView />;

    case 'job-progress':
      return <JobProgressView />;

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

    default:
      return null;
  }
}
