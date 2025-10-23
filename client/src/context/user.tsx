import React, { createContext, useReducer } from "react";
import { UserContextType } from "../types/user";
import useContextWrapper from "../hooks/useContextWrapper";
import { BaseContextResponse, Optional } from "../types";
import { userService } from "../services/user.service";
import { AxiosError } from "axios";
import UserReducer, { INITIAL_USER_CTX_STATE } from "../reducers/user";
import { MinimalUserDto } from "../types/models";

export const FETCH_USER = "FETCH_USER";
export const FETCH_FRIENDS = "FETCH_FRIENDS";
export const ADD_FRIEND = "ADD_FRIEND";
export const REMOVE_FRIEND = "REMOVE_FRIEND";

const UserContext = createContext<Optional<UserContextType>>(null);

const UserProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [state, dispatch] = useReducer(UserReducer, INITIAL_USER_CTX_STATE);

  const setUser = (user: MinimalUserDto) => {
    dispatch({ type: FETCH_USER, payload: user });
  };

  const fetchFriends = async (userId: number): Promise<BaseContextResponse> => {
    try {
      const res = await userService.getFriends(userId.toString());
      dispatch({
        type: FETCH_FRIENDS,
        payload: res.data || [],
      });
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
      await userService.removeFriend(userId.toString(), friendId.toString());
      const updatedFriends = state.friends.filter(
        (friend) => friend.id !== friendId,
      );
      dispatch({ type: REMOVE_FRIEND, payload: updatedFriends });
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

export { useUserCtx, UserProvider };
