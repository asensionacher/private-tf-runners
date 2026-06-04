import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { stacks } from '../lib/api';
import { useAuth } from '../hooks/useAuth';
import { Permissions } from '../lib/permissions';
import type { Stack, RepoInfo } from '../types';

export default function StackDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [stack, setStack] = useState<Stack | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [showRefetchModal, setShowRefetchModal] = useState(false);
  const [refetching, setRefetching] = useState(false);
  const [repoInfo, setRepoInfo] = useState<RepoInfo | null>(null);
  const [selectedBranches, setSelectedBranches] = useState<string[]>([]);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [syncing, setSyncing] = useState(false);
  const [syncError, setSyncError] = useState('');
  const { can } = useAuth();

  const canDelete = can(Permissions.STACK_DELETE);
  const canUpdate = can(Permissions.STACK_CREATE);

  const fetchStack = async () => {
    if (!id) return;
    try {
      const data = await stacks.get(id);
      setStack(data);
      setSelectedBranches(data.published_branches || []);
      setSelectedTags(data.published_tags || []);
    } catch {
      setError('Failed to load stack');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStack();
  }, [id]);

  const handleDelete = async () => {
    if (!id) return;
    setDeleting(true);
    try {
      await stacks.delete(id);
      navigate('/stacks');
    } catch {
      setError('Failed to delete stack');
    } finally {
      setDeleting(false);
    }
  };

  const handleRefetch = async () => {
    if (!id) return;
    setRefetching(true);
    setSyncError('');
    try {
      const info = await stacks.refetchRepo(id);
      setRepoInfo(info);
      setSelectedBranches(info.branches.filter(b => stack?.published_branches.includes(b)));
      setSelectedTags(info.tags.filter(t => stack?.published_tags.includes(t)));
      setShowRefetchModal(true);
    } catch {
      setSyncError('Failed to refetch repository');
    } finally {
      setRefetching(false);
    }
  };

  const toggleBranch = (branch: string) => {
    setSelectedBranches(prev =>
      prev.includes(branch) ? prev.filter(b => b !== branch) : [...prev, branch]
    );
  };

  const toggleTag = (tag: string) => {
    setSelectedTags(prev =>
      prev.includes(tag) ? prev.filter(t => t !== tag) : [...prev, tag]
    );
  };

  const handleSync = async () => {
    if (!id) return;
    setSyncing(true);
    setSyncError('');
    try {
      await stacks.syncRefs(id, {
        branches: selectedBranches,
        tags: selectedTags,
      });
      await fetchStack();
      setShowRefetchModal(false);
    } catch {
      setSyncError('Failed to sync refs');
    } finally {
      setSyncing(false);
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

  if (error && !stack) {
    return (
      <div className="card p-8 text-center">
        <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 dark:bg-red-900/50 dark:text-red-400 flex items-center justify-center mx-auto mb-4">
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <p className="text-red-600 dark:text-red-400 font-medium">{error}</p>
        <Link to="/stacks" className="btn btn-secondary mt-4">Back to Stacks</Link>
      </div>
    );
  }

  if (!stack) {
    return (
      <div className="text-center py-12">
        <p className="text-muted">Stack not found</p>
        <Link to="/stacks" className="btn btn-secondary mt-4">Back to Stacks</Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/stacks" className="btn btn-ghost p-2">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </Link>
          <div>
            <h1 className="page-title">{stack.name}</h1>
            <p className="text-muted mt-1">{stack.description || 'No description'}</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          {canUpdate && (
            <button
              onClick={handleRefetch}
              disabled={refetching}
              className="btn btn-secondary"
            >
              {refetching ? (
                <>
                  <svg className="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Refetching...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Refetch & Manage
                </>
              )}
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="card p-6">
          <h2 className="section-title mb-4">Details</h2>
          <dl className="space-y-4">
            <div>
              <dt className="text-sm text-muted">Provider</dt>
              <dd className="mt-1">
                <span className="badge badge-primary uppercase">{stack.provider}</span>
              </dd>
            </div>
            <div>
              <dt className="text-sm text-muted">Git Folder</dt>
              <dd className="mt-1 text-sm font-mono text-slate-900 dark:text-white">
                {stack.git_folder || '/ (root)'}
              </dd>
            </div>
            <div>
              <dt className="text-sm text-muted">Git Repository</dt>
              <dd className="mt-1 text-sm font-mono text-slate-900 dark:text-white break-all">
                {stack.git_url}
              </dd>
            </div>
            <div>
              <dt className="text-sm text-muted">Created</dt>
              <dd className="mt-1 text-sm text-slate-900 dark:text-white">
                {new Date(stack.created_at).toLocaleDateString('en-US', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </dd>
            </div>
          </dl>
        </div>

        <div className="card p-6">
          <h2 className="section-title mb-4">Published Branches</h2>
          {stack.published_branches.length === 0 ? (
            <p className="text-muted text-sm">No branches published</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {stack.published_branches.map((branch) => (
                <span
                  key={branch}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300 rounded-lg text-sm font-mono"
                >
                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  {branch}
                </span>
              ))}
            </div>
          )}
        </div>

        <div className="card p-6">
          <h2 className="section-title mb-4">Published Tags</h2>
          {stack.published_tags.length === 0 ? (
            <p className="text-muted text-sm">No tags published</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {stack.published_tags.map((tag) => (
                <span
                  key={tag}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300 rounded-lg text-sm font-mono"
                >
                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                  </svg>
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>

      {canDelete && (
        <div className="flex justify-end">
          <button
            onClick={() => setShowDeleteModal(true)}
            className="btn btn-danger"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Delete Stack
          </button>
        </div>
      )}

      {showDeleteModal && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="card p-6 max-w-md w-full">
            <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 dark:bg-red-900/50 dark:text-red-400 flex items-center justify-center mx-auto mb-4">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-center text-slate-900 dark:text-white mb-2">Delete Stack</h3>
            <p className="text-muted text-center mb-6">
              Are you sure you want to delete "{stack.name}"? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowDeleteModal(false)}
                className="btn btn-secondary"
                disabled={deleting}
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                className="btn btn-danger"
                disabled={deleting}
              >
                {deleting ? 'Deleting...' : 'Delete Stack'}
              </button>
            </div>
          </div>
        </div>
      )}

      {showRefetchModal && repoInfo && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="card p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-semibold text-slate-900 dark:text-white">Manage References</h3>
              <button
                onClick={() => setShowRefetchModal(false)}
                className="btn btn-ghost p-2"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {syncError && (
              <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:border-red-800">
                <span className="text-sm text-red-600 dark:text-red-400">{syncError}</span>
              </div>
            )}

            <div className="space-y-6">
              <div>
                <div className="flex items-center justify-between mb-3">
                  <label className="font-medium text-slate-900 dark:text-white">Branches</label>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => setSelectedBranches(repoInfo.branches)}
                      className="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400"
                    >
                      Select all
                    </button>
                    <button
                      type="button"
                      onClick={() => setSelectedBranches([])}
                      className="text-xs text-muted hover:text-slate-600"
                    >
                      Clear
                    </button>
                  </div>
                </div>
                {repoInfo.branches.length === 0 ? (
                  <p className="text-muted text-sm">No branches found</p>
                ) : (
                  <div className="flex flex-wrap gap-2 max-h-48 overflow-y-auto scrollbar-thin p-1">
                    {repoInfo.branches.map((branch) => (
                      <label
                        key={branch}
                        className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg cursor-pointer transition-colors ${
                          selectedBranches.includes(branch)
                            ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/50 dark:text-primary-300'
                            : 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700'
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={selectedBranches.includes(branch)}
                          onChange={() => toggleBranch(branch)}
                          className="sr-only"
                        />
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                        {branch}
                      </label>
                    ))}
                  </div>
                )}
              </div>

              <div>
                <div className="flex items-center justify-between mb-3">
                  <label className="font-medium text-slate-900 dark:text-white">Tags</label>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => setSelectedTags(repoInfo.tags)}
                      className="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400"
                    >
                      Select all
                    </button>
                    <button
                      type="button"
                      onClick={() => setSelectedTags([])}
                      className="text-xs text-muted hover:text-slate-600"
                    >
                      Clear
                    </button>
                  </div>
                </div>
                {repoInfo.tags.length === 0 ? (
                  <p className="text-muted text-sm">No version tags found</p>
                ) : (
                  <div className="flex flex-wrap gap-2 max-h-48 overflow-y-auto scrollbar-thin p-1">
                    {repoInfo.tags.map((tag) => (
                      <label
                        key={tag}
                        className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg cursor-pointer transition-colors ${
                          selectedTags.includes(tag)
                            ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300'
                            : 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700'
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={selectedTags.includes(tag)}
                          onChange={() => toggleTag(tag)}
                          className="sr-only"
                        />
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                        </svg>
                        {tag}
                      </label>
                    ))}
                  </div>
                )}
              </div>
            </div>

            <div className="flex justify-end gap-3 mt-8 pt-6 border-t border-slate-200 dark:border-slate-700">
              <button
                onClick={() => setShowRefetchModal(false)}
                className="btn btn-secondary"
                disabled={syncing}
              >
                Cancel
              </button>
              <button
                onClick={handleSync}
                className="btn btn-primary"
                disabled={syncing || (selectedBranches.length === 0 && selectedTags.length === 0)}
              >
                {syncing ? (
                  <>
                    <svg className="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    Saving...
                  </>
                ) : (
                  'Save Changes'
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}