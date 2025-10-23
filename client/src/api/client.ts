import axios, { type AxiosInstance } from "axios";
import type { paths } from "../types/api";

// Create a typed axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/v1",
  headers: {
    "Content-Type": "application/json",
  },
});

// Add auth token to requests
apiClient.interceptors.request.use((req) => {
  const token = localStorage.getItem("accessToken");
  if (token !== null) {
    req.headers.Authorization = `Bearer ${token}`;
  }

  // Add CF headers for staging & production
  const env = import.meta.env.VITE_ENV;
  if (env !== "dev") {
    req.headers["CF-Access-Client-Id"] =
      import.meta.env.VITE_CF_ACCESS_CLIENT_ID;
    req.headers["CF-Access-Client-Secret"] =
      import.meta.env.VITE_CF_ACCESS_CLIENT_SECRET;
  }

  return req;
});

// Type helpers
type ApiPath = keyof paths;

type ApiMethod<Path extends ApiPath> = keyof paths[Path];

type ApiQueryParams<
  Path extends ApiPath,
  Method extends ApiMethod<Path>,
> = paths[Path][Method] extends {
  parameters: { query?: infer Q };
}
  ? Q
  : never;

type ApiRequest<
  Path extends ApiPath,
  Method extends ApiMethod<Path>,
> = paths[Path][Method] extends {
  requestBody: { content: { "application/json": infer Body } };
}
  ? Body
  : never;

type ApiResponse<
  Path extends ApiPath,
  Method extends ApiMethod<Path>,
> = paths[Path][Method] extends {
  responses: { 200: { content: { "application/json": infer Res } } };
}
  ? Res
  : never;

export { apiClient };
export type { ApiPath, ApiMethod, ApiQueryParams, ApiRequest, ApiResponse };
