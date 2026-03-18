import { useRouter } from '../context/RouterContext';
import { CreateJobView } from './views/CreateJobView';
import { JobProgressView } from './views/JobProgressView';
import { JobHistoryView } from './views/JobHistoryView';

/**
 * Renders the active view based on router state.
 * Uses key to trigger fade-in animation on view change.
 */
export function ViewRouter() {
  const { activeView } = useRouter();

  let content;
  switch (activeView) {
    case 'create-job':
      content = <CreateJobView />;
      break;
    case 'job-progress':
      content = <JobProgressView />;
      break;
    case 'job-history':
      content = <JobHistoryView />;
      break;
    default:
      return null;
  }

  return (
    <div key={activeView} class="view-enter">
      {content}
    </div>
  );
}
