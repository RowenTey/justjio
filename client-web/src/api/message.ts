import { AxiosInstance, AxiosResponse } from "axios";
import { ApiResponse } from ".";
import { IUser } from "../types/user";

interface FetchRoomMessageResponse extends ApiResponse {
  data: {
    messages: {
      id: number;
      roomId: number;
      senderId: number;
      sender: IUser;
      content: string;
      sentAt: string;
    }[];
    page: number;
    pageCount: number;
  };
}

export const sendMessageApi = (
  api: AxiosInstance,
  roomId: string,
  message: string,
  mock: boolean = false,
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
  mock: boolean = false,
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
                roomId: 1,
                senderId: 1,
                sender: {
                  id: 1,
                  name: "User1",
                  email: "user1@example.com",
                  password: "password123",
                  isEmailValid: true,
                  isOnline: true,
                  lastseen: new Date().toISOString(),
                  registeredAt: new Date().toISOString(),
                },
                content: "Hello",
                sentAt: new Date().toISOString(),
              },
              {
                id: 2,
                roomId: 1,
                senderId: 2,
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
                sentAt: new Date().toISOString(),
              },
            ],
            page: 1,
            pageCount: 1,
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
