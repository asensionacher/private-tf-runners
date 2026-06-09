import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import ansiToHtml from 'ansi-to-html';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { a11yLight } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import { runs, runners as runnersApi } from '../lib/api';
import type { Run, Runner } from '../types';

function AnsiOutput({ text, className = '' }: { text: string; className?: string }) {
  const converter = new ansiToHtml();
  const html = converter.toHtml(text);
  return (
    <pre
      className={`bg-slate-900 text-slate-100 p-4 rounded-lg overflow-x-auto text-xs font-mono max-h-[500px] overflow-y-auto ${className}`}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

function parseOutputs(run: Run) {
  const planOutput = run.plan_output || '';
  const applyOutput = run.apply_output || '';
  let tfOutputs: Record<string, { value: unknown; type: string }> = {};

  if (applyOutput) {
    const outputsMatch = applyOutput.match(/=== Terraform Outputs ===\n([\s\S]*?)$/);
    if (outputsMatch) {
      try {
        const parsed = JSON.parse(outputsMatch[1]);
        if (parsed && typeof parsed === 'object') {
          tfOutputs = parsed;
        }
      } catch {
      }
    }
  }

  return { planOutput, applyOutput, tfOutputs };
}

export default function RunDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [run, setRun] = useState<Run | null>(null);
  const [runnersList, setRunnersList] = useState<Runner[]>([]);
  const [error, setError] = useState('');
  const [actionLoading, setActionLoading] = useState(false);
  const [selectedRunner, setSelectedRunner] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(false);

  const loadRun = async () => {
    if (!id) return;
    try {
      const data = await runs.get(id);
      setRun(data);
      setError('');
      if (data.runner_id) {
        setSelectedRunner(data.runner_id);
      }
      if (isNonTerminalStatus(data.status)) {
        setAutoRefresh(true);
      } else {
        setAutoRefresh(false);
      }
    } catch {
      setError('Failed to load run');
      setAutoRefresh(false);
    }
  };

  useEffect(() => {
    loadRun();
    const loadRunners = async () => {
      try {
        const data = await runnersApi.list();
        setRunnersList(Array.isArray(data) ? data : []);
      } catch {
      }
    };
    loadRunners();
  }, [id]);

  const isNonTerminalStatus = (status: string): boolean => {
    return ['pending', 'running', 'planned', 'approved'].includes(status);
  };

  useEffect(() => {
    if (!autoRefresh || !run) return;

    const waitForUpdates = async () => {
      if (!id) return;
      try {
        const updatedRun = await runs.wait(id, 30);
        setRun(updatedRun);
        setError('');
        if (updatedRun.runner_id) {
          setSelectedRunner(updatedRun.runner_id);
        }
        if (isNonTerminalStatus(updatedRun.status)) {
          waitForUpdates();
        } else {
          setAutoRefresh(false);
        }
      } catch {
        setError('Failed to get run updates');
        setAutoRefresh(false);
      }
    };

    waitForUpdates();
  }, [autoRefresh, id, run?.status]);

  const handleAssignRunner = async () => {
    if (!id || !selectedRunner) return;
    setActionLoading(true);
    try {
      await runs.assign(id, selectedRunner);
      await loadRun();
    } catch {
      setError('Failed to assign runner');
    } finally {
      setActionLoading(false);
    }
  };

  const handleApprove = async () => {
    if (!id) return;
    setActionLoading(true);
    try {
      await runs.approve(id);
      await loadRun();
    } catch {
      setError('Failed to approve run');
    } finally {
      setActionLoading(false);
    }
  };

  const handleReject = async () => {
    if (!id) return;
    setActionLoading(true);
    try {
      await runs.reject(id);
      await loadRun();
    } catch {
      setError('Failed to reject run');
    } finally {
      setActionLoading(false);
    }
  };

  if (error && !run) {
    return (
      <div className="card p-8 text-center">
        <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 dark:bg-red-900/50 dark:text-red-400 flex items-center justify-center mx-auto mb-4">
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <p className="text-red-600 dark:text-red-400 font-medium">{error}</p>
        <button onClick={() => navigate('/runs')} className="btn btn-secondary mt-4">
          Back to Runs
        </button>
      </div>
    );
  }

  if (!run) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <svg className="w-8 h-8 animate-spin text-primary-600" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
      </div>
    );
  }

  const { planOutput, applyOutput, tfOutputs } = parseOutputs(run);
  const showApproveButton = (run.status === 'pending' || run.status === 'running' || run.status === 'planned') && run.phase === 'plan' && planOutput;
  const showAssignControl = run.status === 'pending' && !run.runner_id;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <button onClick={() => navigate('/runs')} className="btn btn-secondary btn-sm">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Back
        </button>
        <div>
          <h1 className="page-title">Run Details</h1>
          <p className="text-muted text-sm mt-1">ID: {run.id}</p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="card p-4">
          <span className="text-muted text-sm">Status</span>
          <p className={`font-medium ${
            run.status === 'applied' ? 'text-emerald-600 dark:text-emerald-400' :
            run.status === 'success' ? 'text-green-600 dark:text-green-400' :
            run.status === 'failed' ? 'text-red-600 dark:text-red-400' :
            run.status === 'running' ? 'text-blue-600 dark:text-blue-400' :
            run.status === 'planned' ? 'text-violet-600 dark:text-violet-400' :
            run.status === 'pending' ? 'text-amber-600 dark:text-amber-400' :
            run.status === 'approved' ? 'text-cyan-600 dark:text-cyan-400' :
            run.status === 'rejected' ? 'text-orange-600 dark:text-orange-400' :
            'text-slate-900 dark:text-slate-100'
          }`}>{run.status}</p>
        </div>
        <div className="card p-4">
          <span className="text-muted text-sm">Phase</span>
          <p className="font-medium capitalize">{run.phase || 'plan'}</p>
        </div>
        <div className="card p-4">
          <span className="text-muted text-sm">Branch</span>
          <p className="font-medium">{run.branch}</p>
        </div>
        <div className="card p-4">
          <span className="text-muted text-sm">Stack</span>
          <p className="font-medium">{run.stack_name || run.stack_id}</p>
        </div>
        {run.runner_name && (
          <div className="card p-4">
            <span className="text-muted text-sm">Runner</span>
            <p className="font-medium">{run.runner_name}</p>
          </div>
        )}
        <div className="card p-4">
          <span className="text-muted text-sm">Created</span>
          <p className="font-medium">{new Date(run.created_at).toLocaleString()}</p>
        </div>
      </div>

      {showAssignControl && (
        <div className="card p-6 space-y-4">
          <h2 className="section-title">Assign Runner</h2>
          <div className="flex gap-3">
            <select
              value={selectedRunner}
              onChange={(e) => setSelectedRunner(e.target.value)}
              className="input flex-1"
            >
              <option value="">Select a runner</option>
              {runnersList
                .filter(r => r.status === 'online')
                .map(runner => (
                  <option key={runner.id} value={runner.id}>{runner.name}</option>
                ))}
            </select>
            <button
              onClick={handleAssignRunner}
              disabled={!selectedRunner || actionLoading}
              className="btn btn-primary"
            >
              {actionLoading ? 'Assigning...' : 'Assign'}
            </button>
          </div>
        </div>
      )}

      {run.logs && (
        <div className="card p-6 space-y-4">
          <h2 className="section-title">Logs</h2>
          <pre className="bg-slate-900 text-slate-100 p-4 rounded-lg overflow-x-auto text-xs font-mono max-h-96">
            {run.logs}
          </pre>
        </div>
      )}

      {planOutput !== undefined && (
        <div className="card p-6 space-y-4">
          <h2 className="section-title">Init + Plan Output</h2>
          {planOutput ? (
            <AnsiOutput text={planOutput} />
          ) : (
            <div className="text-muted text-sm">Waiting for output...</div>
          )}
          {showApproveButton && (
            <div className="flex gap-3 justify-end border-t border-slate-200 dark:border-slate-700 pt-4">
              <button
                onClick={handleReject}
                disabled={actionLoading}
                className="btn btn-secondary"
              >
                Reject
              </button>
              <button
                onClick={handleApprove}
                disabled={actionLoading}
                className="btn btn-primary"
              >
                {actionLoading ? 'Approving...' : 'Approve Plan'}
              </button>
            </div>
          )}
        </div>
      )}

      {applyOutput !== undefined && (
        <div className="card p-6 space-y-4">
          <h2 className="section-title">Apply Output</h2>
          {applyOutput ? (
            <AnsiOutput text={applyOutput} />
          ) : (
            <div className="text-muted text-sm">Waiting for apply output...</div>
          )}
        </div>
      )}

      {tfOutputs && Object.keys(tfOutputs).length > 0 && (
        <div className="card p-6 space-y-4">
          <h2 className="section-title">Terraform Outputs</h2>
          <SyntaxHighlighter
            language="json"
            style={a11yLight}
            customStyle={{
              backgroundColor: '#f8fafc',
              borderRadius: '0.5rem',
              padding: '1rem',
              fontSize: '0.75rem',
              maxHeight: '500px',
              overflow: 'auto',
            }}
          >
            {JSON.stringify(tfOutputs, null, 2)}
          </SyntaxHighlighter>
        </div>
      )}
    </div>
  );
}
