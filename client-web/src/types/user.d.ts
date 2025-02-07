import React from "react";

export interface BaseUserInfo {
	id: number;
	email: string;
	username: string;
}

export interface IUser {
	id: number;
	username: string;
	email: string;
	password: string;
	name?: string;
	phoneNum?: string;
	isEmailValid: boolean;
	isOnline: boolean;
	lastSeen: string;
	registeredAt: string;
	updatedAt: string;
}

export interface IFriendRequests {
	id: number;
	senderId: number;
	receiverId: number;
	status: string;
	sentAt: string;
	respondedAt: string | null;
	sender: IUser;
	receiver: IUser;
}

export type UserContextType = {
	user: BaseUserInfo;
	setUser: React.Dispatch<React.SetStateAction<BaseUserInfo>>;
};
