import { useEffect, useState } from 'react';
import { runners } from '../lib/api';
import { useAuth } from '../hooks/useAuth';
import { Permissions } from '../lib/permissions';
import type { Runner, RunnerCreatedResponse } from '../types';

export default function RunnersList() {
  const [runnersList, setRunnersList] = useState<Runner[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [newRunner, setNewRunner] = useState({ name: '' });
  const [creating, setCreating] = useState(false);
  const [createdRunner, setCreatedRunner] = useState<RunnerCreatedResponse | null>(null);
  const { can } = useAuth();

  const canCreate = can(Permissions.RUNNER_CREATE);
  const canDelete = can(Permissions.RUNNER_DELETE);
  const canResetToken = can(Permissions.RUNNER_TOKEN);

  useEffect(() => {
    const fetchRunners = async () => {
      try {
        const data = await runners.list();
        setRunnersList(Array.isArray(data) ? data : []);
      } catch {
        setError('Failed to load runners');
      } finally {
        setLoading(false);
      }
    };
    fetchRunners();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreating(true);
    try {
      const created = await runners.create({ name: newRunner.name }) as RunnerCreatedResponse;
      setRunnersList([...runnersList, { id: created.id, name: created.name, status: 'offline', created_at: new Date().toISOString() }]);
      setNewRunner({ name: '' });
      setCreatedRunner(created);
      setShowModal(false);
    } catch {
      setError('Failed to create runner');
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this runner?')) return;
    try {
      await runners.delete(id);
      setRunnersList(runnersList.filter(r => r.id !== id));
    } catch {
      setError('Failed to delete runner');
    }
  };

  const handleResetToken = async (id: string) => {
    if (!confirm('Are you sure you want to reset the token? The runner will need to be restarted with the new token.')) return;
    try {
      const result = await runners.resetToken(id);
      setCreatedRunner(result);
    } catch {
      setError('Failed to reset token');
    }
  };

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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-title">Runners</h1>
          <p className="text-muted mt-1">Manage terraform runners</p>
        </div>
        {canCreate && (
          <button onClick={() => setShowModal(true)} className="btn btn-primary">
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add Runner
          </button>
        )}
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/50 text-red-600 dark:text-red-400 p-4 rounded-lg">
          {error}
        </div>
      )}

      {runnersList.length === 0 ? (
        <div className="card p-12 text-center">
          <div className="w-16 h-16 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2" />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-slate-900 dark:text-white mb-1">No runners yet</h3>
          <p className="text-muted">Add a runner to start executing terraform runs</p>
        </div>
      ) : (
        <div className="table-container">
<table className="table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Status</th>
                  <th>Current Run</th>
                  <th>Last Seen</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {runnersList.map((runner) => (
                  <tr key={runner.id}>
                    <td className="font-medium text-slate-900 dark:text-white">{runner.name}</td>
                    <td>
                      <span className={`badge ${
                        runner.status === 'online' ? 'badge-success' :
                        runner.status === 'busy' ? 'badge-warning' :
                        'badge-neutral'
                      }`}>
                        {runner.status}
                      </span>
                    </td>
                    <td className="font-mono text-sm">
                      {runner.current_run_id ? (
                        <a href={`/runs/${runner.current_run_id}`} className="text-primary-600 hover:underline">
                          {runner.current_run_id.slice(0, 8)}...
                        </a>
                      ) : (
                        <span className="text-muted">-</span>
                      )}
                    </td>
                    <td className="text-muted">
                      {runner.last_seen ? new Date(runner.last_seen).toLocaleString() : 'Never'}
                    </td>
                    <td>
                      <div className="flex items-center gap-3">
                        {canResetToken && (
                          <button
                            onClick={() => handleResetToken(runner.id)}
                            className="text-blue-600 hover:text-blue-800 dark:text-blue-400"
                            title="Reset token"
                          >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                            </svg>
                          </button>
                        )}
                        {canDelete && (
                          <button
                            onClick={() => handleDelete(runner.id)}
                            className="text-red-600 hover:text-red-800 dark:text-red-400"
                            disabled={runner.current_run_id !== undefined}
                            title={runner.current_run_id ? 'Cannot delete runner with active run' : 'Delete runner'}
                          >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
        </div>
      )}

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-slate-800 rounded-lg p-6 w-full max-w-md">
            <h2 className="text-lg font-medium mb-4">Add Runner</h2>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Name</label>
                <input
                  type="text"
                  value={newRunner.name}
                  onChange={(e) => setNewRunner({ name: e.target.value })}
                  className="input w-full"
                  required
                />
              </div>
              <div className="flex justify-end gap-3 pt-4">
                <button type="button" onClick={() => setShowModal(false)} className="btn btn-secondary">
                  Cancel
                </button>
                <button type="submit" disabled={creating} className="btn btn-primary">
                  {creating ? 'Creating...' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {createdRunner && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-slate-800 rounded-lg p-6 w-full max-w-md">
            <h2 className="text-lg font-medium mb-4 text-green-600 dark:text-green-400">Runner Created</h2>
            <div className="space-y-4">
              <p className="text-sm text-muted">Copy these credentials. You will not be able to see the token again.</p>
              <div>
                <label className="block text-sm font-medium mb-1">Runner ID</label>
                <code className="block bg-slate-100 dark:bg-slate-700 p-2 rounded text-sm font-mono break-all">
                  {createdRunner.id}
                </code>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Token</label>
                <code className="block bg-slate-100 dark:bg-slate-700 p-2 rounded text-sm font-mono break-all">
                  {createdRunner.token}
                </code>
              </div>
              <div className="flex justify-end">
                <button onClick={() => setCreatedRunner(null)} className="btn btn-primary">
                  Done
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}