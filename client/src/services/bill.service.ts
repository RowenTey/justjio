import { apiClient, type ApiRequest, type ApiResponse } from "../api/client";

export const billService = {
  getBillsByRoom: async (roomId: string) => {
    const { data } = await apiClient.get<ApiResponse<"/bills", "get">>(
      "/bills",
      { params: { roomId } },
    );
    return data;
  },

  createBill: async (billData: ApiRequest<"/bills", "post">) => {
    const { data } = await apiClient.post<ApiResponse<"/bills", "post">>(
      "/bills",
      billData,
    );
    return data;
  },

  consolidateBills: async (
    consolidationData: ApiRequest<"/bills/consolidate", "post">,
  ) => {
    const { data } = await apiClient.post<
      ApiResponse<"/bills/consolidate", "post">
    >("/bills/consolidate", consolidationData);
    return data;
  },
};
