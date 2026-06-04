import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { stacks, runs } from '../lib/api';
import type { Run } from '../types';

function StatCard({ title, value, icon, color }: { title: string; value: number | string; icon: React.ReactNode; color: string }) {
  return (
    <div className="card stat-card p-6 group">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-muted text-sm font-medium">{title}</p>
          <p className="text-3xl font-bold text-slate-900 dark:text-white mt-2">{value}</p>
        </div>
        <div className={`w-12 h-12 rounded-xl ${color} flex items-center justify-center group-hover:scale-110 transition-transform duration-200`}>
          {icon}
        </div>
      </div>
    </div>
  );
}

export default function Dashboard() {
  const [stacksCount, setStacksCount] = useState(0);
  const [recentRuns, setRecentRuns] = useState<Run[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const stacksList = await stacks.list();
        setStacksCount(Array.isArray(stacksList) ? stacksList.length : 0);

        const runsData = await runs.list(1, 5);
        setRecentRuns(Array.isArray(runsData.data) ? runsData.data : []);
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
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

  return (
    <div className="space-y-8">
      <div>
        <h1 className="page-title">Dashboard</h1>
        <p className="text-muted mt-1">Overview of your infrastructure</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Stacks"
          value={stacksCount}
          color="bg-primary-100 text-primary-600 dark:bg-primary-900/50 dark:text-primary-400"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          }
        />
        <StatCard
          title="Recent Runs"
          value={recentRuns.length}
          color="bg-emerald-100 text-emerald-600 dark:bg-emerald-900/50 dark:text-emerald-400"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          }
        />
        <Link to="/runs/new" className="card stat-card p-6 group hover:border-primary-300 dark:hover:border-primary-700 transition-colors">
          <div className="flex items-center justify-center h-full py-4">
            <div className="text-center">
              <div className="w-12 h-12 rounded-xl bg-emerald-100 text-emerald-600 dark:bg-emerald-900/50 dark:text-emerald-400 flex items-center justify-center mx-auto mb-3 group-hover:scale-110 transition-transform duration-200">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <p className="font-medium text-slate-900 dark:text-white">Create Run</p>
              <p className="text-sm text-muted mt-1">Deploy a stack</p>
            </div>
          </div>
        </Link>
        <Link to="/stacks/new" className="card stat-card p-6 group hover:border-primary-300 dark:hover:border-primary-700 transition-colors">
          <div className="flex items-center justify-center h-full py-4">
            <div className="text-center">
              <div className="w-12 h-12 rounded-xl bg-primary-100 text-primary-600 dark:bg-primary-900/50 dark:text-primary-400 flex items-center justify-center mx-auto mb-3 group-hover:scale-110 transition-transform duration-200">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
              </div>
              <p className="font-medium text-slate-900 dark:text-white">Create Stack</p>
              <p className="text-sm text-muted mt-1">Add a new repository</p>
            </div>
          </div>
        </Link>
      </div>

      <div className="card overflow-hidden">
        <div className="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <h2 className="section-title">Recent Activity</h2>
          <Link to="/runs" className="text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
            View all runs
          </Link>
        </div>
        {recentRuns.length === 0 ? (
          <div className="p-8 text-center">
            <div className="w-16 h-16 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <p className="text-muted">No runs yet</p>
            <p className="text-sm text-muted mt-1">Create a stack to start running Terraform</p>
          </div>
        ) : (
          <div className="table-container !border-0 !rounded-none">
            <table className="table">
              <thead>
                <tr>
                  <th>Stack</th>
                  <th>Branch</th>
                  <th>Status</th>
                  <th>Trigger</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {recentRuns.map((run) => (
                  <tr key={run.id}>
                    <td className="font-medium text-slate-900 dark:text-white">{run.stack_name || run.stack_id}</td>
                    <td className="font-mono text-sm">{run.branch}</td>
                    <td>
                      <span className={`badge ${
                        run.status === 'success' ? 'badge-success' :
                        run.status === 'failed' ? 'badge-danger' :
                        run.status === 'running' ? 'badge-info' :
                        'badge-warning'
                      }`}>
                        {run.status}
                      </span>
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
        )}
      </div>
    </div>
  );
}