import axios from 'axios';
import type {
  LoginRequest,
  LoginResponse,
  CreateStackRequest,
  UpdateStackRequest,
  CreateRunRequest,
  SyncRefsRequest,
  Stack,
  Run,
  RepoInfo,
  PaginatedResponse,
  ErrorResponse,
  Runner,
  CreateRunnerRequest,
  UpdateRunnerRequest,
  RunnerCreatedResponse,
  BackendSchema,
  BackendSchemas,
  User,
  CreateUserRequest,
  UpdateUserRequest,
  UserListResponse,
} from '../types';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  const csrfToken = sessionStorage.getItem('csrf_token');
  if (csrfToken && config.method !== 'get') {
    config.headers['X-CSRF-Token'] = csrfToken;
  }

  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      const isLoginPage = window.location.pathname === '/login';
      if (!isLoginPage) {
        localStorage.removeItem('token');
        localStorage.removeItem('refresh_token');
        sessionStorage.removeItem('csrf_token');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export const auth = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/login', data);
    localStorage.setItem('token', response.data.token);
    localStorage.setItem('refresh_token', response.data.refresh_token);
    return response.data;
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout');
    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
    sessionStorage.removeItem('csrf_token');
  },

  me: async (): Promise<LoginResponse['user']> => {
    const response = await api.get('/auth/me');
    return response.data;
  },

  getCsrfToken: async (): Promise<string> => {
    const response = await api.get<{ csrf_token: string }>('/auth/csrf');
    const token = response.data.csrf_token;
    sessionStorage.setItem('csrf_token', token);
    return token;
  },
};

export const stacks = {
  list: async (): Promise<Stack[]> => {
    const response = await api.get<Stack[]>('/stacks');
    return response.data;
  },

  get: async (id: string): Promise<Stack> => {
    const response = await api.get<Stack>(`/stacks/${id}`);
    return response.data;
  },

  getWithRefs: async (id: string): Promise<Stack> => {
    const response = await api.get<Stack>(`/stacks/${id}/refs`);
    return response.data;
  },

  create: async (data: CreateStackRequest): Promise<Stack> => {
    const response = await api.post<Stack>('/stacks', data);
    return response.data;
  },

  update: async (id: string, data: UpdateStackRequest): Promise<Stack> => {
    const response = await api.put<Stack>(`/stacks/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/stacks/${id}`);
  },

  validateRepo: async (url: string): Promise<RepoInfo> => {
    const response = await api.get<RepoInfo>('/stacks/validate-repo', { params: { url } });
    return response.data;
  },

  refetchRepo: async (id: string): Promise<RepoInfo> => {
    const response = await api.get<RepoInfo>(`/stacks/${id}/refetch-repo`);
    return response.data;
  },

  syncRefs: async (id: string, data: SyncRefsRequest): Promise<Stack> => {
    const response = await api.put<Stack>(`/stacks/${id}/refs`, data);
    return response.data;
  },

  getRuns: async (id: string): Promise<Run[]> => {
    const response = await api.get<Run[]>(`/stacks/${id}/runs`);
    return response.data;
  },
};

export const runs = {
  list: async (page = 1, pageSize = 20): Promise<PaginatedResponse<Run>> => {
    const response = await api.get<PaginatedResponse<Run>>('/runs', {
      params: { page, page_size: pageSize },
    });
    return response.data;
  },

  get: async (id: string): Promise<Run> => {
    const response = await api.get<Run>(`/runs/${id}`);
    return response.data;
  },

  create: async (data: CreateRunRequest): Promise<Run> => {
    const response = await api.post<Run>('/runs', data);
    return response.data;
  },

  assign: async (runId: string, runnerId: string): Promise<void> => {
    await api.post(`/runs/${runId}/assign`, { runner_id: runnerId });
  },

  approve: async (runId: string): Promise<void> => {
    await api.post(`/runs/${runId}/approve`);
  },

  reject: async (runId: string): Promise<void> => {
    await api.post(`/runs/${runId}/reject`);
  },

  cancel: async (runId: string): Promise<void> => {
    await api.post(`/runs/${runId}/cancel`);
  },

  wait: async (runId: string, timeout = 30): Promise<Run> => {
    const response = await api.get<Run>(`/runs/${runId}/wait`, {
      params: { timeout },
    });
    return response.data;
  },
};

export const runners = {
  list: async (): Promise<Runner[]> => {
    const response = await api.get<Runner[]>('/runners');
    return response.data;
  },

  get: async (id: string): Promise<Runner> => {
    const response = await api.get<Runner>(`/runners/${id}`);
    return response.data;
  },

  create: async (data: CreateRunnerRequest): Promise<RunnerCreatedResponse> => {
    const response = await api.post<RunnerCreatedResponse>('/runners', data);
    return response.data;
  },

  update: async (id: string, data: UpdateRunnerRequest): Promise<Runner> => {
    const response = await api.put<Runner>(`/runners/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/runners/${id}`);
  },

  getRuns: async (id: string): Promise<Run[]> => {
    const response = await api.get<Run[]>(`/runners/${id}/runs`);
    return response.data;
  },

  resetToken: async (id: string): Promise<RunnerCreatedResponse> => {
    const response = await api.post<RunnerCreatedResponse>(`/runners/${id}/reset-token`);
    return response.data;
  },
};

export const backends = {
  list: async (): Promise<unknown[]> => {
    const response = await api.get<unknown[]>('/backends');
    return response.data;
  },

  get: async (id: string): Promise<unknown> => {
    const response = await api.get<unknown>(`/backends/${id}`);
    return response.data;
  },

  create: async (data: { name: string; type: string; config: string }): Promise<unknown> => {
    const response = await api.post<unknown>('/backends', data);
    return response.data;
  },

  update: async (id: string, data: { name?: string; type?: string; config?: string }): Promise<unknown> => {
    const response = await api.put<unknown>(`/backends/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/backends/${id}`);
  },

  getSchemas: async (): Promise<BackendSchemas> => {
    const response = await api.get<BackendSchemas>('/backends/schemas');
    return response.data;
  },

  getSchema: async (type: string): Promise<BackendSchema> => {
    const response = await api.get<BackendSchema>(`/backends/schemas/${type}`);
    return response.data;
  },
};

export const users = {
  list: async (page = 1, pageSize = 20): Promise<UserListResponse> => {
    const response = await api.get<UserListResponse>('/users', {
      params: { page, page_size: pageSize },
    });
    return response.data;
  },

  get: async (id: string): Promise<User> => {
    const response = await api.get<User>(`/users/${id}`);
    return response.data;
  },

  create: async (data: CreateUserRequest): Promise<User> => {
    const response = await api.post<User>('/users', data);
    return response.data;
  },

  update: async (id: string, data: UpdateUserRequest): Promise<User> => {
    const response = await api.put<User>(`/users/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/users/${id}`);
  },

  resetPassword: async (id: string, newPassword: string): Promise<void> => {
    await api.post(`/users/${id}/reset-password`, { new_password: newPassword });
  },
};

export type { ErrorResponse };
