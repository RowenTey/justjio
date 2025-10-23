import { apiClient, type ApiRequest, type ApiResponse } from "../api/client";

export const authService = {
  login: async (credentials: ApiRequest<"/auth", "post">) => {
    const { data } = await apiClient.post<ApiResponse<"/auth", "post">>(
      "/auth",
      credentials,
    );
    return data;
  },

  signup: async (signupData: ApiRequest<"/auth/signup", "post">) => {
    const { data } = await apiClient.post<ApiResponse<"/auth/signup", "post">>(
      "/auth/signup",
      signupData,
    );
    return data;
  },

  googleAuth: async (googleData: ApiRequest<"/auth/google", "post">) => {
    const { data } = await apiClient.post<ApiResponse<"/auth/google", "post">>(
      "/auth/google",
      googleData,
    );
    return data;
  },

  sendOTP: async (otpData: ApiRequest<"/auth/otp", "post">) => {
    const { data } = await apiClient.post<ApiResponse<"/auth/otp", "post">>(
      "/auth/otp",
      otpData,
    );
    return data;
  },

  verifyOTP: async (otpData: ApiRequest<"/auth/verify", "post">) => {
    const { data } = await apiClient.post<ApiResponse<"/auth/verify", "post">>(
      "/auth/verify",
      otpData,
    );
    return data;
  },

  resetPassword: async (resetData: ApiRequest<"/auth/reset", "post">) => {
    const { data } = await apiClient.post<ApiResponse<"/auth/reset", "post">>(
      "/auth/reset",
      resetData,
    );
    return data;
  },
};
