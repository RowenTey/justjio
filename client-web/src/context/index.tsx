import { ElementType } from "react";
import { AuthProvider } from "./auth";
import { UserProvider } from "./user";
import { RoomProvider } from "./room";
import { WebSocketProvider } from "./ws";
import { TransactionProvider } from "./transaction";
import { ToastProvider } from "./toast";

type ProvidersType = [ElementType, Record<string, unknown>];
type ChildrenType = { children: React.ReactNode };

const buildProvidersTree = (componentsWithProps: Array<ProvidersType>) => {
  const initialComponent = ({ children }: ChildrenType) => <>{children}</>;
  return componentsWithProps.reduce(
    (
      AccumulatedComponents: ElementType,
      [Provider, props = {}]: ProvidersType,
    ) => {
      return ({ children }: ChildrenType) => {
        return (
          <AccumulatedComponents>
            <Provider {...props}>{children}</Provider>
          </AccumulatedComponents>
        );
      };
    },
    initialComponent,
  );
};

// order is important
const providers: ProvidersType[] = [
  [UserProvider, {}],
  [AuthProvider, {}],
  [WebSocketProvider, {}],
  [ToastProvider, {}],
  [RoomProvider, {}],
  [TransactionProvider, {}],
];

export const AppContextProvider = buildProvidersTree(providers);
