import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { users as usersApi } from '../lib/api';
import { useAuth } from '../hooks/useAuth';
import { Permissions } from '../lib/permissions';
import type { User, UserRole, CreateUserRequest } from '../types';

export default function UsersList() {
  const navigate = useNavigate();
  const [usersList, setUsersList] = useState<User[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [newPassword, setNewPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [resettingPassword, setResettingPassword] = useState(false);
  const [creating, setCreating] = useState(false);
  const [formData, setFormData] = useState<CreateUserRequest>({
    username: '',
    email: '',
    password: '',
    role: 'viewer',
  });
  const [formError, setFormError] = useState('');
  const { can } = useAuth();

  useEffect(() => {
    if (!can(Permissions.USER_ADMIN)) {
      navigate('/');
    }
  }, [can, navigate]);

  if (!can(Permissions.USER_ADMIN)) {
    return null;
  }

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const data = await usersApi.list();
        setUsersList(data.data || []);
      } catch {
        setError('Failed to load users');
        setUsersList([]);
      } finally {
        setLoading(false);
      }
    };
    fetchUsers();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');
    setCreating(true);

    try {
      await usersApi.create(formData);
      const data = await usersApi.list();
      setUsersList(data.data || []);
      setShowCreateModal(false);
      setFormData({ username: '', email: '', password: '', role: 'viewer' });
    } catch (err: unknown) {
      const errorMessage = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Failed to create user';
      setFormError(errorMessage);
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id: string, username: string) => {
    if (!confirm(`Delete user "${username}"?`)) return;
    try {
      await usersApi.delete(id);
      const data = await usersApi.list();
      setUsersList(data.data || []);
    } catch {
      alert('Failed to delete user');
    }
  };

  const handleResetPassword = async (user: User) => {
    setSelectedUser(user);
    setNewPassword('');
    setPasswordError('');
    setShowPasswordModal(true);
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedUser) return;

    if (newPassword.length < 8) {
      setPasswordError('Password must be at least 8 characters');
      return;
    }

    setResettingPassword(true);
    setPasswordError('');

    try {
      await usersApi.resetPassword(selectedUser.id, newPassword);
      setShowPasswordModal(false);
      setSelectedUser(null);
      setNewPassword('');
    } catch (err: unknown) {
      const errorMessage = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Failed to reset password';
      setPasswordError(errorMessage);
    } finally {
      setResettingPassword(false);
    }
  };

  const getRoleBadgeClass = (role: UserRole) => {
    switch (role) {
      case 'admin':
        return 'badge-danger';
      case 'operator':
        return 'badge-warning';
      case 'viewer':
        return 'badge-secondary';
      default:
        return 'badge-secondary';
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

  const users = usersList || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-title">Users</h1>
          <p className="text-muted mt-1">Manage user accounts and permissions</p>
        </div>
        <button onClick={() => setShowCreateModal(true)} className="btn btn-primary">
          <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Create User
        </button>
      </div>

      {users.length === 0 ? (
        <div className="card p-12 text-center">
          <div className="w-16 h-16 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-slate-900 dark:text-white mb-1">No users yet</h3>
          <p className="text-muted mb-6">Create your first user to get started</p>
          <button onClick={() => setShowCreateModal(true)} className="btn btn-primary">
            Create Your First User
          </button>
        </div>
      ) : (
        <div className="table-container">
          <table className="table">
            <thead>
              <tr>
                <th>Username</th>
                <th>Email</th>
                <th>Role</th>
                <th>Created</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id}>
                  <td className="font-medium text-slate-900 dark:text-white">{user.username}</td>
                  <td className="text-muted">{user.email}</td>
                  <td>
                    <span className={`badge ${getRoleBadgeClass(user.role as UserRole)} uppercase text-xs`}>
                      {user.role}
                    </span>
                  </td>
                  <td className="text-muted text-sm">{new Date(user.created_at).toLocaleDateString()}</td>
                  <td className="text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => handleResetPassword(user)}
                        className="btn btn-ghost btn-sm text-blue-600 hover:text-blue-700"
                        title="Reset password"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                        </svg>
                      </button>
                      <button
                        onClick={() => handleDelete(user.id, user.username)}
                        className="btn btn-ghost btn-sm text-red-600 hover:text-red-700"
                        title="Delete user"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-slate-800 rounded-lg shadow-xl w-full max-w-md p-6">
            <h2 className="text-xl font-bold text-slate-900 dark:text-white mb-4">Create User</h2>

            {formError && (
              <div className="mb-4 p-3 bg-red-100 text-red-700 rounded-lg text-sm">{formError}</div>
            )}

            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                  Username
                </label>
                <input
                  type="text"
                  required
                  minLength={3}
                  maxLength={50}
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  className="input w-full"
                  placeholder="Enter username"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                  Email
                </label>
                <input
                  type="email"
                  required
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="input w-full"
                  placeholder="user@example.com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                  Password
                </label>
                <input
                  type="password"
                  required
                  minLength={8}
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="input w-full"
                  placeholder="Minimum 8 characters"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                  Role
                </label>
                <select
                  value={formData.role}
                  onChange={(e) => setFormData({ ...formData, role: e.target.value as UserRole })}
                  className="input w-full"
                >
                  <option value="viewer">Viewer - Read-only access</option>
                  <option value="operator">Operator - Can manage stacks and runs</option>
                  <option value="admin">Admin - Full access including user management</option>
                </select>
              </div>

              <div className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-3 text-sm">
                <h4 className="font-medium text-slate-700 dark:text-slate-300 mb-2">Role Permissions:</h4>
                <ul className="text-slate-600 dark:text-slate-400 space-y-1">
                  <li><span className="font-medium">Viewer:</span> Read stacks, runs, runners</li>
                  <li><span className="font-medium">Operator:</span> Create/delete stacks and runs, manage runners</li>
                  <li><span className="font-medium">Admin:</span> All operator permissions + user management</li>
                </ul>
              </div>

              <div className="flex justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowCreateModal(false);
                    setFormError('');
                  }}
                  className="btn btn-secondary"
                >
                  Cancel
                </button>
                <button type="submit" disabled={creating} className="btn btn-primary">
                  {creating ? 'Creating...' : 'Create User'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showPasswordModal && selectedUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-slate-800 rounded-lg shadow-xl w-full max-w-md p-6">
            <h2 className="text-xl font-bold text-slate-900 dark:text-white mb-4">Reset Password</h2>
            <p className="text-muted mb-4">Enter a new password for user: <strong>{selectedUser.username}</strong></p>

            {passwordError && (
              <div className="mb-4 p-3 bg-red-100 text-red-700 rounded-lg text-sm">{passwordError}</div>
            )}

            <form onSubmit={handlePasswordSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                  New Password
                </label>
                <input
                  type="password"
                  required
                  minLength={8}
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="input w-full"
                  placeholder="Minimum 8 characters"
                />
              </div>

              <div className="flex justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowPasswordModal(false);
                    setSelectedUser(null);
                    setPasswordError('');
                  }}
                  className="btn btn-secondary"
                >
                  Cancel
                </button>
                <button type="submit" disabled={resettingPassword} className="btn btn-primary">
                  {resettingPassword ? 'Resetting...' : 'Reset Password'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}