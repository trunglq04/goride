"use client";

import React, { createContext, useContext, useEffect, useState, useCallback } from "react";
import { API_URL } from "../constants";
import { BackendEndpoints } from "../contracts";
import { AuthResponse, AuthUser } from "../types";

interface AuthContextType {
  user: AuthUser | null;
  accessToken: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  register: (data: {
    fullName: string;
    email: string;
    phone: string;
    password: string;
    role: "rider" | "driver";
  }) => Promise<{ success: boolean; userId?: string; error?: string }>;
  verifyOTP: (
    userId: string,
    code: string
  ) => Promise<{ success: boolean; user?: AuthUser; error?: string }>;
  resendOTP: (userId: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const STORAGE_KEYS = {
  ACCESS_TOKEN: "goride_access_token",
  REFRESH_TOKEN: "goride_refresh_token",
  USER: "goride_user",
};

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Load stored credentials on mount
  useEffect(() => {
    try {
      const storedToken = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN);
      const storedRefresh = localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);
      const storedUser = localStorage.getItem(STORAGE_KEYS.USER);

      if (storedToken && storedUser) {
        setAccessToken(storedToken);
        setRefreshToken(storedRefresh);
        setUser(JSON.parse(storedUser));
      }
    } catch (e) {
      console.error("Failed to restore auth session from localStorage", e);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const saveSession = useCallback((tokens: { accessToken: string; refreshToken: string; user: AuthUser }) => {
    setAccessToken(tokens.accessToken);
    setRefreshToken(tokens.refreshToken);
    setUser(tokens.user);

    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, tokens.accessToken);
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, tokens.refreshToken);
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(tokens.user));
  }, []);

  const clearSession = useCallback(() => {
    setAccessToken(null);
    setRefreshToken(null);
    setUser(null);

    localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.USER);
  }, []);

  const login = async (email: string, password: string) => {
    try {
      const res = await fetch(`${API_URL}${BackendEndpoints.AUTH_LOGIN}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      const data = await res.json();
      if (!res.ok) {
        return { success: false, error: data.error || "Login failed" };
      }

      const authData = data.data as AuthResponse;
      saveSession({
        accessToken: authData.accessToken,
        refreshToken: authData.refreshToken,
        user: authData.user,
      });

      return { success: true };
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : "Network error. Please try again.";
      return { success: false, error: errorMsg };
    }
  };

  const register = async (formData: {
    fullName: string;
    email: string;
    phone: string;
    password: string;
    role: "rider" | "driver";
  }) => {
    try {
      const res = await fetch(`${API_URL}${BackendEndpoints.AUTH_REGISTER}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(formData),
      });

      const data = await res.json();
      if (!res.ok) {
        return { success: false, error: data.error || "Registration failed" };
      }

      const userId = data.data?.user_id || data.data?.userId;
      return { success: true, userId };
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : "Network error. Please try again.";
      return { success: false, error: errorMsg };
    }
  };

  const verifyOTP = async (userId: string, code: string) => {
    try {
      const res = await fetch(`${API_URL}${BackendEndpoints.AUTH_VERIFY_OTP}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ userID: userId, code }),
      });

      const data = await res.json();
      if (!res.ok) {
        return { success: false, error: data.error || "Verification failed" };
      }

      const authData = data.data as AuthResponse;
      saveSession({
        accessToken: authData.accessToken,
        refreshToken: authData.refreshToken,
        user: authData.user,
      });

      return { success: true, user: authData.user };
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : "Network error. Please try again.";
      return { success: false, error: errorMsg };
    }
  };

  const resendOTP = async (userId: string) => {
    try {
      const res = await fetch(`${API_URL}${BackendEndpoints.AUTH_RESEND_OTP}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ userID: userId }),
      });

      const data = await res.json();
      if (!res.ok) {
        return { success: false, error: data.error || "Failed to resend code" };
      }

      return { success: true };
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : "Network error. Please try again.";
      return { success: false, error: errorMsg };
    }
  };

  const logout = async () => {
    try {
      if (refreshToken) {
        await fetch(`${API_URL}${BackendEndpoints.AUTH_LOGOUT}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refreshToken }),
        });
      }
    } catch (e) {
      console.warn("Logout request failed on server", e);
    } finally {
      clearSession();
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        accessToken,
        isLoading,
        login,
        register,
        verifyOTP,
        resendOTP,
        logout,
        isAuthenticated: !!user && !!accessToken,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
