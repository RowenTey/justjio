/* eslint-disable react-hooks/exhaustive-deps */
import { useEffect, useState } from "react";
import { useUserCtx } from "../context/user";
import { api } from "../api";
import { useToast } from "../context/toast";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { CheckIcon } from "@heroicons/react/24/solid";
import Spinner from "../components/Spinner";
import {
  getNotificationsApi,
  markNotificationAsReadApi,
} from "../api/notifications";
import { INotification } from "../types/notifications";

const NotificationsTopBar: React.FC = () => {
  return (
    <div className="relative top-0 flex h-[8%] items-center justify-center w-full py-4 px-6 bg-purple-200">
      <h1 className="text-xl font-bold text-secondary">Notifications</h1>
    </div>
  );
};

const NotificationsPage = () => {
  const { loadingStates, startLoading, stopLoading } = useLoadingAndError();
  const [notifications, setNotifications] = useState<INotification[]>([]);
  const { user } = useUserCtx();
  const { showToast } = useToast();

  useEffect(() => {
    const fetchNotifications = async () => {
      const res = await getNotificationsApi(api);
      setNotifications(res.data.data);
    };

    startLoading();
    fetchNotifications().then(() => stopLoading());
  }, [user.id]);

  const handleReadNotification = async (notificationId: number) => {
    try {
      await markNotificationAsReadApi(api, user.id, notificationId);
      const updatedNotifications = notifications.map((notification) => {
        if (notification.id === notificationId) {
          return { ...notification, isRead: true };
        }
        return notification;
      });
      setNotifications(updatedNotifications);
    } catch (error) {
      console.error(
        "An error occurred while marking notification as read: ",
        error,
      );
      showToast("Error occurred, please try again later.", true);
    }
  };

  return (
    <div className="h-full flex flex-col items-center gap-4 bg-gray-200">
      <NotificationsTopBar />

      <div className="w-full h-[85%] flex flex-col items-center gap-3">
        {loadingStates[0] ? (
          <Spinner spinnerSize={{ width: "w-10", height: "h-10" }} />
        ) : (
          <div
            className={`w-full h-full overflow-y-auto flex flex-col items-center gap-4 ${
              notifications.length > 0 ? "" : "justify-center"
            }`}
          >
            {notifications.length > 0 ? (
              notifications.map((notification) => (
                <div
                  key={notification.id}
                  className={`w-4/5 flex items-center justify-between py-2 px-3 ${
                    notification.isRead ? "bg-gray-100" : "bg-white"
                  } rounded-xl shadow-md`}
                >
                  <div className="flex flex-col justify-between w-[90%]">
                    <p className="text-base text-secondary font-semibold">
                      {notification.title}
                    </p>
                    <p className="text-sm text-black text-wrap">
                      {notification.content}
                    </p>
                  </div>
                  {!notification.isRead && (
                    <CheckIcon
                      className="w-5 h-5 text-green-500 cursor-pointer"
                      onClick={() => handleReadNotification(notification.id)}
                    />
                  )}
                </div>
              ))
            ) : (
              <p className="text-lg font-semibold text-gray-500">
                No notifications found
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default NotificationsPage;
