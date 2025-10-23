import {
  apiClient,
  type ApiQueryParams,
  type ApiResponse,
} from "../api/client";

export const transactionService = {
  getTransactions: async (params?: ApiQueryParams<"/transactions", "get">) => {
    const { data } = await apiClient.get<ApiResponse<"/transactions", "get">>(
      "/transactions",
      {
        params,
      },
    );
    return data;
  },

  settleTransaction: async (txId: string) => {
    const { data } = await apiClient.patch<
      ApiResponse<"/transactions/{txId}/settle", "patch">
    >(`/transactions/${txId}/settle`);
    return data;
  },
};
