import { BaseContextResponse } from ".";
import { MinimalUserDto } from "./models";

export interface UserState {
  user: MinimalUserDto;
  friends: MinimalUserDto[];
}

export type UserContextType = {
  user: MinimalUserDto;
  setUser: (user: MinimalUserDto) => void;
  friends: MinimalUserDto[];
  fetchFriends: (userId: number) => Promise<BaseContextResponse>;
  removeFriend: (
    userId: number,
    friendId: number,
  ) => Promise<BaseContextResponse>;
};

type UserActionTypes =
  | {
      type: "FETCH_FRIENDS" | "ADD_FRIEND" | "REMOVE_FRIEND";
      payload: MinimalUserDto[];
    }
  | { type: "FETCH_USER"; payload: MinimalUserDto };
