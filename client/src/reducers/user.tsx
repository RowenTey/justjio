import { FETCH_FRIENDS, FETCH_USER, REMOVE_FRIEND } from "../context/user";
import { UserActionTypes, UserState } from "../types/user";

export const initialUserState: UserState = {
  user: {
    id: -1,
    email: "",
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
    case FETCH_FRIENDS:
      return {
        ...state,
        friends: payload.data,
      };
    case REMOVE_FRIEND:
      return {
        ...state,
        friends: payload.data,
      };
    default:
      throw new Error(`No case for type ${type} found in UserReducer.`);
  }
};

export default UserReducer;
