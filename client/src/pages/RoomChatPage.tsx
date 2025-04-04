/* eslint-disable react-hooks/exhaustive-deps */
import React, { useEffect, useMemo, useRef, useState } from "react";
import RoomTopBar from "../components/top-bar/TopBarWithBackArrow";
import { channelTypes, useWs } from "../context/ws";
import { useUserCtx } from "../context/user";
import { fetchRoomMessageApi, sendMessageApi } from "../api/message";
import { api } from "../api";
import useMandatoryParam from "../hooks/useMandatoryParam";
import useLoadingAndError from "../hooks/useLoadingAndError";
import Spinner from "../components/Spinner";
import { useToast } from "../context/toast";
import { AxiosError } from "axios";

type Message = {
  id: number;
  userId: number;
  username: string;
  content: string;
  time: Date;
};

const RoomChatPage: React.FC = () => {
  const { loadingStates, startLoading, stopLoading } = useLoadingAndError();
  const [messages, setMessages] = useState<Message[]>([]);
  const [page, setPage] = useState<number>(1);
  const [pageCount, setPageCount] = useState<number>();
  const [isNewMessage, setIsNewMessage] = useState<boolean>(false);
  const { user } = useUserCtx();
  const { showToast } = useToast();
  const roomId = useMandatoryParam("roomId");
  const [subscribe, unsubscribe] = useWs();

  useEffect(() => {
    const channel = channelTypes.createMessageInChat(roomId);

    subscribe(channel, (message) => {
      console.log("[RoomChatPage] Received message", message);
      setMessages((prev) => [
        ...prev,
        {
          id: message.id,
          userId: Number(message.senderId),
          username: message.senderName,
          content: message.content,
          time: new Date(message.sentAt),
        },
      ]);
      setIsNewMessage(true);
    });
    console.log("[RoomChatPage] Subscribed to channel", channel);

    return () => {
      console.log("[RoomChatPage] Unsubscribing from channel", channel);
      unsubscribe(channel);
    };
  }, [roomId, subscribe, unsubscribe]);

  useEffect(() => {
    const fetchMessages = async () => {
      const res = await fetchRoomMessageApi(api, roomId, page);

      const { data } = res.data;

      console.log("[RoomChatPage] Messages fetched: ", data);
      const newMsgs = data.messages.map((msg) => ({
        id: msg.id,
        userId: msg.sender.id,
        username: msg.sender.username,
        content: msg.content,
        time: new Date(msg.sentAt),
      }));

      setIsNewMessage(page == 1 ? true : false);
      setMessages((prev) => {
        // deduplicate messages
        const allMessages = [...newMsgs, ...prev];
        const uniqueMessagesMap = new Map();
        allMessages.forEach((msg) => uniqueMessagesMap.set(msg.id, msg));
        // sort by time in chronological order
        return Array.from(uniqueMessagesMap.values()).sort(
          (a, b) => a.time.getTime() - b.time.getTime(),
        );
      });
      setPageCount(data.pageCount);
    };

    startLoading();
    fetchMessages()
      .then(() => stopLoading())
      .catch(() => {
        stopLoading();
        showToast("Failed to fetch messages", true);
      });
  }, [roomId, page]);

  const fetchMoreMessages = () => {
    console.log("[RoomChatPage] Fetching more messages...");
    if (page < (pageCount || 0)) {
      setPage((prevPage) => prevPage + 1);
    }
  };

  const handleSend = async (text: string) => {
    try {
      await sendMessageApi(api, roomId, text);
    } catch (error) {
      console.error("[RoomChatPage] Failed to send message", error);
      switch ((error as AxiosError).response?.status) {
        case 400:
          showToast("Invalid message", true);
          break;
        case 404:
          showToast("Room / User not found", true);
          break;
        case 500:
        default:
          showToast("Failed to send message", true);
          break;
      }
    }
  };

  return (
    <div className="h-full flex flex-col bg-gray-200">
      <RoomTopBar title="Chat" />
      {loadingStates[0] ? (
        <Spinner spinnerSize={{ width: "w-10", height: "h-10" }} />
      ) : (
        <ChatMessages
          messages={messages}
          currentUserId={user.id}
          isNewMessage={isNewMessage}
          fetchMore={fetchMoreMessages}
        />
      )}
      <ChatInput onSend={handleSend} />
    </div>
  );
};

const ChatMessages: React.FC<{
  messages: Message[];
  currentUserId: number;
  isNewMessage: boolean;
  fetchMore: () => void;
}> = ({ messages, currentUserId, isNewMessage, fetchMore }) => {
  const latestMessageRef = useRef<HTMLDivElement>(null);
  const messageListRef = useRef<HTMLDivElement>(null);
  const groupedMessages = useMemo(() => {
    const grouped: { date: string; messages: Message[] }[] = [];
    let currentDate = "";

    messages.forEach((msg) => {
      // Extract date from message time
      const dateString = msg.time.toLocaleDateString("en-US", {
        weekday: "long",
        month: "short",
        day: "numeric",
      });

      if (dateString !== currentDate) {
        // Start a new group for new date
        grouped.push({ date: dateString, messages: [msg] });
        currentDate = dateString;
      } else {
        // Add to current date group
        grouped[grouped.length - 1].messages.push(msg);
      }
    });

    console.log("[ChatMessages] Grouped messages: ", grouped);
    return grouped;
  }, [messages]);

  // scroll to latest message whenever new message is added
  useEffect(() => {
    if (latestMessageRef.current && isNewMessage) {
      latestMessageRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages, isNewMessage]);

  const handleScroll = () => {
    if (messageListRef.current?.scrollTop === 0) {
      fetchMore();
    }
  };

  return (
    <div
      ref={messageListRef}
      onScroll={handleScroll}
      className="flex-1 p-4 overflow-y-auto"
    >
      {groupedMessages.map((group, groupIndex) => (
        <div key={groupIndex} className="mb-6">
          <div className="flex justify-center mb-4">
            <span className="px-3 py-1 bg-gray-300 text-gray-700 text-sm rounded-full">
              {group.date}
            </span>
          </div>

          {group.messages.map((message, index) => (
            <div
              key={index}
              className={`max-w-[85%] w-fit mb-4 px-3 py-2 rounded-xl text-black bg-white border-[1.5px] border-secondary ${
                message.userId === currentUserId
                  ? "ml-auto rounded-br-none"
                  : "rounded-bl-none"
              }`}
            >
              <div
                className={`flex gap-3 ${
                  message.userId === currentUserId
                    ? "flex-row-reverse "
                    : "flex-row"
                }`}
              >
                <img
                  src="https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg"
                  alt="User Profile Picture"
                  className="w-6 h-6 rounded-full"
                />
                <div className="flex flex-col">
                  <div
                    className={`flex items-start justify-between gap-2 ${
                      message.userId === currentUserId
                        ? "flex-row-reverse"
                        : "flex-row"
                    }`}
                  >
                    <h3 className="text-md font-semibold leading-none">
                      {message.username}
                    </h3>
                    <span className="text-sm text-gray-500 leading-[1.275]">
                      {message.time.toLocaleTimeString("en-US", {
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </span>
                  </div>
                  <p className="text-lg text-wrap break-word leading-tight">
                    {message.content}
                  </p>
                </div>
              </div>

              {groupIndex === groupedMessages.length - 1 &&
              index === group.messages.length - 1 ? (
                <div ref={latestMessageRef} />
              ) : null}
            </div>
          ))}
        </div>
      ))}
    </div>
  );
};

const ChatInput: React.FC<{ onSend: (message: string) => void }> = ({
  onSend,
}) => {
  const [input, setInput] = useState("");

  const handleSend = () => {
    if (!input.trim()) {
      return;
    }

    onSend(input);
    setInput("");
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handleSend();
    }
  };

  return (
    <div className="w-full px-4 pt-3 pb-4 flex">
      <input
        type="text"
        className="flex-1 p-2 border border-gray-900 rounded-lg rounded-tr-none rounded-br-none text-black placeholder-black border-r-0 bg-primary focus:outline-none"
        value={input}
        onKeyDown={handleKeyDown}
        onChange={(e) => setInput(e.target.value)}
        placeholder="Type a message..."
      />
      <button
        onClick={handleSend}
        className="bg-secondary hover:bg-purple-800 text-white rounded-tl-none rounded-bl-none border-l-0 font-bold py-2 px-4 rounded"
      >
        Send
      </button>
    </div>
  );
};

export default RoomChatPage;
