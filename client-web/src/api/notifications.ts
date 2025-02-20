import { AxiosInstance, AxiosResponse } from "axios";
import { ApiResponse } from ".";
import { INotification } from "../types/notifications";

export interface NotificationResponse extends ApiResponse {
  data: INotification[];
}

export const createNotificationApi = (
  api: AxiosInstance,
  userId: number,
  message: string,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.post<ApiResponse>("/notifications", {
      userId,
      content: message,
    });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            id: 1,
            userId: 1,
            content: message,
            isRead: false,
            createdAt: new Date().toISOString(),
          },
          message: "Notification created successfully",
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

export const getNotificationsApi = (
  api: AxiosInstance,
  mock: boolean = false,
): Promise<AxiosResponse<NotificationResponse>> => {
  if (!mock) {
    return api.get<NotificationResponse>(`/notifications`);
  }

  return new Promise<AxiosResponse<NotificationResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              id: 1,
              userId: 1,
              content: "Test notification 1",
              isRead: false,
              createdAt: new Date().toISOString(),
            },
            {
              id: 2,
              userId: 1,
              content: "Test notification 2",
              isRead: true,
              createdAt: new Date().toISOString(),
            },
          ],
          message: "Notifications retrieved successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<NotificationResponse>);
    }, 1500);
  });
};

export const markNotificationAsReadApi = (
  api: AxiosInstance,
  userId: number,
  notificationId: number,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.patch<ApiResponse>(
      `/users/${userId}/notifications/${notificationId}`,
    );
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {},
          message: "Notification marked as read successfully",
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
