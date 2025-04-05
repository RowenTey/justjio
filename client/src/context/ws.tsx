/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, createContext, useRef } from "react";
import { useUserCtx } from "./user";
import { useAuth } from "./auth";
import useContextWrapper from "../hooks/useContextWrapper";

const channelTypes = {
  createMessageInChat: (roomId: string) => `CREATE_MESSAGE_${roomId}`,
  createMessage: () => "CREATE_MESSAGE",
};

type ChannelCallback = (data: any) => void;

interface Channels {
  [key: string]: ChannelCallback;
}

const WebSocketContext = createContext<
  [
    subscribe: (channel: string, callback: ChannelCallback) => void,
    unsubscribe: (channel: string) => void,
  ]
>([() => {}, () => {}]);

const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const ws = useRef<WebSocket | null>(null);
  const channels = useRef<Channels>({});
  const { user } = useUserCtx();
  const { getAccessToken } = useAuth();

  const subscribe = (channel: string, callback: ChannelCallback) => {
    channels.current[channel] = callback;
  };

  const unsubscribe = (channel: string) => {
    delete channels.current[channel];
  };

  const connectWebSocket = () => {
    if (!user || user.id === -1) {
      ws.current?.close();
      return;
    }

    ws.current = new WebSocket(
      `${import.meta.env.VITE_WS_URL}?token=${getAccessToken()}`,
    );

    ws.current.onopen = () => {
      console.log("[WS] Opened WS connection");
    };

    ws.current.onclose = () => {
      console.log("[WS] Closed WS connection");
    };

    ws.current.onmessage = (message) => {
      console.log("[WS] Received message", JSON.parse(message.data));
      const { type, data } = JSON.parse(message.data);
      const roomChannel = `${type}_${data.roomId}`;

      console.log("[WS] Received message", type, data, roomChannel);

      // users currently in room => subscribed to chat channel
      if (channels.current[roomChannel]) channels.current[roomChannel](data);
      // not subscribed to chat channel yet
      else channels.current[type]?.(data);
    };
  };

  // connect to ws server when user is logged in
  useEffect(() => {
    connectWebSocket();

    return () => {
      ws.current?.close();
    };
  }, [user]);

  // Periodically check the WebSocket connection status and reconnect if needed
  useEffect(() => {
    // Check every 5 seconds
    const interval = setInterval(() => {
      if (
        user &&
        user.id !== -1 &&
        ws.current?.readyState === WebSocket.CLOSED
      ) {
        console.log("[WS] Reconnecting WS connection");
        connectWebSocket();
      }
    }, 5000);

    return () => clearInterval(interval);
  }, [user]);

  return (
    <WebSocketContext.Provider value={[subscribe, unsubscribe]}>
      {children}
    </WebSocketContext.Provider>
  );
};

const useWs = () => useContextWrapper(WebSocketContext);

export { useWs, WebSocketProvider, channelTypes };
