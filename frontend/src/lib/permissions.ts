import type { UserRole } from '../types';

export const Permissions = {
  STACK_READ: 'stack:read',
  STACK_CREATE: 'stack:create',
  STACK_DELETE: 'stack:delete',
  RUN_READ: 'run:read',
  RUN_CREATE: 'run:create',
  RUN_DELETE: 'run:delete',
  RUNNER_READ: 'runner:read',
  RUNNER_CREATE: 'runner:create',
  RUNNER_DELETE: 'runner:delete',
  RUNNER_TOKEN: 'runner:token',
  USER_CREATE: 'user:create',
  USER_ADMIN: 'user:admin',
} as const;

export type Permission = typeof Permissions[keyof typeof Permissions];

const RolePermissions: Record<UserRole, Permission[]> = {
  viewer: [
    Permissions.STACK_READ,
    Permissions.RUN_READ,
    Permissions.RUNNER_READ,
  ],
  operator: [
    Permissions.STACK_READ,
    Permissions.STACK_CREATE,
    Permissions.STACK_DELETE,
    Permissions.RUN_READ,
    Permissions.RUN_CREATE,
    Permissions.RUN_DELETE,
    Permissions.RUNNER_READ,
    Permissions.RUNNER_CREATE,
    Permissions.RUNNER_DELETE,
    Permissions.RUNNER_TOKEN,
  ],
  admin: [
    Permissions.STACK_READ,
    Permissions.STACK_CREATE,
    Permissions.STACK_DELETE,
    Permissions.RUN_READ,
    Permissions.RUN_CREATE,
    Permissions.RUN_DELETE,
    Permissions.RUNNER_READ,
    Permissions.RUNNER_CREATE,
    Permissions.RUNNER_DELETE,
    Permissions.RUNNER_TOKEN,
    Permissions.USER_CREATE,
    Permissions.USER_ADMIN,
  ],
};

export function hasPermission(role: UserRole, permission: Permission): boolean {
  return RolePermissions[role]?.includes(permission) ?? false;
}

export function hasAnyPermission(role: UserRole, permissions: Permission[]): boolean {
  return permissions.some((p) => hasPermission(role, p));
}

export function hasAllPermissions(role: UserRole, permissions: Permission[]): boolean {
  return permissions.every((p) => hasPermission(role, p));
}

export const PermissionsByRole = {
  viewer: ['View stacks, runs, runners'],
  operator: ['All viewer permissions + create/delete stacks and runs, manage runners'],
  admin: ['All operator permissions + user management'],
};