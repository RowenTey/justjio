import { AxiosInstance, AxiosResponse } from "axios";
import { ApiResponse } from ".";
import { IFriendRequests, IUser } from "../types/user";

interface UpdateUserRequest {
  field: string;
  value: string;
}

interface GetNumFriendsResponse extends ApiResponse {
  data: {
    numFriends: number;
  };
}

interface FetchFriendsResponse extends ApiResponse {
  data: IUser[];
}

interface FetchFriendRequestsResponse extends ApiResponse {
  data: IFriendRequests[];
}

interface CountPendingFriendRequestsResponse extends ApiResponse {
  data: {
    count: number;
  };
}

interface UpdateUserResponse extends ApiResponse {
  data: UpdateUserRequest;
}

export const getNumFriendsApi = (
  api: AxiosInstance,
  userId: number,
  mock: boolean = false,
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

export const updateUserApi = (
  api: AxiosInstance,
  userId: number,
  field: string,
  value: string,
  mock: boolean = false,
): Promise<AxiosResponse<UpdateUserResponse>> => {
  if (!mock) {
    return api.patch<UpdateUserResponse>(`/users/${userId}`, {
      field,
      value,
    });
  }

  return new Promise<AxiosResponse<UpdateUserResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            field,
            value,
          },
          message: "User successfully updated",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<UpdateUserResponse>);
    }, 1500);
  });
};

export const fetchFriendsApi = (
  api: AxiosInstance,
  userId: number,
  mock: boolean = false,
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

export const fetchFriendRequestsApi = (
  api: AxiosInstance,
  userId: number,
  status: "pending" | "accepted" | "rejected",
  mock: boolean = false,
): Promise<AxiosResponse<FetchFriendRequestsResponse>> => {
  if (!mock) {
    return api.get<FetchFriendRequestsResponse>(
      `/users/${userId}/friendRequests?status=${status}`,
    );
  }

  return new Promise<AxiosResponse<FetchFriendRequestsResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              id: 1,
              senderId: 2,
              receiverId: 1,
              status: "pending",
              sender: {
                id: 2,
                username: "testuser2",
                email: "test@test.com",
              },
              receiver: {
                id: 1,
                username: "testuser1",
                email: "test@test.com",
              },
              sentAt: new Date().toISOString(),
              respondedAt: null,
            },
          ],
          message: "Friend requests retrieved successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchFriendRequestsResponse>);
    }, 1500);
  });
};

export const searchFriendsApi = (
  api: AxiosInstance,
  userId: number,
  query: string,
  mock: boolean = false,
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

export const sendFriendRequestApi = (
  api: AxiosInstance,
  userId: number,
  friendId: number,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.post<ApiResponse>(`/users/${userId}/friendRequests`, {
      friendId,
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

export const removeFriendApi = (
  api: AxiosInstance,
  userId: number,
  friendId: number,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.delete<ApiResponse>(`/users/${userId}/friends/${friendId}`);
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          message: "Friend removed successfully",
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

export const respondToFriendRequestApi = (
  api: AxiosInstance,
  userId: number,
  friendRequestId: number,
  action: "accept" | "reject",
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.patch<ApiResponse>(`/users/${userId}/friendRequests`, {
      requestId: friendRequestId,
      action,
    });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          message: "Friend request responded to successfully",
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

export const countPendingFriendRequestsApi = (
  api: AxiosInstance,
  userId: number,
  mock: boolean = false,
): Promise<AxiosResponse<CountPendingFriendRequestsResponse>> => {
  if (!mock) {
    return api.get<CountPendingFriendRequestsResponse>(
      `/users/${userId}/friendRequests/count`,
    );
  }

  return new Promise<AxiosResponse<CountPendingFriendRequestsResponse>>(
    (resolve) => {
      setTimeout(() => {
        resolve({
          data: {
            data: {
              count: 1,
            },
            message: "Number of pending friend requests retrieved successfully",
            status: "success",
          },
          status: 200,
          statusText: "OK",
          headers: {},
          config: {},
        } as AxiosResponse<CountPendingFriendRequestsResponse>);
      }, 1500);
    },
  );
};
