import { AxiosInstance, AxiosResponse } from "axios";
import { IRoom, IRoomInvite, IVenue } from "../types/room";
import { ApiResponse } from ".";
import { IUser } from "../types/user";

interface FetchRoomsResponse extends ApiResponse {
  data: IRoom[];
}

interface GetNumRoomsResponse extends ApiResponse {
  data: {
    count: number;
  };
}

interface CreateRoomResponse extends ApiResponse {
  data: {
    room: IRoom;
    invites: [];
  };
}

interface RespondToInviteResponse extends ApiResponse {
  data: {
    room?: IRoom;
    attendees?: [];
  };
}

interface FetchRoomInvitesResponse extends ApiResponse {
  data: IRoomInvite[];
}

interface FetchNumRoomInvitesResponse extends ApiResponse {
  data: {
    count: number;
  };
}

interface FetchRoomResponse extends ApiResponse {
  data: IRoom;
}

interface FetchRoomAttendeesResponse extends ApiResponse {
  data: IUser[];
}

interface QueryVenueResponse extends ApiResponse {
  data: IVenue[];
}

export type UpdateRoomRequest = {
  venue: string;
  placeId: string;
  date: string;
  time: string;
  description: string;
};

export const fetchRecentRoomsApi = (
  api: AxiosInstance,
  mock: boolean = false,
): Promise<AxiosResponse<FetchRoomsResponse>> => {
  if (!mock) {
    return api.get<FetchRoomsResponse>("/rooms");
  }

  return new Promise<AxiosResponse<FetchRoomsResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              id: "1",
              name: "Test Room",
              time: "5:00pm",
              venue: "ntu hall 9",
              date: "2022-09-04T00:00:00Z",
              hostId: 6,
              host: {},
              createdAt: "2021-09-25T02:00:00Z",
              updatedAt: "2021-09-25T02:00:00Z",
              attendeesCount: 1,
              url: "",
              isClosed: false,
            },
          ],
          message: "Fetched recent rooms successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchRoomsResponse>);
    }, 1500);
  });
};

export const fetchExploreMoreRoomsApi = (
  api: AxiosInstance,
  mock: boolean = false,
): Promise<AxiosResponse<FetchRoomsResponse>> => {
  if (!mock) {
    return api.get<FetchRoomsResponse>("/rooms/public");
  }

  return new Promise<AxiosResponse<FetchRoomsResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              id: "1",
              name: "Test Room",
              time: "5:00pm",
              venue: "ntu hall 9",
              date: "2022-09-04T00:00:00Z",
              hostId: 6,
              host: {},
              createdAt: "2021-09-25T02:00:00Z",
              updatedAt: "2021-09-25T02:00:00Z",
              attendeesCount: 1,
              url: "",
              isClosed: false,
            },
          ],
          message: "Fetched recent rooms successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchRoomsResponse>);
    }, 1500);
  });
};

export const fetchNumRoomsApi = (
  api: AxiosInstance,
  mock: boolean = false,
): Promise<AxiosResponse<GetNumRoomsResponse>> => {
  if (!mock) {
    return api.get<GetNumRoomsResponse>(`/rooms/count`);
  }

  return new Promise<AxiosResponse<GetNumRoomsResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            count: 1,
          },
          message: "Retrieved number of rooms successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<GetNumRoomsResponse>);
    }, 1500);
  });
};

export const createRoomApi = (
  api: AxiosInstance,
  roomData: Partial<IRoom>,
  inviteesId: string[],
  mock: boolean = false,
): Promise<AxiosResponse<CreateRoomResponse>> => {
  if (!mock) {
    return api.post<CreateRoomResponse>("/rooms", {
      room: roomData,
      invitees: inviteesId,
    });
  }

  return new Promise<AxiosResponse<CreateRoomResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            room: {
              id: "1",
              name: roomData.name,
              time: roomData.time,
              venue: roomData.venue,
              date: roomData.date,
              hostId: 6,
              host: {},
              createdAt: "2021-09-25T02:00:00Z",
              updatedAt: "2021-09-25T02:00:00Z",
              attendeesCount: 1,
              url: "",
              isClosed: false,
            },
            invites: [],
          },
          message: "Room created successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<CreateRoomResponse>);
    }, 1500);
  });
};

export const updateRoomApi = (
  api: AxiosInstance,
  roomId: string,
  data: UpdateRoomRequest,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.patch<ApiResponse>(`/rooms/${roomId}/edit`, {
      ...data,
    });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          message: "Room updated successfully",
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

export const joinRoomApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<RespondToInviteResponse>> => {
  if (!mock) {
    return api.patch<RespondToInviteResponse>(`/rooms/${roomId}/join`);
  }

  return new Promise<AxiosResponse<RespondToInviteResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            room: {
              id: roomId,
              name: "Test Room",
              time: "5:00pm",
              venue: "ntu hall 9",
              date: "2022-09-04T00:00:00Z",
              hostId: 6,
              host: {},
              createdAt: "2021-09-25T02:00:00Z",
              updatedAt: "2021-09-25T02:00:00Z",
              attendeesCount: 2,
              url: "",
              isClosed: false,
            },
            attendees: [],
          },
          message: "Joined room successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<RespondToInviteResponse>);
    }, 1500);
  });
};

export const respondToInviteApi = (
  api: AxiosInstance,
  roomId: string,
  accept: boolean,
  mock: boolean = false,
): Promise<AxiosResponse<RespondToInviteResponse>> => {
  if (!mock) {
    return api.patch<RespondToInviteResponse>(`/rooms/${roomId}`, { accept });
  }

  return new Promise<AxiosResponse<RespondToInviteResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {},
          message: "Responded to invite successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<RespondToInviteResponse>);
    }, 1500);
  });
};

export const closeRoomApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.patch<ApiResponse>(`/rooms/${roomId}/close`);
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          message: "Closed room successfully",
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

export const leaveRoomApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.patch<ApiResponse>(`/rooms/${roomId}/leave`);
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          message: "Left room successfully",
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

export const fetchRoomInvitesApi = (
  api: AxiosInstance,
  mock: boolean = false,
): Promise<AxiosResponse<FetchRoomInvitesResponse>> => {
  if (!mock) {
    return api.get<FetchRoomInvitesResponse>("/rooms/invites");
  }

  return new Promise<AxiosResponse<FetchRoomInvitesResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              id: 1,
              roomId: "1",
              room: {
                id: "1",
                name: "Test Room",
                time: "5:00pm",
                venue: "ntu hall 9",
                date: "2022-09-04T00:00:00Z",
                hostId: 6,
                host: {
                  username: "testuser",
                },
                createdAt: "2021-09-25T02:00:00Z",
                updatedAt: "2021-09-25T02:00:00Z",
                attendeesCount: 1,
                url: "",
                isClosed: false,
              },
            },
          ],
          message: "Fetched room invites successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchRoomInvitesResponse>);
    }, 1500);
  });
};

export const fetchNumRoomInvitesApi = (
  api: AxiosInstance,
  mock: boolean = false,
): Promise<AxiosResponse<FetchNumRoomInvitesResponse>> => {
  if (!mock) {
    return api.get<FetchNumRoomInvitesResponse>("/rooms/invites/count");
  }

  return new Promise<AxiosResponse<FetchNumRoomInvitesResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            count: 1,
          },
          message: "Fetched room invites successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchNumRoomInvitesResponse>);
    }, 1500);
  });
};

export const fetchRoomApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<FetchRoomResponse>> => {
  if (!mock) {
    return api.get<FetchRoomResponse>(`/rooms/${roomId}`);
  }

  return new Promise<AxiosResponse<FetchRoomResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {
            id: "1",
            name: "Test Room",
            time: "5:00pm",
            venue: "ntu hall 9",
            date: "2022-09-04T00:00:00Z",
            hostId: 6,
            host: {
              username: "testuser",
            },
            createdAt: "2021-09-25T02:00:00Z",
            updatedAt: "2021-09-25T02:00:00Z",
            attendeesCount: 1,
            url: "",
            isClosed: false,
          },
          message: "Fetched room successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchRoomResponse>);
    }, 1500);
  });
};

export const fetchRoomAttendeesApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<FetchRoomAttendeesResponse>> => {
  if (!mock) {
    return api.get<FetchRoomAttendeesResponse>(`/rooms/${roomId}/attendees`);
  }

  return new Promise<AxiosResponse<FetchRoomAttendeesResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              id: 1,
              username: "testuser",
              email: "test@test.com",
              password: "test",
              pictureUrl: "test.png",
              isEmailValid: true,
              isOnline: false,
              lastSeen: "2021-09-25T02:00:00Z",
              registeredAt: "2021-09-25T02:00:00Z",
            },
          ],
          message: "Fetched room attendees successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchRoomAttendeesResponse>);
    }, 1500);
  });
};

export const getUninvitedFriendsForRoomApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<FetchRoomAttendeesResponse>> => {
  if (!mock) {
    return api.get<FetchRoomAttendeesResponse>(`/rooms/${roomId}/uninvited`);
  }

  return new Promise<AxiosResponse<FetchRoomAttendeesResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              id: 1,
              username: "testuser",
              email: "test@test.com",
            },
          ],
          message: "Fetched uninvited friends successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchRoomAttendeesResponse>);
    }, 1500);
  });
};

export const inviteUsersToRoomApi = (
  api: AxiosInstance,
  roomId: string,
  inviteesId: string[],
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.post<ApiResponse>(`/rooms/${roomId}`, {
      invitees: inviteesId,
    });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          message: "Invited user to room successfully",
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

export const queryVenueApi = (
  api: AxiosInstance,
  query: string,
  mock: boolean = false,
): Promise<AxiosResponse<QueryVenueResponse>> => {
  if (!mock) {
    return api.get<QueryVenueResponse>(
      `/rooms/venues/search?query=${encodeURIComponent(query)}`,
    );
  }

  return new Promise<AxiosResponse<QueryVenueResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              name: "Test Venue",
              address: "123 Test Street, Singapore",
              googleMapsPlaceId: "ChIJN1t_tDeuEmsRUKwX8c4g2k0",
            },
          ],
          message: "Queried venue successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<QueryVenueResponse>);
    }, 1500);
  });
};
