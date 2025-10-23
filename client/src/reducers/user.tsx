import {
  ADD_FRIEND,
  FETCH_FRIENDS,
  FETCH_USER,
  REMOVE_FRIEND,
} from "../context/user";
import { UserActionTypes, UserState } from "../types/user";

export const INITIAL_USER_CTX_STATE: UserState = {
  user: {
    id: -1,
    username: "",
    pictureUrl: "",
  },
  friends: [],
};

const UserReducer = (state: UserState, action: UserActionTypes): UserState => {
  const { type, payload } = action;

  switch (type) {
    case FETCH_USER:
      return {
        ...state,
        user: payload,
      };
    case ADD_FRIEND:
    case REMOVE_FRIEND:
    case FETCH_FRIENDS:
      return {
        ...state,
        friends: payload,
      };
    default:
      throw new Error(`No case for type ${type} found in UserReducer.`);
  }
};

export default UserReducer;
