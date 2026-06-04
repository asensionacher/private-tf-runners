import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { stacks as stacksApi } from '../lib/api';
import { useAuth } from '../hooks/useAuth';
import { Permissions } from '../lib/permissions';
import type { Stack } from '../types';

export default function StacksList() {
  const [stacksList, setStacksList] = useState<Stack[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { can } = useAuth();

  const canCreate = can(Permissions.STACK_CREATE);

  useEffect(() => {
    const fetchStacks = async () => {
      try {
        const data = await stacksApi.list();
        setStacksList(data || []);
      } catch {
        setError('Failed to load stacks');
        setStacksList([]);
      } finally {
        setLoading(false);
      }
    };
    fetchStacks();
  }, []);

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

  const stacks = stacksList || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-title">Stacks</h1>
          <p className="text-muted mt-1">Manage your Terraform repositories</p>
        </div>
        {canCreate && (
          <Link to="/stacks/new" className="btn btn-primary">
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Create Stack
          </Link>
        )}
      </div>

      {stacks.length === 0 ? (
        <div className="card p-12 text-center">
          <div className="w-16 h-16 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-slate-900 dark:text-white mb-1">No stacks yet</h3>
          <p className="text-muted mb-6">Get started by creating your first stack</p>
          {canCreate ? (
            <Link to="/stacks/new" className="btn btn-primary">
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Create Your First Stack
            </Link>
          ) : (
            <p className="text-muted">No stacks have been created yet.</p>
          )}
        </div>
      ) : (
        <div className="table-container">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Provider</th>
                <th>Git Repository</th>
                <th>Published</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {stacks.map((stack) => (
                <tr key={stack.id}>
                  <td className="font-medium text-slate-900 dark:text-white">{stack.name}</td>
                  <td>
                    <span className="badge badge-primary uppercase text-xs">{stack.provider}</span>
                  </td>
                  <td className="font-mono text-sm text-muted max-w-xs truncate">{stack.git_url}</td>
                  <td className="text-muted">
                    {stack.published_branches.length + stack.published_tags.length} refs
                  </td>
                  <td className="text-right">
                    <div className="flex items-center justify-end gap-2">
                      <Link
                        to={`/runs/new?stack_id=${stack.id}`}
                        className="btn btn-primary btn-sm"
                        title="Run this stack"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                      </Link>
                      <Link
                        to={`/stacks/${stack.id}`}
                        className="btn btn-secondary btn-sm"
                      >
                        View
                      </Link>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}