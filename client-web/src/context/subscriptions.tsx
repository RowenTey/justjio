/* eslint-disable react-hooks/exhaustive-deps */
import React, { createContext, useEffect, useState } from "react";
import { useUserCtx } from "./user";
import useContextWrapper from "../hooks/useContextWrapper";
import { api } from "../api";
import {
  createSubscriptionApi,
  removeSubscriptionApi,
} from "../api/subscription";
import { ISubscription } from "../types/subscription";

interface SubscriptionState {
  subscription: ISubscription | null;
  isSubscribed: boolean;
}

interface SubscriptionContextType {
  subscribe: () => Promise<boolean>;
  unsubscribe: () => Promise<boolean>;
  subscriptionState: SubscriptionState;
}

const SubscriptionContext = createContext<SubscriptionContextType | null>(null);

const SubscriptionProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [subscriptionState, setSubscriptionState] = useState<SubscriptionState>(
    {
      subscription: null,
      isSubscribed: false,
    },
  );

  const { user } = useUserCtx();

  const urlBase64ToUint8Array = (base64String: string) => {
    const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
    const base64 = (base64String + padding)
      .replace(/-/g, "+")
      .replace(/_/g, "/");
    const rawData = window.atob(base64);
    return Uint8Array.from([...rawData].map((char) => char.charCodeAt(0)));
  };

  const subscribe = async (): Promise<boolean> => {
    console.log("[Push] Subscribing to push notifications...");
    if (!("serviceWorker" in navigator)) {
      console.error("[Push] Service workers are not supported by this browser");
      return false;
    }

    try {
      navigator.serviceWorker.register("service-worker.js");
      const registration = await navigator.serviceWorker.ready;
      const existingSubscription =
        await registration.pushManager.getSubscription();

      if (existingSubscription) {
        console.log(
          "[Push] Subscription already exists:",
          existingSubscription,
        );
        const existingSubJson = existingSubscription.toJSON();
        // TODO: Figure out this part
        // Most likely call backend?
        setSubscriptionState({
          subscription: {
            id: "",
            userId: 0,
            endpoint: existingSubscription.endpoint,
            auth: existingSubJson.keys?.auth || "",
            p256dh: existingSubJson.keys?.p256dh || "",
          },
          isSubscribed: true,
        });
        return true;
      }

      const publicVapidKey = import.meta.env.VITE_VAPID_PUBLIC_KEY || "";

      // Subscribe to push notifications
      // It will request permissions if not already granted
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(publicVapidKey),
      });
      const subJson = subscription.toJSON();
      console.log("[Push] Subscription object:", subJson);

      // Send subscription to backend
      const { data: res } = await createSubscriptionApi(api, {
        userId: user.id,
        endpoint: subJson.endpoint,
        auth: subJson.keys?.auth,
        p256dh: subJson.keys?.p256dh,
      });
      console.log("[Push] Subscription created: ", res.data);

      setSubscriptionState({
        subscription: res.data,
        isSubscribed: true,
      });
      console.log("[Push] Subscription object:", subscriptionState);

      return true;
    } catch (error) {
      console.error("[Push] Error subscribing to push notifications:", error);
      return false;
    }
  };

  const unsubscribe = async (): Promise<boolean> => {
    try {
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.getSubscription();

      if (subscription) {
        await subscription.unsubscribe();

        // Notify backend about unsubscription
        console.log("[Push] Subscription object:", subscriptionState);
        subscriptionState.subscription &&
          (await removeSubscriptionApi(api, subscriptionState.subscription.id));

        setSubscriptionState({
          subscription: null,
          isSubscribed: false,
        });
      }

      console.log("[Push] Unsubscribed from push notifications");
      return true;
    } catch (error) {
      console.error(
        "[Push] Error unsubscribing from push notifications:",
        error,
      );
      return false;
    }
  };

  // subscribe when user is logged in
  useEffect(() => {
    if (user && user.id !== -1) subscribe();
  }, [user]);

  return (
    <SubscriptionContext.Provider
      value={{
        subscribe,
        unsubscribe,
        subscriptionState,
      }}
    >
      {children}
    </SubscriptionContext.Provider>
  );
};

const useSubscription = () => useContextWrapper(SubscriptionContext);

export { useSubscription, SubscriptionProvider };
