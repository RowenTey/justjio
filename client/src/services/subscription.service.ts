import { apiClient, type ApiRequest, type ApiResponse } from "../api/client";

export const subscriptionService = {
  createSubscription: async (
    subscriptionData: ApiRequest<"/subscriptions", "post">,
  ) => {
    const { data } = await apiClient.post<
      ApiResponse<"/subscriptions", "post">
    >("/subscriptions", subscriptionData);
    return data;
  },

  getSubscriptionByEndpoint: async (endpoint: string) => {
    const { data } = await apiClient.get<
      ApiResponse<"/subscriptions/{endpoint}", "get">
    >(`/subscriptions/${endpoint}`);
    return data;
  },

  deleteSubscription: async (subId: string) => {
    const { data } = await apiClient.delete<
      ApiResponse<"/subscriptions/{subId}", "delete">
    >(`/subscriptions/${subId}`);
    return data;
  },
};
