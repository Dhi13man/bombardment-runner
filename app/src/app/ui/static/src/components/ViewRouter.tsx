import { useRouter } from '../context/RouterContext';
import { CreateJobView } from './views/CreateJobView';
import { JobProgressView } from './views/JobProgressView';
import { JobHistoryView } from './views/JobHistoryView';

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
      return <JobHistoryView />;
    default:
      return null;
  }
}
