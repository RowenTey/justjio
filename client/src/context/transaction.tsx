import React, { createContext, useReducer } from "react";
import useContextWrapper from "../hooks/useContextWrapper";
import { TransactionContextType } from "../types/transaction";
import TransactionReducer, {
  INITIAL_TRANSACTION_CTX_STATE,
} from "../reducers/transaction";
import { transactionService } from "../services/transaction.service";
import { useUserCtx } from "./user";
import { BaseContextResponse, Optional } from "../types";
import { AxiosError } from "axios";
import { TransactionDto } from "../types/models";

export const FETCH_TRANSACTIONS = "FETCH_TRANSACTIONS";
export const SETTLE_TRANSACTION = "SETTLE_TRANSACTION";

const TransactionContext =
  createContext<Optional<TransactionContextType>>(null);

const TransactionProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [state, dispatch] = useReducer(
    TransactionReducer,
    INITIAL_TRANSACTION_CTX_STATE,
  );
  const { user } = useUserCtx();

  const fetchTransactions = async (): Promise<BaseContextResponse> => {
    try {
      const response = await transactionService.getTransactions();
      const toPay = new Array<TransactionDto>();
      const toReceive = new Array<TransactionDto>();

      response.data!.forEach((tx: TransactionDto) => {
        if (tx.payer.id === user.id) {
          toPay.push(tx);
        } else if (tx.payee.id === user.id) {
          toReceive.push(tx);
        }
      });

      const payload = {
        toPay: toPay,
        toReceive: toReceive,
      };
      dispatch({ type: FETCH_TRANSACTIONS, payload });
      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to fetch transactions", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
  };

  const settleTransaction = async (
    transactionId: number,
  ): Promise<BaseContextResponse> => {
    try {
      await transactionService.settleTransaction(transactionId.toString());
      dispatch({ type: SETTLE_TRANSACTION, payload: transactionId });
      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to settle transaction", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
  };

  return (
    <TransactionContext.Provider
      value={{
        toPay: state.toPay,
        toReceive: state.toReceive,
        fetchTransactions,
        settleTransaction,
      }}
    >
      {children}
    </TransactionContext.Provider>
  );
};

const useTransactionCtx = () => useContextWrapper(TransactionContext);

export { useTransactionCtx, TransactionProvider };
