import { useEffect, useState } from 'react';
import { runs } from '../lib/api';
import type { Run } from '../types';

export default function RunsList() {
  const [runsList, setRunsList] = useState<Run[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    const fetchRuns = async () => {
      setLoading(true);
      try {
        const data = await runs.list(page, 20);
        setRunsList(Array.isArray(data.data) ? data.data : []);
        setTotalPages(typeof data.total_pages === 'number' ? data.total_pages : 1);
      } catch {
        setError('Failed to load runs');
        setRunsList([]);
      } finally {
        setLoading(false);
      }
    };
    fetchRuns();
  }, [page]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <svg className="w-8 h-8 animate-spin text-primary-600" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
      </div>
    );
  }

  if (error) {
    return (
      <div className="card p-8 text-center">
        <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 dark:bg-red-900/50 dark:text-red-400 flex items-center justify-center mx-auto mb-4">
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <p className="text-red-600 dark:text-red-400 font-medium">{error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="page-title">Runs</h1>
        <p className="text-muted mt-1">View and manage Terraform executions</p>
      </div>

      {runsList.length === 0 ? (
        <div className="card p-12 text-center">
          <div className="w-16 h-16 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-slate-900 dark:text-white mb-1">No runs yet</h3>
          <p className="text-muted">Create a stack to start running Terraform</p>
        </div>
      ) : (
        <>
          <div className="table-container">
            <table className="table">
              <thead>
                <tr>
                  <th>Stack</th>
                  <th>Branch</th>
                  <th>Status</th>
                  <th>Phase</th>
                  <th>Runner</th>
                  <th>Trigger</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {runsList.map((run) => (
                  <tr key={run.id} className="hover:bg-slate-50 dark:hover:bg-slate-800 cursor-pointer" onClick={() => window.location.href = `/runs/${run.id}`}>
                    <td className="font-medium text-slate-900 dark:text-white">{run.stack_name || run.stack_id}</td>
                    <td className="font-mono text-sm">{run.branch}</td>
                    <td>
                      <span className={`badge ${
                        run.status === 'applied' ? 'badge-applied' :
                        run.status === 'success' ? 'badge-success' :
                        run.status === 'failed' ? 'badge-danger' :
                        run.status === 'planned' ? 'badge-planned' :
                        run.status === 'rejected' ? 'badge-rejected' :
                        run.status === 'running' ? 'badge-info' :
                        run.status === 'approved' ? 'badge-success' :
                        'badge-warning'
                      }`}>
                        {run.status}
                      </span>
                    </td>
                    <td className="capitalize">{run.phase || 'plan'}</td>
                    <td className="font-mono text-sm">
                      {run.runner_name || (run.runner_id ? run.runner_id.slice(0, 8) + '...' : '-')}
                    </td>
                    <td className="text-muted capitalize">{run.trigger_type}</td>
                    <td className="text-muted">
                      {new Date(run.created_at).toLocaleDateString('en-US', {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="btn btn-secondary btn-sm"
              >
                <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
                Previous
              </button>
              <span className="px-4 py-2 text-sm text-muted">
                Page {page} of {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="btn btn-secondary btn-sm"
              >
                Next
                <svg className="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}