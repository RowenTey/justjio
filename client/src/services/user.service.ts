import {
  apiClient,
  ApiQueryParams,
  type ApiRequest,
  type ApiResponse,
} from "../api/client";

export const userService = {
  getUserById: async (userId: string) => {
    const { data } = await apiClient.get<ApiResponse<"/users/{userId}", "get">>(
      `/users/${userId}`,
    );
    return data;
  },

  updateUsername: async (
    userId: string,
    usernameData: ApiRequest<"/users/{userId}/username", "patch">,
  ) => {
    const { data } = await apiClient.patch<
      ApiResponse<"/users/{userId}/username", "patch">
    >(`/users/${userId}/username`, usernameData);
    return data;
  },

  // Friend requests
  getFriendRequests: async (
    userId: string,
    params?: ApiQueryParams<"/users/{userId}/friendRequests", "get">,
  ) => {
    const { data } = await apiClient.get<
      ApiResponse<"/users/{userId}/friendRequests", "get">
    >(`/users/${userId}/friendRequests`, { params });
    return data;
  },

  sendFriendRequest: async (
    userId: string,
    requestData: ApiRequest<"/users/{userId}/friendRequests", "post">,
  ) => {
    const { data } = await apiClient.post<
      ApiResponse<"/users/{userId}/friendRequests", "post">
    >(`/users/${userId}/friendRequests`, requestData);
    return data;
  },

  respondToFriendRequest: async (
    userId: string,
    responseData: ApiRequest<"/users/{userId}/friendRequests", "patch">,
  ) => {
    const { data } = await apiClient.patch<
      ApiResponse<"/users/{userId}/friendRequests", "patch">
    >(`/users/${userId}/friendRequests`, responseData);
    return data;
  },

  getFriendRequestsCount: async (userId: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/users/{userId}/friendRequests/count", "get">
    >(`/users/${userId}/friendRequests/count`);
    return data;
  },

  // Friends
  getFriends: async (userId: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/users/{userId}/friends", "get">
    >(`/users/${userId}/friends`);
    return data;
  },

  removeFriend: async (userId: string, friendId: string) => {
    const { data } = await apiClient.delete<
      ApiResponse<"/users/{userId}/friends/{friendId}", "delete">
    >(`/users/${userId}/friends/${friendId}`);
    return data;
  },

  getFriendsCount: async (userId: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/users/{userId}/friends/count", "get">
    >(`/users/${userId}/friends/count`);
    return data;
  },

  searchFriends: async (userId: string, query?: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/users/{userId}/friends/search", "get">
    >(`/users/${userId}/friends/search`, { params: { query } });
    return data;
  },

  // Notifications
  getUserNotifications: async (userId: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/users/{userId}/notifications", "get">
    >(`/users/${userId}/notifications`);
    return data;
  },

  updateNotification: async (userId: string, notificationId: string) => {
    const { data } = await apiClient.patch<
      ApiResponse<"/users/{userId}/notifications/{notificationId}", "patch">
    >(`/users/${userId}/notifications/${notificationId}`);
    return data;
  },
};
