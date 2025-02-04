import { AxiosInstance, AxiosResponse } from "axios";
import { ITransaction } from "../types/transaction";
import { ApiResponse } from ".";

interface FetchTransactionResponse extends ApiResponse {
	data: ITransaction[];
}

export const fetchTransactionsApi = async (
	api: AxiosInstance,
	isPaid: boolean = false,
	mock: boolean = false
): Promise<AxiosResponse<FetchTransactionResponse>> => {
	if (!mock) {
		return api.get<FetchTransactionResponse>(`/transactions?isPaid=${isPaid}`);
	}

	return new Promise<AxiosResponse<FetchTransactionResponse>>((resolve) => {
		setTimeout(() => {
			resolve({
				data: {
					data: [
						{
							id: 1,
							consolidationId: 1,
							payerId: 1,
							payer: {
								id: 1,
								username: "John Doe",
								email: "",
							},
							payeeId: 2,
							payee: {
								id: 2,
								username: "Jane Doe",
								email: "",
							},
							amount: 100,
							isPaid: false,
							paidOn: "",
						},
					],
					message: "Fetched bills successfully",
					status: "success",
				},
				status: 200,
				statusText: "OK",
				headers: {},
				config: {},
			} as AxiosResponse<FetchTransactionResponse>);
		}, 1500);
	});
};

export const settleTransactionApi = async (
	api: AxiosInstance,
	transactionId: number,
	mock: boolean = false
): Promise<AxiosResponse<ApiResponse>> => {
	if (!mock) {
		return api.patch<ApiResponse>(`/transactions/${transactionId}/settle`);
	}

	return new Promise<AxiosResponse<ApiResponse>>((resolve) => {
		setTimeout(() => {
			resolve({
				data: {
					message: "Transaction settled successfully",
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
