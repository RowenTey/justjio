import { SubmitHandler, useForm } from "react-hook-form";
import useLoadingAndError from "../hooks/useLoadingAndError";
import InputField from "../components/InputField";
import Spinner from "../components/Spinner";
import { sendOtpEmailApi } from "../api/auth";
import { api } from "../api";
import { Link, useNavigate } from "react-router-dom";
import { AxiosError } from "axios";

const ForgotPasswordPage = () => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<{ email: string }>();
  const { loadingStates, startLoading, stopLoading, errorStates, setErrorMsg } =
    useLoadingAndError();
  const navigate = useNavigate();

  const onSubmit: SubmitHandler<{ email: string }> = async (data) => {
    startLoading();
    try {
      await sendOtpEmailApi(api, data.email, "reset-password");
      navigate("/otp", {
        state: {
          email: data.email,
          from: "/forgotPassword",
        },
      });
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
      stopLoading();
    }
  };

  return (
    <div className="h-full flex flex-col justify-center items-center xs:border-y-1 border-black overflow-y-auto bg-primary py-4">
      <img src="/favicon.svg" alt="JustJio Logo" className="w-32 h-32" />

      <h1 className="text-4xl font-bold text-secondary mt-3">
        Forgot Password
      </h1>
      <p className="w-4/5 text-base text-center text-wrap text-tertiary leading-tight mt-2">
        Please enter the email used to register your account.
      </p>

      <form
        onSubmit={handleSubmit(onSubmit)}
        id="forgot-password-form"
        className="flex flex-col items-center gap-3 p-2 w-[70%] mt-2"
      >
        <InputField
          name="email"
          type="email"
          placeholder="Enter your email"
          register={register}
          errors={errors}
          validation={{
            required: "Email is required",
            pattern: {
              value: /^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$/,
              message: "Enter a valid email address",
            },
          }}
        />

        {errorStates[0] && (
          <p className="text-error text-md font-semibold text-wrap text-center leading-tight">
            {errorStates[0]}
          </p>
        )}

        <button
          className={`bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full w-3/5 mt-1`}
          form="forgot-password-form"
        >
          {loadingStates[0] ? (
            <Spinner
              spinnerColor="border-white"
              spinnerSize={{ width: "w-6", height: "h-6" }}
            />
          ) : (
            "Confirm"
          )}
        </button>
      </form>

      <p className="text-secondary text-sm text-center leading-snug">
        Remembered your password?{" "}
        <Link to="/login" className="underline cursor-pointer">
          Login
        </Link>
      </p>
    </div>
  );
};

export default ForgotPasswordPage;
