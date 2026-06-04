import { useState, useEffect, type FormEvent } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { stacks, runs, backends, runners } from '../lib/api';
import { useAuth } from '../hooks/useAuth';
import { Permissions } from '../lib/permissions';
import type { Stack, CreateRunRequest, BackendSchemas, BackendSchema, BackendField, AuthMethod, Runner } from '../types';

type BackendType = keyof BackendSchemas;

export default function CreateRunPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const preselectedStackId = searchParams.get('stack_id');

  const [stacksList, setStacksList] = useState<Stack[]>([]);
  const [backendSchemas, setBackendSchemas] = useState<BackendSchemas>({});
  const [runnersList, setRunnersList] = useState<Runner[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const [selectedStack, setSelectedStack] = useState<Stack | null>(null);
  const [selectedBranch, setSelectedBranch] = useState('');
  const [selectedRunner, setSelectedRunner] = useState<string>('');
  const [backendType, setBackendType] = useState<BackendType | ''>('');
  const [selectedAuthMethod, setSelectedAuthMethod] = useState<string>('');
  const [backendConfig, setBackendConfig] = useState<Record<string, string>>({});
  const [selectedTfvarsFiles, setSelectedTfvarsFiles] = useState<string[]>([]);
  const [tfvarsFileInput, setTfvarsFileInput] = useState('');
  const [tfvarsInline, setTfvarsInline] = useState('');
  const [createdRun] = useState<{ id: string; commands: string } | null>(null);
  const { can } = useAuth();
  const canCreateRun = can(Permissions.RUN_CREATE);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [stacksData, schemasData, runnersData] = await Promise.all([
        stacks.list(),
        backends.getSchemas(),
        runners.list(),
      ]);

      const stacksWithRefs: Stack[] = [];
      for (const stack of stacksData || []) {
        try {
          const fullStack = await stacks.getWithRefs(stack.id);
          stacksWithRefs.push(fullStack);
        } catch {
          stacksWithRefs.push({ ...stack, published_branches: stack.published_branches, published_tags: stack.published_tags } as Stack);
        }
      }
      setStacksList(stacksWithRefs);
      setBackendSchemas(schemasData);
      setRunnersList(runnersData || []);

      if (preselectedStackId) {
        const preselected = stacksWithRefs.find(s => s.id === preselectedStackId);
        if (preselected) {
          setSelectedStack(preselected);
        }
      }
    } catch (err) {
      setError('Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const handleStackChange = async (stackId: string) => {
    const stack = stacksList.find(s => s.id === stackId);
    if (stack) {
      setSelectedStack(stack);
      setSelectedBranch('');
      setSelectedRunner('');
      setBackendType('');
      setSelectedAuthMethod('');
      setBackendConfig({});
      setSelectedTfvarsFiles([]);
      setTfvarsFileInput('');
      setTfvarsInline('');
    }
  };

  const handleBackendTypeChange = (newType: BackendType | '') => {
    setBackendType(newType);
    setSelectedAuthMethod('');
    setBackendConfig({});
  };

  const getCurrentSchema = (): BackendSchema | null => {
    if (!backendType || !backendSchemas[backendType]) return null;
    return backendSchemas[backendType] as BackendSchema;
  };

  const getRequiredFields = (): BackendField[] => {
    const schema = getCurrentSchema();
    return schema?.required_fields || [];
  };

  const getOptionalFields = (): BackendField[] => {
    const schema = getCurrentSchema();
    return schema?.optional_fields || [];
  };

  const getAuthMethods = (): AuthMethod[] => {
    const schema = getCurrentSchema();
    return schema?.auth_methods || [];
  };

  const getSelectedAuthMethodFields = (): BackendField[] => {
    const methods = getAuthMethods();
    const selected = methods.find(m => m.name === selectedAuthMethod);
    return selected?.fields || [];
  };

  const getAllDisplayFields = (): Array<{ key: string; field: BackendField; source: string }> => {
    const fields: Array<{ key: string; field: BackendField; source: string }> = [];
    const required = getRequiredFields();
    for (const f of required) {
      fields.push({ key: f.key, field: f, source: 'required' });
    }
    const optional = getOptionalFields();
    for (const f of optional) {
      fields.push({ key: f.key, field: f, source: 'optional' });
    }
    const authFields = getSelectedAuthMethodFields();
    for (const f of authFields) {
      fields.push({ key: f.key, field: f, source: 'auth' });
    }
    return fields;
  };

  const addTfvarsFile = () => {
    const file = tfvarsFileInput.trim();
    if (file && !selectedTfvarsFiles.includes(file)) {
      setSelectedTfvarsFiles([...selectedTfvarsFiles, file]);
      console.log("Added tfvars file:", file, "state now:", [...selectedTfvarsFiles, file]);
    }
    setTfvarsFileInput('');
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!selectedStack || !selectedBranch) return;

    setSubmitting(true);
    setError('');

    try {
      const tfvarsValues: Record<string, string> = {};
      tfvarsInline.split('\n').forEach(line => {
        const trimmed = line.trim();
        if (!trimmed) return;
        const eqIndex = trimmed.indexOf('=');
        if (eqIndex > 0) {
          const key = trimmed.substring(0, eqIndex);
          const value = trimmed.substring(eqIndex + 1);
          if (key) tfvarsValues[key] = value;
        }
      });

      const backendConfigEnv: Record<string, string> = {};
      const backendConfigNonSensitive: Record<string, string> = {};

      const displayFields = getAllDisplayFields();
      for (const { key, field } of displayFields) {
        const value = backendConfig[key];
        if (value && value.trim() !== '') {
          if (field.env_var) {
            backendConfigEnv[field.env_var] = value;
          } else {
            backendConfigNonSensitive[key] = value;
          }
        }
      }

      const missingRequired = getRequiredFields().filter(f => !backendConfig[f.key] || backendConfig[f.key].trim() === '');
      if (missingRequired.length > 0) {
        setError(`Missing required fields: ${missingRequired.map(f => f.key).join(', ')}`);
        setSubmitting(false);
        return;
      }

      const req: CreateRunRequest = {
        stack_id: selectedStack.id,
        runner_id: selectedRunner,
        branch: selectedBranch,
        commit_sha: selectedBranch,
        backend_type: backendType || undefined,
        backend_config: Object.keys(backendConfigNonSensitive).length > 0
          ? Object.fromEntries(Object.entries(backendConfigNonSensitive).filter(([_, v]) => v !== ''))
          : undefined,
        backend_config_env: Object.keys(backendConfigEnv).length > 0 ? backendConfigEnv : undefined,
        tfvars_files: selectedTfvarsFiles.length > 0 ? selectedTfvarsFiles : undefined,
        tfvars_values: Object.keys(tfvarsValues).length > 0 ? tfvarsValues : undefined,
      };
      console.log("Submitting with tfvars_files:", selectedTfvarsFiles);

      const run = await runs.create(req);
      navigate(`/runs/${run.id}`);
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'response' in err) {
        const axiosError = err as { response?: { data?: { error?: string } } };
        setError(axiosError.response?.data?.error || 'Failed to create run');
      } else {
        setError('Failed to create run');
      }
    } finally {
      setSubmitting(false);
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

  if (createdRun) {
    return (
      <div className="max-w-3xl mx-auto">
        <div className="card p-8">
          <div className="flex items-center gap-4 mb-6">
            <div className="w-12 h-12 rounded-full bg-emerald-100 text-emerald-600 flex items-center justify-center">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div>
              <h2 className="text-xl font-semibold text-slate-900 dark:text-white">Run Created!</h2>
              <p className="text-muted">ID: {createdRun.id}</p>
            </div>
          </div>

          <div className="mb-6">
            <h3 className="section-title mb-3">Commands to Execute</h3>
            <p className="text-sm text-muted mb-3">
              Copy and run these commands on a machine with {selectedStack?.provider === 'terraform' ? 'Terraform' : 'OpenTofu'} installed:
            </p>
            <pre className="bg-slate-900 text-slate-100 p-4 rounded-lg overflow-x-auto text-sm font-mono">
              {createdRun.commands}
            </pre>
          </div>

          <div className="flex justify-end gap-3">
            <button onClick={() => navigate('/runs')} className="btn btn-secondary">
              View All Runs
            </button>
            <button onClick={() => navigate(`/runs/${createdRun.id}`)} className="btn btn-primary">
              View Run Details
            </button>
          </div>
        </div>
      </div>
    );
  }

  const stacksWithRefs = stacksList;

  const allBranches = selectedStack
    ? [...(selectedStack.published_branches || []), ...(selectedStack.published_tags || [])]
    : [];

  const hasTfvarsFile = selectedStack?.tfvars_files && selectedStack.tfvars_files.length > 0;

  const currentSchema = getCurrentSchema();
  const authMethods = getAuthMethods();

  if (!canCreateRun) {
    return (
      <div className="max-w-3xl mx-auto">
        <div className="card p-12 text-center">
          <div className="w-16 h-16 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-slate-900 dark:text-white mb-1">Permission Denied</h3>
          <p className="text-muted mb-6">You don't have permission to create runs.</p>
          <a href="/" className="btn btn-secondary">Back to Dashboard</a>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto">
      <div className="mb-6">
        <h1 className="page-title">Create Run</h1>
        <p className="text-muted mt-1">Configure and execute a Terraform deployment</p>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl dark:bg-red-900/20 dark:border-red-800">
          <span className="text-sm text-red-600 dark:text-red-400">{error}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="card p-6 space-y-6">
          <h2 className="section-title">Stack Configuration</h2>

          <div>
            <label htmlFor="stack" className="label">Select Stack *</label>
            <select
              id="stack"
              value={selectedStack?.id || ''}
              onChange={(e) => handleStackChange(e.target.value)}
              className="input"
              required
            >
              <option value="">Choose a stack...</option>
              {stacksWithRefs.map((stack) => (
                <option key={stack.id} value={stack.id}>
                  {stack.name} ({stack.provider})
                </option>
              ))}
            </select>
          </div>

          {selectedStack && (
            <div>
              <label htmlFor="branch" className="label">Branch or Tag *</label>
              <select
                id="branch"
                value={selectedBranch}
                onChange={(e) => setSelectedBranch(e.target.value)}
                className="input"
                required
              >
                <option value="">Choose a branch or tag...</option>
                {allBranches.map((ref) => (
                  <option key={ref} value={ref}>{ref}</option>
                ))}
              </select>
            </div>
          )}

          {selectedStack && (
            <div>
              <label htmlFor="runner" className="label">Runner *</label>
              <select
                id="runner"
                value={selectedRunner}
                onChange={(e) => setSelectedRunner(e.target.value)}
                className="input"
                required
              >
                <option value="">Choose a runner...</option>
                {runnersList.map((runner) => (
                  <option key={runner.id} value={runner.id}>
                    {runner.name} ({runner.status})
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>

        {selectedStack && (
          <>
            <div className="card p-6 space-y-6">
              <h2 className="section-title">Backend Configuration</h2>

              <div>
                <label htmlFor="backendType" className="label">Backend Type</label>
                <select
                  id="backendType"
                  value={backendType}
                  onChange={(e) => handleBackendTypeChange(e.target.value as BackendType | '')}
                  className="input"
                >
                  <option value="">Local state (no backend)</option>
                  <option value="azurerm">Azure</option>
                  <option value="s3">S3 (AWS)</option>
                  <option value="pg">PostgreSQL</option>
                  <option value="http">HTTP</option>
                </select>
                {currentSchema && (
                  <p className="mt-1 text-sm text-muted">
                    {currentSchema.description}
                  </p>
                )}
              </div>

              {currentSchema && (
                <>
                  {authMethods.length > 0 && (
                    <div>
                      <label htmlFor="authMethod" className="label">Authentication Method</label>
                      <select
                        id="authMethod"
                        value={selectedAuthMethod}
                        onChange={(e) => {
                          setSelectedAuthMethod(e.target.value);
                          const newConfig: Record<string, string> = {};
                          for (const key of Object.keys(backendConfig)) {
                            const field = [...getRequiredFields(), ...getOptionalFields(), ...getSelectedAuthMethodFields()].find(f => f.key === key);
                            if (!field) delete newConfig[key];
                          }
                          setBackendConfig(newConfig);
                        }}
                        className="input"
                      >
                        <option value="">Select an authentication method...</option>
                        {authMethods.map((method) => (
                          <option key={method.name} value={method.name}>
                            {method.description} ({method.name})
                          </option>
                        ))}
                      </select>
                    </div>
                  )}

                  <div className="space-y-4">
                    <h3 className="text-sm font-medium text-slate-700 dark:text-slate-300">Backend Arguments</h3>
                    {getAllDisplayFields().map(({ key, field }) => (
                      <div key={key}>
                        <label htmlFor={`backend-${key}`} className="label flex items-center gap-2">
                          <span className="font-mono">{key}</span>
                          {field.required && <span className="badge badge-error text-xs">Required</span>}
                          {field.sensitive && <span className="badge badge-warning text-xs">Sensitive</span>}
                          {field.env_var && <span className="badge badge-info text-xs">Env: {field.env_var}</span>}
                        </label>
                        {field.sensitive ? (
                          <input
                            id={`backend-${key}`}
                            type="password"
                            value={backendConfig[key] || ''}
                            onChange={(e) => setBackendConfig({ ...backendConfig, [key]: e.target.value })}
                            className="input font-mono text-sm"
                            placeholder={field.env_var ? `Will be set as ${field.env_var}` : 'Value'}
                          />
                        ) : (
                          <input
                            id={`backend-${key}`}
                            type="text"
                            value={backendConfig[key] || ''}
                            onChange={(e) => setBackendConfig({ ...backendConfig, [key]: e.target.value })}
                            className="input font-mono text-sm"
                            placeholder={field.default ? `Default: ${field.default}` : ''}
                          />
                        )}
                        <p className="mt-1 text-xs text-muted">
                          {field.description}
                          {field.default && <span className="ml-2 text-slate-400">Default: {field.default}</span>}
                        </p>
                      </div>
                    ))}
                  </div>
                </>
              )}
            </div>

            <div className="card p-6 space-y-6">
              <h2 className="section-title">Variables</h2>

              <div>
                <label htmlFor="tfvarsFiles" className="label">Terraform Variables Files</label>
                <p className="text-muted text-sm mb-3">
                  Select one or more tfvars files. These will be passed with -var-file:
                </p>
                {hasTfvarsFile && (
                  <div className="flex flex-wrap gap-2 mb-4">
                    {selectedStack.tfvars_files.map((file: string) => (
                      <label
                        key={file}
                        className={`flex items-center gap-2 px-3 py-2 rounded-lg border cursor-pointer transition-colors ${
                          selectedTfvarsFiles.includes(file)
                            ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                            : 'border-slate-300 dark:border-slate-600 hover:border-primary-400'
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={selectedTfvarsFiles.includes(file)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setSelectedTfvarsFiles([...selectedTfvarsFiles, file]);
                            } else {
                              setSelectedTfvarsFiles(selectedTfvarsFiles.filter(f => f !== file));
                            }
                          }}
                          className="w-4 h-4 text-primary-600 rounded border-slate-300 focus:ring-primary-500"
                        />
                        <span className="font-mono text-sm">{file}</span>
                      </label>
                    ))}
                  </div>
                )}
                {selectedTfvarsFiles.length > 0 && (
                  <div className="flex flex-wrap gap-2 mb-4">
                    {selectedTfvarsFiles.map((file: string) => (
                      <span
                        key={file}
                        className="flex items-center gap-1 px-3 py-1 rounded-lg bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 font-mono text-sm"
                      >
                        {file}
                        <button
                          type="button"
                          onClick={() => setSelectedTfvarsFiles(selectedTfvarsFiles.filter(f => f !== file))}
                          className="ml-1 text-primary-500 hover:text-primary-700 dark:hover:text-primary-400"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                          </svg>
                        </button>
                      </span>
                    ))}
                  </div>
                )}
                <div className="flex gap-2">
                  <input
                    type="text"
                    id="tfvarsFiles"
                    value={tfvarsFileInput}
                    onChange={(e) => setTfvarsFileInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        e.preventDefault();
                        addTfvarsFile();
                      }
                    }}
                    className="input font-mono text-sm flex-1"
                    placeholder="prod.tfvars"
                  />
                  <button
                    type="button"
                    onClick={addTfvarsFile}
                    className="btn btn-secondary"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                    Add
                  </button>
                </div>
              </div>

              <div>
                <label htmlFor="tfvarsInline" className="label">Inline Terraform Variables</label>
                <p className="text-muted text-sm mb-3">
                  Enter variables in KEY=value format, one per line:
                </p>
                <textarea
                  id="tfvarsInline"
                  value={tfvarsInline}
                  onChange={(e) => setTfvarsInline(e.target.value)}
                  className="input font-mono text-sm h-32"
                  placeholder={"DB_HOST=localhost\nDB_PORT=5432\nAWS_REGION=us-west-2"}
                />
                <p className="mt-2 text-sm text-muted">
                  These will be passed to terraform plan/apply using -var arguments.
                </p>
              </div>
            </div>
          </>
        )}

        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate('/runs')}
            className="btn btn-secondary"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={submitting || !selectedStack || !selectedBranch || !selectedRunner}
            className="btn btn-primary"
          >
            {submitting ? (
              <>
                <svg className="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Creating...
              </>
            ) : (
              'Create Run'
            )}
          </button>
        </div>
      </form>
    </div>
  );
}