import { useContext, Context } from "react";

const useContextWrapper = <T>(context: Context<T>) => {
  const ctx = useContext(context);
  if (ctx === undefined || ctx === null) {
    throw new Error(
      "useContextWrapper must be used within a corresponding Provider",
    );
  }

  return ctx;
};

export default useContextWrapper;
