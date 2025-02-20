/* eslint-disable react-hooks/exhaustive-deps */
import { useEffect, useState } from "react";
import { IFriendRequests } from "../types/user";
import { fetchFriendRequestsApi, respondToFriendRequestApi } from "../api/user";
import { useUserCtx } from "../context/user";
import { api } from "../api";
import { useToast } from "../context/toast";
import TopBarWithBackArrow from "../components/top-bar/TopBarWithBackArrow";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { CheckIcon, XMarkIcon } from "@heroicons/react/24/solid";
import Spinner from "../components/Spinner";
import { formatDate } from "../utils/date";
import { AxiosError } from "axios";

const FriendRequestsPage = () => {
  const { loadingStates, startLoading, stopLoading } = useLoadingAndError();
  const [friendRequests, setFriendRequests] = useState<IFriendRequests[]>([]);
  const { user } = useUserCtx();
  const { showToast } = useToast();

  useEffect(() => {
    const fetchFriends = async () => {
      const res = await fetchFriendRequestsApi(api, user.id, "pending");
      setFriendRequests(res.data.data);
    };

    startLoading();
    fetchFriends().then(() => stopLoading());
  }, [user.id]);

  const handleRespondToFriendRequest = async (
    requestId: number,
    accept: boolean,
  ) => {
    try {
      await respondToFriendRequestApi(
        api,
        user.id,
        requestId,
        accept ? "accept" : "reject",
      );

      showToast("Friend request responded to successfully!", false);
      setFriendRequests((prev) =>
        prev.filter((friendRequest) => friendRequest.id !== requestId),
      );
    } catch (err) {
      console.error(
        "An error occurred while responding to friend request: ",
        err,
      );
      switch ((err as AxiosError).response?.status) {
        case 400:
          showToast("Bad request, please check request body.", true);
          break;
        case 404:
          showToast("User not found, please try again later.", true);
          break;
        case 409:
          showToast(
            (err as AxiosError<{ message: string }>).response?.data?.message ||
              "An error occurred, please try again later.",
            true,
          );
          break;
        case 500:
        default:
          showToast("An error occurred, please try again later.", true);
          break;
      }
    }
  };

  return (
    <div className="h-full flex flex-col items-center gap-4 bg-gray-200">
      <TopBarWithBackArrow title="Friend Requests" />

      <div className="w-full h-full flex flex-col items-center px-4 gap-3">
        {loadingStates[0] ? (
          <Spinner spinnerSize={{ width: "w-10", height: "h-10" }} />
        ) : (
          <div
            className={`w-full h-[85%] overflow-y-auto flex flex-col items-center ${
              friendRequests.length > 0 ? "" : "justify-center"
            } gap-4`}
          >
            {friendRequests.length > 0 ? (
              friendRequests.map((friendRequest) => (
                <div
                  key={friendRequest.id}
                  className="w-4/5 flex items-center justify-between py-2 px-3 bg-white rounded-xl shadow-md"
                >
                  <div className="flex items-center gap-2">
                    <img
                      src="https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg"
                      alt="Profile Image"
                      className="w-7 h-7 rounded-full"
                    />
                    <div className="flex flex-col">
                      <p className="text-black">
                        {friendRequest.sender.username}
                      </p>
                      <p className="text-xs text-gray-500">
                        Sent on: {formatDate(friendRequest.sentAt)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <CheckIcon
                      onClick={() =>
                        handleRespondToFriendRequest(friendRequest.id, true)
                      }
                      className="h-6 w-6 text-success cursor-pointer hover:scale-110"
                    />
                    <XMarkIcon
                      onClick={() =>
                        handleRespondToFriendRequest(friendRequest.id, true)
                      }
                      className="h-6 w-6 text-error cursor-pointer hover:scale-110"
                    />
                  </div>
                </div>
              ))
            ) : (
              <p className="text-lg font-semibold text-gray-500">
                No friend requests found
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default FriendRequestsPage;
