import axios from "axios";

export const api = axios.create({
	baseURL: import.meta.env.VITE_API_URL,
});

api.interceptors.request.use((req) => {
	const token = localStorage.getItem("accessToken");
	if (token !== null) {
		req.headers.Authorization = `Bearer ${token}`;
	}
	return req;
});

export interface ApiResponse {
	data: object;
	message: string;
	status: string;
}
