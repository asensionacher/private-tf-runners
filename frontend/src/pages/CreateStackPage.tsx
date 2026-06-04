import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { stacks } from '../lib/api';
import type { Provider, RepoInfo } from '../types';

export default function CreateStackPage() {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [gitUrl, setGitUrl] = useState('');
  const [gitFolder, setGitFolder] = useState('');
  const [provider, setProvider] = useState<Provider>('opentofu');
  const [repoInfo, setRepoInfo] = useState<RepoInfo | null>(null);
  const [selectedBranches, setSelectedBranches] = useState<string[]>([]);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [error, setError] = useState('');

  const handleValidateRepo = async () => {
    if (!gitUrl) return;
    setError('');
    setFetching(true);
    try {
      const info = await stacks.validateRepo(gitUrl);
      setRepoInfo(info);
    } catch {
      setError('Failed to fetch repository. Please check the URL and try again.');
    } finally {
      setFetching(false);
    }
  };

  const toggleBranch = (branch: string) => {
    setSelectedBranches((prev) =>
      prev.includes(branch) ? prev.filter((b) => b !== branch) : [...prev, branch]
    );
  };

  const toggleTag = (tag: string) => {
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]
    );
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!name || !gitUrl) return;
    if (repoInfo && selectedBranches.length === 0 && selectedTags.length === 0) {
      setError('Select at least one branch or tag');
      return;
    }
    setError('');
    setLoading(true);
    try {
      await stacks.create({
        name,
        description,
        git_url: gitUrl,
        git_folder: gitFolder,
        provider,
        branch: selectedBranches.join(','),
        tags: selectedTags.join(','),
      });
      navigate('/stacks');
    } catch {
      setError('Failed to create stack');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6">
        <h1 className="page-title">Create Stack</h1>
        <p className="text-muted mt-1">Add a new Terraform repository</p>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl dark:bg-red-900/20 dark:border-red-800">
          <div className="flex items-center gap-2 text-red-600 dark:text-red-400">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm font-medium">{error}</span>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="card p-6 space-y-6">
        <div>
          <label htmlFor="name" className="label">Name *</label>
          <input
            id="name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="input"
            placeholder="my-awesome-stack"
            required
            maxLength={100}
          />
        </div>

        <div>
          <label htmlFor="description" className="label">Description</label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="input min-h-[80px] resize-none"
            placeholder="What does this stack do?"
            maxLength={500}
          />
        </div>

        <div>
          <label htmlFor="gitUrl" className="label">Git Repository URL *</label>
          <div className="flex gap-3">
            <input
              id="gitUrl"
              type="url"
              value={gitUrl}
              onChange={(e) => {
                setGitUrl(e.target.value);
                setRepoInfo(null);
              }}
              className="input flex-1"
              placeholder="https://github.com/org/repo.git"
              required
            />
            <button
              type="button"
              onClick={handleValidateRepo}
              disabled={!gitUrl || fetching}
              className="btn btn-secondary whitespace-nowrap"
            >
              {fetching ? (
                <>
                  <svg className="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Fetching...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Fetch Refs
                </>
              )}
            </button>
          </div>
        </div>

        <div>
          <label htmlFor="gitFolder" className="label">Git Folder</label>
          <input
            id="gitFolder"
            type="text"
            value={gitFolder}
            onChange={(e) => setGitFolder(e.target.value)}
            className="input"
            placeholder="Leave empty for root"
          />
          <p className="mt-1 text-sm text-muted">Subdirectory containing Terraform files</p>
        </div>

        <div>
          <label className="label">Provider *</label>
          <div className="flex gap-4">
            <label className="flex items-center gap-3 p-4 border border-slate-200 dark:border-slate-700 rounded-lg cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors has-[:checked]:border-primary-500 has-[:checked]:bg-primary-50 dark:has-[:checked]:bg-primary-900/20">
              <input
                type="radio"
                name="provider"
                value="opentofu"
                checked={provider === 'opentofu'}
                onChange={() => setProvider('opentofu')}
                className="w-4 h-4 text-primary-600"
              />
              <div>
                <span className="font-medium text-slate-900 dark:text-white">OpenTofu</span>
                <p className="text-sm text-muted">Open source Terraform alternative</p>
              </div>
            </label>
            <label className="flex items-center gap-3 p-4 border border-slate-200 dark:border-slate-700 rounded-lg cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors has-[:checked]:border-primary-500 has-[:checked]:bg-primary-50 dark:has-[:checked]:bg-primary-900/20">
              <input
                type="radio"
                name="provider"
                value="terraform"
                checked={provider === 'terraform'}
                onChange={() => setProvider('terraform')}
                className="w-4 h-4 text-primary-600"
              />
              <div>
                <span className="font-medium text-slate-900 dark:text-white">Terraform</span>
                <p className="text-sm text-muted">HashiCorp Terraform</p>
              </div>
            </label>
          </div>
        </div>

        {repoInfo && (
          <div className="border-t border-slate-200 dark:border-slate-700 pt-6 space-y-6">
            <h3 className="font-semibold text-slate-900 dark:text-white">Select References to Publish</h3>

            <div>
              <div className="flex items-center justify-between mb-3">
                <label className="label mb-0">Branches</label>
                <button
                  type="button"
                  onClick={() => setSelectedBranches(repoInfo.branches)}
                  className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400"
                >
                  Select all
                </button>
              </div>
              {repoInfo.branches.length === 0 ? (
                <p className="text-muted text-sm">No branches found</p>
              ) : (
                <div className="flex flex-wrap gap-2 max-h-40 overflow-y-auto scrollbar-thin p-1">
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
                <label className="label mb-0">Tags</label>
                <button
                  type="button"
                  onClick={() => setSelectedTags(repoInfo.tags)}
                  className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400"
                >
                  Select all
                </button>
              </div>
              {repoInfo.tags.length === 0 ? (
                <p className="text-muted text-sm">No version tags found</p>
              ) : (
                <div className="flex flex-wrap gap-2 max-h-40 overflow-y-auto scrollbar-thin p-1">
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

            {selectedBranches.length === 0 && selectedTags.length === 0 && (
              <p className="text-sm text-amber-600 dark:text-amber-400">
                Select at least one branch or tag to publish
              </p>
            )}
          </div>
        )}

        <div className="flex justify-end gap-3 pt-4 border-t border-slate-200 dark:border-slate-700">
          <button
            type="button"
            onClick={() => navigate('/stacks')}
            className="btn btn-secondary"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading || !name || !gitUrl || !!(repoInfo && selectedBranches.length === 0 && selectedTags.length === 0)}
            className="btn btn-primary"
          >
            {loading ? (
              <>
                <svg className="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Creating...
              </>
            ) : (
              'Create Stack'
            )}
          </button>
        </div>
      </form>
    </div>
  );
}