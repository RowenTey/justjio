/**
 * Type aliases for OpenAPI-generated schemas
 * Import these throughout your app instead of using the verbose components["schemas"]["..."] syntax
 */

import type { components } from "./api";

// User types
export type User = components["schemas"]["model.User"];
export type UserFriend = components["schemas"]["model.User"];

// Room types
export type Room = components["schemas"]["model.Room"];
export type RoomInviteDto = components["schemas"]["response.RoomInviteDto"];
export type RoomDto = components["schemas"]["response.RoomDto"];
export type RoomListDto = components["schemas"]["response.RoomListDto"];
export type SimplifiedRoomDto =
  components["schemas"]["response.SimplifiedRoomDto"];

// Message types
export type Message = components["schemas"]["model.Message"];

// Friend Request types
export type FriendRequestDto =
  components["schemas"]["response.FriendRequestDto"];

// Venue types
export type Venue = components["schemas"]["model_location.Venue"];

// Request types
export type LoginRequest = components["schemas"]["request.LoginRequest"];
export type SignUpRequest = components["schemas"]["request.SignUpRequest"];
export type GoogleAuthRequest =
  components["schemas"]["request.GoogleAuthRequest"];
export type CreateRoomRequest =
  components["schemas"]["request.CreateRoomRequest"];
export type EditRoomRequest = components["schemas"]["request.EditRoomRequest"];
export type CreateBillRequest =
  components["schemas"]["request.CreateBillRequest"];
export type ConsolidateBillsRequest =
  components["schemas"]["request.ConsolidateBillsRequest"];
export type CreateMessageRequest =
  components["schemas"]["request.CreateMessageRequest"];
export type InviteUserRequest =
  components["schemas"]["request.InviteUserRequest"];
export type RespondToRoomInviteRequest =
  components["schemas"]["request.RespondToRoomInviteRequest"];
export type SendOTPEmailRequest =
  components["schemas"]["request.SendOTPEmailRequest"];
export type VerifyOTPRequest =
  components["schemas"]["request.VerifyOTPRequest"];
export type ResetPasswordRequest =
  components["schemas"]["request.ResetPasswordRequest"];
export type UpdateUsernameRequest =
  components["schemas"]["request.UpdateUsernameRequest"];
export type ModifyFriendRequest =
  components["schemas"]["request.ModifyFriendRequest"];
export type RespondToFriendRequestRequest =
  components["schemas"]["request.RespondToFriendRequestRequest"];
export type CreateNotificationRequest =
  components["schemas"]["request.CreateNotificationRequest"];
export type CreateSubscriptionRequest =
  components["schemas"]["request.CreateSubscriptionRequest"];

// Response types
export type AuthResponse = components["schemas"]["response.AuthResponse"];
export type BillDto = components["schemas"]["response.BillDto"];
export type GetMessagesResponse =
  components["schemas"]["response.GetMessagesResponse"];
export type MessageDto = components["schemas"]["response.MessageDto"];
export type MinimalUserDto = components["schemas"]["response.MinimalUserDto"];
export type NotificationDto = components["schemas"]["response.NotificationDto"];
export type SubscriptionDto = components["schemas"]["response.SubscriptionDto"];
export type TransactionDto = components["schemas"]["response.TransactionDto"];
export type EmptyApiResponse = components["schemas"]["utils.EmptyApiResponse"];

// Helper type to ensure all required fields are present (converts optional to required)
export type Required<T> = {
  [P in keyof T]-?: T[P];
};

// Helper type for partial updates
export type PartialBy<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;
