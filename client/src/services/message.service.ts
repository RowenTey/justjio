import { apiClient, type ApiRequest, type ApiResponse } from "../api/client";

export const messageService = {
  getRoomMessages: async (roomId: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/rooms/{roomId}/messages", "get">
    >(`/rooms/${roomId}/messages`);
    return data;
  },

  createMessage: async (
    roomId: string,
    messageData: ApiRequest<"/rooms/{roomId}/messages", "post">,
  ) => {
    const { data } = await apiClient.post<
      ApiResponse<"/rooms/{roomId}/messages", "post">
    >(`/rooms/${roomId}/messages`, messageData);
    return data;
  },

  getMessageById: async (roomId: string, msgId: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/rooms/{roomId}/messages/{msgId}", "get">
    >(`/rooms/${roomId}/messages/${msgId}`);
    return data;
  },
};
