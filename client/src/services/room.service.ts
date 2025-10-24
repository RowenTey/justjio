import { apiClient, type ApiRequest, type ApiResponse } from "../api/client";

export const roomService = {
  getRooms: async (page?: number) => {
    const { data } = await apiClient.get<ApiResponse<"/rooms", "get">>(
      "/rooms",
      { params: { page } },
    );
    return data;
  },

  createRoom: async (roomData: ApiRequest<"/rooms", "post">) => {
    const { data } = await apiClient.post<ApiResponse<"/rooms", "post">>(
      "/rooms",
      roomData,
    );
    return data;
  },

  getRoomById: async (roomId: string) => {
    const { data } = await apiClient.get<ApiResponse<"/rooms/{roomId}", "get">>(
      `/rooms/${roomId}`,
    );
    return data;
  },

  editRoom: async (
    roomId: string,
    roomData: ApiRequest<"/rooms/{roomId}/edit", "patch">,
  ) => {
    const { data } = await apiClient.patch<
      ApiResponse<"/rooms/{roomId}/edit", "patch">
    >(`/rooms/${roomId}/edit`, roomData);
    return data;
  },

  joinRoom: async (roomId: string) => {
    const { data } = await apiClient.patch<
      ApiResponse<"/rooms/{roomId}/join", "patch">
    >(`/rooms/${roomId}/join`);
    return data;
  },

  leaveRoom: async (roomId: string) => {
    const { data } = await apiClient.delete<
      ApiResponse<"/rooms/{roomId}/leave", "delete">
    >(`/rooms/${roomId}/leave`);
    return data;
  },

  closeRoom: async (roomId: string) => {
    const { data } = await apiClient.patch<
      ApiResponse<"/rooms/{roomId}/close", "patch">
    >(`/rooms/${roomId}/close`);
    return data;
  },

  inviteUsers: async (
    roomId: string,
    inviteData: ApiRequest<"/rooms/{roomId}", "post">,
  ) => {
    const { data } = await apiClient.post<
      ApiResponse<"/rooms/{roomId}", "post">
    >(`/rooms/${roomId}`, inviteData);
    return data;
  },

  respondToInvite: async (
    roomId: string,
    responseData: ApiRequest<"/rooms/{roomId}", "patch">,
  ) => {
    const { data } = await apiClient.patch<
      ApiResponse<"/rooms/{roomId}", "patch">
    >(`/rooms/${roomId}`, responseData);
    return data;
  },

  getUninvitedUsers: async (roomId: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/rooms/{roomId}/uninvited", "get">
    >(`/rooms/${roomId}/uninvited`);
    return data;
  },

  getRoomsCount: async () => {
    const { data } =
      await apiClient.get<ApiResponse<"/rooms/count", "get">>("/rooms/count");
    return data;
  },

  getRoomInvites: async () => {
    const { data } =
      await apiClient.get<ApiResponse<"/rooms/invites", "get">>(
        "/rooms/invites",
      );
    return data;
  },

  getRoomInvitesCount: async () => {
    const { data } = await apiClient.get<
      ApiResponse<"/rooms/invites/count", "get">
    >("/rooms/invites/count");
    return data;
  },

  getPublicRooms: async () => {
    const { data } =
      await apiClient.get<ApiResponse<"/rooms/public", "get">>("/rooms/public");
    return data;
  },

  searchVenues: async (query?: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/rooms/venues/search", "get">
    >("/rooms/venues/search", { params: { query } });
    return data;
  },
};
