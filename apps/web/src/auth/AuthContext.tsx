import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useLogin, useLogout, useMe, useRegister, useVerifyRegistration } from '../api/hooks';
import type { SessionUser } from '../api/types';

interface AuthContextValue {
  user: SessionUser | null;
  isLoading: boolean;
  error: Error | null;
  isLoggingOut: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<unknown>;
  register: (email: string, username: string, password: string) => Promise<unknown>;
  verifyRegistration: (email: string, code: string) => Promise<unknown>;
  logout: () => Promise<unknown>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const me = useMe();
  const loginMut = useLogin();
  const registerMut = useRegister();
  const verifyMut = useVerifyRegistration();
  const logoutMut = useLogout();

  const value = useMemo<AuthContextValue>(() => ({
    user: me.data ?? null,
    isLoading: me.isLoading,
    error: me.error,
    isLoggingOut: logoutMut.isPending,
    isAuthenticated: !!me.data,
    login: (email, password) => loginMut.mutateAsync({ email, password }),
    register: (email, username, password) => registerMut.mutateAsync({ email, username, password }),
    verifyRegistration: (email, code) => verifyMut.mutateAsync({ email, code }),
    logout: () => logoutMut.mutateAsync(),
  }), [me.data, me.isLoading, me.error, loginMut, registerMut, verifyMut, logoutMut]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
