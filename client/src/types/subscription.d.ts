import { SubscriptionDto } from "./models";

export interface SubscriptionState {
  subscription: Optional<SubscriptionDto>;
  isSubscribed: boolean;
}

export interface SubscriptionContextType {
  subscribe: () => Promise<boolean>;
  unsubscribe: () => Promise<boolean>;
  subscriptionState: SubscriptionState;
}
