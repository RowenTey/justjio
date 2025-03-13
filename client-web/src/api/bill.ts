import { AxiosInstance, AxiosResponse } from "axios";
import { ApiResponse } from ".";
import { IBill } from "../types/bill";

interface CreateBillRequest {
  name: string;
  amount: number;
  includeOwner: boolean;
  roomId: string;
  payers: number[];
}

interface FetchBillResponse extends ApiResponse {
  data: IBill[];
}

interface IsRoomBillConsolidatedResponse extends ApiResponse {
  data: {
    isConsolidated: boolean;
  };
}

export const createBillApi = (
  api: AxiosInstance,
  billData: CreateBillRequest,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.post<ApiResponse>("/bills", billData);
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {},
          message: "Bill created successfully",
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

export const fetchBillApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<FetchBillResponse>> => {
  if (!mock) {
    return api.get<FetchBillResponse>(`/bills?roomId=${roomId}`);
  }

  return new Promise<AxiosResponse<FetchBillResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: [
            {
              id: "1",
              name: "Test Bill",
              amount: 100,
              includeOwner: true,
              date: "2021-10-10",
              roomId: "1",
              payers: [1],
              ownerId: 1,
              owner: {
                id: 1,
                username: "testuser",
              },
            },
          ],
          message: "Fetched bills successfully",
          status: "success",
        },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {},
      } as AxiosResponse<FetchBillResponse>);
    }, 1500);
  });
};

export const consolidateBillApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<ApiResponse>> => {
  if (!mock) {
    return api.post<ApiResponse>("/bills/consolidate", { roomId });
  }

  return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
    setTimeout(() => {
      resolve({
        data: {
          data: {},
          message: "Bill consolidated successfully",
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

export const isRoomBillConsolidatedApi = (
  api: AxiosInstance,
  roomId: string,
  mock: boolean = false,
): Promise<AxiosResponse<IsRoomBillConsolidatedResponse>> => {
  if (!mock) {
    return api.get<IsRoomBillConsolidatedResponse>(
      `/bills/consolidate/${roomId}`,
    );
  }

  return new Promise<AxiosResponse<IsRoomBillConsolidatedResponse>>(
    (resolve) => {
      setTimeout(() => {
        resolve({
          data: {
            data: {
              isConsolidated: false,
            },
            message: "Retrieved consolidation status successfully",
            status: "success",
          },
          status: 200,
          statusText: "OK",
          headers: {},
          config: {},
        } as AxiosResponse<IsRoomBillConsolidatedResponse>);
      }, 1500);
    },
  );
};
