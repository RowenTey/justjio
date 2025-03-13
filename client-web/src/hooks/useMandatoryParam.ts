import { useParams } from "react-router-dom";

function useMandatoryParam(paramName: string): string {
  const params = useParams();
  const paramValue = params[paramName];

  if (!paramValue) {
    console.error(`Missing required parameter: ${paramName}`);
    throw new Error(`Missing required parameter: ${paramName}`);
  }

  return paramValue;
}

export default useMandatoryParam;
