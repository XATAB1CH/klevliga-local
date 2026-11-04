import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useLogin, useLogout, useMe, useRegister } from '../api/hooks';
import type { User } from '../api/types';

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<unknown>;
  register: (email: string, username: string, password: string) => Promise<unknown>;
  logout: () => Promise<unknown>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const me = useMe();
  const loginMut = useLogin();
  const registerMut = useRegister();
  const logoutMut = useLogout();

  const value = useMemo<AuthContextValue>(() => ({
    user: me.data ?? null,
    isLoading: me.isLoading,
    isAuthenticated: !!me.data,
    login: (email, password) => loginMut.mutateAsync({ email, password }),
    register: (email, username, password) => registerMut.mutateAsync({ email, username, password }),
    logout: () => logoutMut.mutateAsync(),
  }), [me.data, me.isLoading, loginMut, registerMut, logoutMut]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
