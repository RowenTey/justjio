import { apiClient, type ApiRequest, type ApiResponse } from "../api/client";

export const notificationService = {
  createNotification: async (
    notificationData: ApiRequest<"/notifications", "post">,
  ) => {
    const { data } = await apiClient.post<
      ApiResponse<"/notifications", "post">
    >("/notifications", notificationData);
    return data;
  },

  getNotificationById: async (id: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/notifications/{id}", "get">
    >(`/notifications/${id}`);
    return data;
  },
};
