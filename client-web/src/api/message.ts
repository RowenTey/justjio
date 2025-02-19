import { AxiosInstance, AxiosResponse } from "axios";
import { ApiResponse } from ".";
import { IUser } from "../types/user";

interface FetchRoomMessageResponse extends ApiResponse {
	data: {
		messages: {
			id: number;
			room_id: number;
			sender_id: number;
			sender: IUser;
			content: string;
			sent_at: string;
		}[];
		page: number;
		page_count: number;
	};
}

export const sendMessageApi = (
	api: AxiosInstance,
	roomId: string,
	message: string,
	mock: boolean = false
): Promise<AxiosResponse<ApiResponse>> => {
	if (!mock) {
		return api.post<ApiResponse>(`rooms/${roomId}/messages`, {
			content: message,
		});
	}

	return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
		setTimeout(() => {
			resolve({
				data: {
					message: "Message saved successfully",
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

export const fetchRoomMessageApi = (
	api: AxiosInstance,
	roomId: string,
	page: number,
	mock: boolean = false
): Promise<AxiosResponse<FetchRoomMessageResponse>> => {
	if (!mock) {
		return api.get<FetchRoomMessageResponse>(`rooms/${roomId}/messages`, {
			params: {
				page,
				asc: false,
			},
		});
	}

	return new Promise<AxiosResponse<FetchRoomMessageResponse>>((resolve) => {
		setTimeout(() => {
			resolve({
				data: {
					data: {
						messages: [
							{
								id: 1,
								room_id: 1,
								sender_id: 1,
								sender: {
									id: 1,
									name: "User1",
									email: "user1@example.com",
									password: "password123",
									phone_num: "1234567890",
									is_email_valid: true,
									is_online: true,
									last_seen: new Date().toISOString(),
									registered_at: new Date().toISOString(),
								},
								content: "Hello",
								sent_at: new Date().toISOString(),
							},
							{
								id: 2,
								room_id: 1,
								sender_id: 2,
								sender: {
									id: 2,
									username: "User2",
									pictureUrl: "https://example.com/picture.jpg",
									email: "user2@example.com",
									password: "password456",
									isEmailValid: true,
									isOnline: true,
									lastSeen: new Date().toISOString(),
									registeredAt: new Date().toISOString(),
								} as IUser,
								content: "Hi",
								sent_at: new Date().toISOString(),
							},
						],
						page: 1,
						page_count: 1,
					},
					message: "Fetched messages successfully",
					status: "success",
				},
				status: 200,
				statusText: "OK",
				headers: {},
				config: {},
			} as AxiosResponse<FetchRoomMessageResponse>);
		}, 1500);
	});
};
