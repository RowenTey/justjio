import { AxiosInstance, AxiosResponse } from "axios";
import { ApiResponse } from ".";

export interface LoginResponse extends ApiResponse {
  data: {
    id: number;
    username: string;
    email: string;
    pictureUrl: string;
  };
  token: string;
}

export const loginApi = (
  api: AxiosInstance,
  username: string,
  password: string,
  mock: boolean = false,
): Promise<AxiosResponse<LoginResponse>> => {
  if (!mock) {
    return api.post<LoginResponse>("/auth", { username, password });
  }

  return new Promise<AxiosResponse<LoginResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            id: 6,
            username: username,
            email: "test123@test.com",
          },
          message: "Login successfully",
          status: "success",
          token: "token123",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<LoginResponse>);
    }, 1500);
  });
};

export const googleLoginApi = (
  api: AxiosInstance,
  code: string,
  mock: boolean = false,
): Promise<AxiosResponse<LoginResponse>> => {
  if (!mock) {
    return api.post<LoginResponse>("/auth/google", { code });
  }

  return new Promise<AxiosResponse<LoginResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            id: 6,
            username: "test",
            email: "test123@test.com",
          },
          message: "Login successfully",
          status: "success",
          token: "token123",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<LoginResponse>);
    }, 1500);
  });
};

export const signUpApi = (
  api: AxiosInstance,
  username: string,
  // phoneNum: string,
  email: string,
  password: string,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.post<ApiResponse>("/auth/signup", {
      username,
      email,
      password,
    });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            id: 6,
            username: "test",
            email: "test@test.com",
          },
          message: "Sign up successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<ApiResponse>);
    }, 1500);
  });
};

export const verifyOtpApi = (
  api: AxiosInstance,
  email: string,
  otp: string,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.post<ApiResponse>("/auth/verify", { email, otp });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {},
          message: "OTP verified successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<ApiResponse>);
    }, 1500);
  });
};

export const sendOtpEmailApi = (
  api: AxiosInstance,
  email: string,
  purpose: "reset-password" | "verify-email",
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.post<ApiResponse>("/auth/otp", { email, purpose });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {},
          message: "OTP sent successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<ApiResponse>);
    }, 1500);
  });
};

export const resetPasswordApi = (
  api: AxiosInstance,
  email: string,
  password: string,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.patch<ApiResponse>("/auth/reset", { email, password });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {},
          message: "Password reset successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<ApiResponse>);
    }, 1500);
  });
};
