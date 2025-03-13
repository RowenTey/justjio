import { BaseContextResponse } from ".";

export interface BaseUserInfo {
  id: number;
  email: string;
  username: string;
  pictureUrl: string;
}

export interface IUser {
  id: number;
  username: string;
  email: string;
  password: string;
  pictureUrl: string;
  isEmailValid: boolean;
  isOnline: boolean;
  lastSeen: string;
  registeredAt: string;
  updatedAt: string;
}

export interface IFriendRequests {
  id: number;
  senderId: number;
  receiverId: number;
  status: string;
  sentAt: string;
  respondedAt: string | null;
  sender: IUser;
  receiver: IUser;
}

export interface UserState {
  user: BaseUserInfo;
  friends: IUser[];
}

export type UserContextType = {
  user: BaseUserInfo;
  setUser: (user: BaseUserInfo) => void;
  friends: IUser[];
  fetchFriends: (userId: number) => Promise<BaseContextResponse>;
  removeFriend: (
    userId: number,
    friendId: number,
  ) => Promise<BaseContextResponse>;
};

interface FriendsPayload {
  data: IUser[];
}

type UserActionTypes =
  | {
      type: "FETCH_FRIENDS" | "ADD_FRIEND" | "REMOVE_FRIEND";
      payload: FriendsPayload;
    }
  | { type: "FETCH_USER"; payload: BaseUserInfo };
