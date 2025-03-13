import React, { createContext, useReducer } from "react";
import { BaseUserInfo, UserContextType } from "../types/user";
import useContextWrapper from "../hooks/useContextWrapper";
import { BaseContextResponse } from "../types";
import { api } from "../api";
import { fetchFriendsApi, removeFriendApi } from "../api/user";
import { AxiosError } from "axios";
import UserReducer, { initialUserState } from "../reducers/user";

export const FETCH_USER = "FETCH_USER";
export const FETCH_FRIENDS = "FETCH_FRIENDS";
export const ADD_FRIEND = "ADD_FRIEND";
export const REMOVE_FRIEND = "REMOVE_FRIEND";

const UserContext = createContext<UserContextType | null>(null);

const UserProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [state, dispatch] = useReducer(UserReducer, initialUserState);

  const setUser = (user: BaseUserInfo) => {
    dispatch({ type: FETCH_USER, payload: user });
  };

  const fetchFriends = async (userId: number): Promise<BaseContextResponse> => {
    try {
      const { data: res } = await fetchFriendsApi(api, userId);
      dispatch({ type: FETCH_FRIENDS, payload: res });
      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to fetch friends", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
  };

  const removeFriend = async (
    userId: number,
    friendId: number,
  ): Promise<BaseContextResponse> => {
    try {
      await removeFriendApi(api, userId, friendId);
      const updatedFriends = state.friends.filter(
        (friend) => friend.id !== friendId,
      );
      dispatch({ type: REMOVE_FRIEND, payload: { data: updatedFriends } });
      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to remove friend", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
  };

  return (
    <UserContext.Provider
      value={{
        user: state.user,
        setUser,
        friends: state.friends,
        fetchFriends,
        removeFriend,
      }}
    >
      {children}
    </UserContext.Provider>
  );
};

const useUserCtx = () => useContextWrapper(UserContext);

export { useUserCtx, initialUserState, UserProvider };
