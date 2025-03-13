import axios from "axios";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
});

api.interceptors.request.use((req) => {
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

export interface ApiResponse {
  data: object;
  message: string;
  status: string;
}
