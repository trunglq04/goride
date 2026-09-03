import { apiClient } from "./apiClient";
import { BackendEndpoints } from "@/types/contracts";
import { AuthResponse, AuthUser } from "@/types";

export interface RegisterRequest {
  fullName: string;
  email: string;
  phone: string;
  password: string;
  role: "rider" | "driver";
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface VerifyOTPRequest {
  userID: string;
  code: string;
}

export interface ResendOTPRequest {
  userID: string;
}

export const authService = {
  async register(data: RegisterRequest): Promise<{ userId: string }> {
    return apiClient<{ userId: string }>(BackendEndpoints.AUTH_REGISTER, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async login(data: LoginRequest): Promise<AuthResponse> {
    return apiClient<AuthResponse>(BackendEndpoints.AUTH_LOGIN, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async verifyOTP(data: VerifyOTPRequest): Promise<AuthResponse> {
    return apiClient<AuthResponse>(BackendEndpoints.AUTH_VERIFY_OTP, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async resendOTP(data: ResendOTPRequest): Promise<{ message: string }> {
    return apiClient<{ message: string }>(BackendEndpoints.AUTH_RESEND_OTP, {
      method: "POST",
      body: JSON.stringify(data),
    });
  },

  async refreshToken(refreshToken: string): Promise<AuthResponse> {
    return apiClient<AuthResponse>(BackendEndpoints.AUTH_REFRESH, {
      method: "POST",
      body: JSON.stringify({ refreshToken }),
    });
  },

  async logout(refreshToken: string): Promise<void> {
    return apiClient<void>(BackendEndpoints.AUTH_LOGOUT, {
      method: "POST",
      body: JSON.stringify({ refreshToken }),
    });
  },

  async getMe(token?: string | null): Promise<AuthUser> {
    return apiClient<AuthUser>(BackendEndpoints.AUTH_ME, {
      method: "GET",
      token,
    });
  },
};
