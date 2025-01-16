/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, createContext, useRef } from "react";
import { useUserCtx } from "./user";
import { useAuth } from "./auth";
import useContextWrapper from "../hooks/useContextWrapper";

const channelTypes = {
	createMessageInChat: (roomId: string) => `CREATE_MESSAGE_${roomId}`,
	deleteMessageInChat: (roomId: string) => `DELETE_MESSAGE_${roomId}`,
	createMessage: () => "CREATE_MESSAGE",
};

type ChannelCallback = (data: any) => void;

interface Channels {
	[key: string]: ChannelCallback;
}

const WebSocketContext = createContext<
	[
		subscribe: (channel: string, callback: ChannelCallback) => void,
		unsubscribe: (channel: string) => void
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

	// connect to ws server when user is logged in
	useEffect(() => {
		if (!user || user.uid === -1) {
			ws.current?.close();
			return;
		}

		ws.current = new WebSocket(
			`${import.meta.env.VITE_WS_URL}?token=${getAccessToken()}`
		);

		ws.current.onopen = () => {
			console.log("[WS] Opened WS connection");
		};

		ws.current.onclose = () => {
			console.log("[WS] Closed WS connection");
		};

		ws.current.onmessage = (message) => {
			const { type, data } = JSON.parse(message.data);
			const roomChannel = `${type}_${data.room_id}`;

			console.log("[WS] Received message", type, data, roomChannel);

			// users currently in room => subscribed to chat channel
			if (channels.current[roomChannel]) channels.current[roomChannel](data);

			// not subsribed to chat channel yet
			else channels.current[type]?.(data);
		};

		return () => {
			ws.current?.close();
		};
	}, [getAccessToken, user]);

	return (
		<WebSocketContext.Provider value={[subscribe, unsubscribe]}>
			{children}
		</WebSocketContext.Provider>
	);
};

const useWs = () => useContextWrapper(WebSocketContext);

export { useWs, WebSocketProvider, channelTypes };
