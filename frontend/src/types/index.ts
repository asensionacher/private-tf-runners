export type Provider = 'opentofu' | 'terraform';
export type TriggerType = 'manual' | 'push' | 'schedule';
export type RunStatus = 'pending' | 'running' | 'success' | 'failed' | 'approved' | 'planned' | 'applied' | 'rejected';
export type RunPhase = 'plan' | 'apply' | 'finish';
export type RunnerStatus = 'online' | 'offline' | 'busy';
export type UserRole = 'admin' | 'operator' | 'viewer';

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  provider: Provider;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
  last_login_ip?: string;
  two_factor_enabled: boolean;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role: UserRole;
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  role?: UserRole;
}

export interface UserListResponse {
  data: User[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface Stack {
  id: string;
  name: string;
  description: string;
  git_url: string;
  git_folder: string;
  provider: Provider;
  published_branches: string[];
  published_tags: string[];
  tfvars_files: string[];
  created_at: string;
  updated_at: string;
}

export interface CreateStackRequest {
  name: string;
  description: string;
  git_url: string;
  git_folder: string;
  provider: Provider;
  branch?: string;
  tags?: string;
}

export interface UpdateStackRequest {
  name?: string;
  description?: string;
  git_url?: string;
  git_folder?: string;
  provider?: Provider;
  published_branches?: string[];
  published_tags?: string[];
}

export interface Run {
  id: string;
  stack_id: string;
  stack_name?: string;
  trigger_type: TriggerType;
  branch: string;
  commit_sha: string;
  status: RunStatus;
  phase?: RunPhase;
  logs?: string;
  plan_output?: string;
  apply_output?: string;
  backend_type?: string;
  backend_key?: string;
  tfvars_files?: string[];
  tfvars_values?: string;
  env_vars?: string;
  commands?: string;
  runner_id?: string;
  runner_name?: string;
  created_at: string;
  started_at?: string;
  finished_at?: string;
}

export interface Runner {
  id: string;
  name: string;
  status: RunnerStatus;
  current_run_id?: string;
  last_seen?: string;
  created_at: string;
}

export interface CreateRunnerRequest {
  name: string;
}

export interface UpdateRunnerRequest {
  name?: string;
  status?: string;
}

export interface RunnerCreatedResponse {
  id: string;
  name: string;
  token: string;
}

export interface RepoInfo {
  branches: string[];
  tags: string[];
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  refresh_token: string;
  expires_at: number;
  user: User;
}

export interface SyncRefsRequest {
  branches: string[];
  tags: string[];
}

export interface CreateRunRequest {
  stack_id: string;
  runner_id: string;
  branch: string;
  commit_sha: string;
  backend_type?: string;
  backend_config?: Record<string, string>;
  backend_config_env?: Record<string, string>;
  tfvars_files?: string[];
  tfvars_values?: Record<string, string>;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ErrorResponse {
  error: string;
  code?: string;
  details?: string;
}

export interface CreateRunnerRequest {
  name: string;
  url?: string;
}

export interface UpdateRunnerRequest {
  name?: string;
  url?: string;
  status?: string;
}

export interface RunnerCreatedResponse {
  id: string;
  name: string;
  token: string;
}

export interface BackendField {
  key: string;
  description: string;
  required: boolean;
  sensitive: boolean;
  env_var?: string;
  default?: string;
}

export interface AuthMethod {
  name: string;
  description: string;
  fields: BackendField[];
}

export interface BackendSchema {
  name: string;
  description: string;
  auth_methods: AuthMethod[];
  common_fields?: BackendField[];
  required_fields?: BackendField[];
  optional_fields?: BackendField[];
}

export interface BackendSchemas {
  azurerm?: BackendSchema;
  s3?: BackendSchema;
  http?: BackendSchema;
  pg?: BackendSchema;
  kubernetes?: BackendSchema;
  gcs?: BackendSchema;
}