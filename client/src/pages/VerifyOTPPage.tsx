import { useEffect, useState } from "react";
import useLoadingAndError from "../hooks/useLoadingAndError";
import Spinner from "../components/Spinner";
import { useLocation, useNavigate } from "react-router-dom";
import { sendOtpEmailApi, verifyOtpApi } from "../api/auth";
import { api } from "../api";
import { useToast } from "../context/toast";
import { AxiosError } from "axios";

const VerifyOTPPage = () => {
  const { loadingStates, startLoading, stopLoading, errorStates, setErrorMsg } =
    useLoadingAndError();
  const { state } = useLocation();
  const email = (state?.email as string) || "";
  const from = (state?.from as string) || "";
  const [otp, setOtp] = useState("");
  const [canResend, setCanResend] = useState(false);
  const [countdown, setCountdown] = useState(30);
  const { showToast } = useToast();
  const navigate = useNavigate();

  // auto-focus first input on page load
  useEffect(() => {
    const firstInput = document.querySelector(
      `input[name="otp-0"]`,
    ) as HTMLInputElement;
    if (firstInput) firstInput.focus();
  }, []);

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(
        () => setCountdown((prevCount) => prevCount - 1),
        1000,
      );
      return () => clearTimeout(timer);
    } else if (countdown === 0) {
      setCanResend(true);
    }
  }, [countdown]);

  const handleResend = async () => {
    if (!canResend) return;

    setCanResend(false);
    setCountdown(30);

    try {
      await sendOtpEmailApi(api, email, "verify-email");
      showToast("OTP sent successfully!", false);
    } catch (error) {
      console.error(error);
      switch ((error as AxiosError).response?.status) {
        case 400:
          setErrorMsg("Bad request. Please check request body.");
          break;
        case 404:
          setErrorMsg("User not found.");
          break;
        case 409:
          setErrorMsg("User already verified.");
          break;
        case 500:
        default:
          setErrorMsg("An error occurred. Please try again later.");
          break;
      }
    }
  };

  const handleVerify = async () => {
    if (otp.length !== 6) {
      setErrorMsg("Please enter a valid 6-digit OTP.");
      return;
    } else if (email === "") {
      setErrorMsg("Invalid email received, please try again later.");
      return;
    }

    startLoading(0);
    try {
      console.log("[VerifyOTPPage] Verifying OTP:", otp, email);
      await verifyOtpApi(api, email, otp);
      showToast("Email verified successfully!", false);
      setTimeout(() => {
        if (from === "/password/forgot")
          navigate("/password/reset", { state: { email, from: "/otp" } });
        else navigate("/login", { state: { from: "/otp" } });
      }, 1000);
    } catch (error) {
      console.error(error);
      switch ((error as AxiosError).response?.status) {
        case 400:
          setErrorMsg("Bad request. Please check request body.");
          break;
        case 404:
          setErrorMsg("User not found.");
          break;
        case 500:
        default:
          setErrorMsg("An error occurred. Please try again later.");
          break;
      }
    } finally {
      stopLoading(0);
    }
  };

  return (
    <div className="h-full flex flex-col justify-center items-center xs:border-y-1 border-black overflow-y-auto bg-primary py-4">
      <img src="/favicon.svg" alt="JustJio Logo" className="w-32 h-32" />

      <h1 className="text-4xl font-bold text-secondary mt-3">Verify OTP</h1>
      <p className="w-4/5 text-base text-center text-wrap text-tertiary leading-tight mt-2">
        Please enter the 6-digit code sent to your registered email.
      </p>

      <div className="w-full flex flex-col gap-3 items-center justify-center mt-4">
        <div className="flex gap-2">
          {[...Array(6)].map((_, index) => (
            <input
              key={index}
              type="text"
              maxLength={1}
              value={otp[index] || ""}
              onChange={(e) => {
                // replace non-numeric characters
                const value = e.target.value.replace(/\D/g, "");
                const newOtp = otp.split("");
                newOtp[index] = value;
                setOtp(newOtp.join(""));

                // Auto-focus next input
                if (value && index < 5) {
                  const nextInput = document.querySelector(
                    `input[name="otp-${index + 1}"]`,
                  ) as HTMLInputElement;
                  if (nextInput) nextInput.focus();
                }
              }}
              onKeyDown={(e) => {
                // Handle backspace to focus previous input
                if (e.key === "Backspace" && !otp[index] && index > 0) {
                  const prevInput = document.querySelector(
                    `input[name="otp-${index - 1}"]`,
                  ) as HTMLInputElement;
                  if (prevInput) prevInput.focus();
                }
                // Handle enter key on last input
                if (e.key === "Enter" && index === 5) {
                  handleVerify();
                }
              }}
              name={`otp-${index}`}
              className="w-12 h-12 text-center border text-xl text-black font-bold bg-white border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-secondary"
              inputMode="numeric"
              pattern="[0-9]*"
            />
          ))}
        </div>

        {errorStates[0] && (
          <p className="text-error text-md font-semibold text-wrap text-center leading-tight">
            {errorStates[0]}
          </p>
        )}

        <button
          className={`bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full cursor-pointer w-2/5 ${errorStates[0] ? "" : "mt-3"}
					}`}
          onClick={handleVerify}
        >
          {loadingStates[0] ? (
            <Spinner
              spinnerColor="border-white"
              spinnerSize={{ width: "w-6", height: "h-6" }}
            />
          ) : (
            "Verify"
          )}
        </button>

        <p className="w-4/5 text-sm text-center text-wrap text-tertiary leading-tight mt-1">
          Didn't receive an email?{" "}
          {canResend ? (
            <span className="underline cursor-pointer" onClick={handleResend}>
              Try again
            </span>
          ) : (
            <span className="text-gray-600">Try again in {countdown}s</span>
          )}
        </p>
      </div>
    </div>
  );
};

export default VerifyOTPPage;
