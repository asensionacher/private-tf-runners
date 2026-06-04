import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { auth } from '../lib/api';
import type { User, LoginRequest, UserRole } from '../types';
import { hasPermission, type Permission } from '../lib/permissions';

interface AuthContextType {
  user: User | null;
  role: UserRole | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  can: (permission: Permission) => boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const checkAuth = useCallback(async () => {
    try {
      const token = localStorage.getItem('token');
      if (token) {
        const userData = await auth.me();
        setUser(userData);
      }
    } catch {
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  const login = async (credentials: LoginRequest) => {
    await auth.login(credentials);
    await checkAuth();
  };

  const logout = async () => {
    await auth.logout();
    setUser(null);
  };

  const role = (user?.role as UserRole) ?? null;
  const can = useCallback(
    (permission: Permission) => role ? hasPermission(role, permission) : false,
    [role]
  );

  return (
    <AuthContext.Provider
      value={{
        user,
        role,
        isLoading,
        isAuthenticated: !!user,
        can,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
