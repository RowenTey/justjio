import { AxiosInstance, AxiosResponse } from "axios";
import { ApiResponse } from ".";
import { IUser } from "../types/user";

interface GetNumFriendsResponse extends ApiResponse {
	data: {
		numFriends: number;
	};
}

interface FetchFriendsResponse extends ApiResponse {
	data: IUser[];
}

export const getNumFriendsApi = (
	api: AxiosInstance,
	userId: number,
	mock: boolean = false
): Promise<AxiosResponse<GetNumFriendsResponse>> => {
	if (!mock) {
		return api.get<GetNumFriendsResponse>(`/users/${userId}/friends/count`);
	}

	return new Promise<AxiosResponse<GetNumFriendsResponse>>((resolve) => {
		setTimeout(() => {
			resolve({
				data: {
					data: {
						numFriends: 5,
					},
					message: "Number of friends retrieved successfully",
					status: "success",
				},
				status: 200,
				statusText: "OK",
				headers: {},
				config: {},
			} as AxiosResponse<GetNumFriendsResponse>);
		}, 1500);
	});
};

export const fetchFriendsApi = (
	api: AxiosInstance,
	userId: number,
	mock: boolean = false
): Promise<AxiosResponse<FetchFriendsResponse>> => {
	if (!mock) {
		return api.get<FetchFriendsResponse>(`/users/${userId}/friends`);
	}

	return new Promise<AxiosResponse<FetchFriendsResponse>>((resolve) => {
		setTimeout(() => {
			resolve({
				data: {
					data: [
						{
							id: 1,
							username: "testuser1",
							email: "test@test.com",
						},
					],
					message: "Friends retrieved successfully",
					status: "success",
				},
				status: 200,
				statusText: "OK",
				headers: {},
				config: {},
			} as AxiosResponse<FetchFriendsResponse>);
		}, 1500);
	});
};

export const searchFriendsApi = (
	api: AxiosInstance,
	userId: number,
	query: string,
	mock: boolean = false
): Promise<AxiosResponse<FetchFriendsResponse>> => {
	if (!mock) {
		return api.get<FetchFriendsResponse>(`/users/${userId}/friends/search`, {
			params: {
				query,
			},
		});
	}

	return new Promise<AxiosResponse<FetchFriendsResponse>>((resolve) => {
		setTimeout(() => {
			resolve({
				data: {
					data: [
						{
							id: 1,
							username: "testuser1",
							email: "test@test.com",
						},
					],
					message: "Friends retrieved successfully",
					status: "success",
				},
				status: 200,
				statusText: "OK",
				headers: {},
				config: {},
			} as AxiosResponse<FetchFriendsResponse>);
		}, 1500);
	});
};

export const addFriendApi = (
	api: AxiosInstance,
	userId: number,
	friendId: number,
	mock: boolean = false
): Promise<AxiosResponse<ApiResponse>> => {
	if (!mock) {
		return api.post<ApiResponse>(`/users/${userId}/friends`, {
			friendId: friendId.toString(),
		});
	}

	return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
		setTimeout(() => {
			resolve({
				data: {
					message: "Friend added successfully",
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
