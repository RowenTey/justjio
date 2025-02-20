import { AxiosInstance, AxiosResponse } from "axios";
import { ApiResponse } from ".";
import { ISubscription } from "../types/subscription";

export interface SubscriptionResponse extends ApiResponse {
  data: ISubscription;
}

export const createSubscriptionApi = (
  api: AxiosInstance,
  subscription: Partial<ISubscription>,
  mock: boolean = false,
): Promise<AxiosResponse<SubscriptionResponse>> => {
  if (!mock) {
    return api.post<SubscriptionResponse>("/subscriptions", subscription);
  }

  return new Promise<AxiosResponse<SubscriptionResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: { id: "92389183", ...subscription },
          message: "Subscription created successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<SubscriptionResponse>);
    }, 1500);
  });
};

export const removeSubscriptionApi = (
  api: AxiosInstance,
  subscriptionId: string,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.delete<ApiResponse>(`/subscriptions/${subscriptionId}`);
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {},
          message: "Subscription deleted successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<ApiResponse>);
    }, 1500);
  });
};
