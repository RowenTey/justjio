import { SubmitHandler, useForm } from "react-hook-form";
import useLoadingAndError from "../hooks/useLoadingAndError";
import InputField from "../components/InputField";
import Spinner from "../components/Spinner";
import { useLocation, useNavigate } from "react-router-dom";
import { resetPasswordApi } from "../api/auth";
import { api } from "../api";
import { useToast } from "../context/toast";
import { AxiosError } from "axios";

const ResetPasswordPage = () => {
  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<{ password: string; confirmPassword: string }>();
  const password = watch("password");
  const { loadingStates, startLoading, stopLoading, errorStates, setErrorMsg } =
    useLoadingAndError();
  const { state } = useLocation();
  const email = (state?.email as string) || "";
  const { showToast } = useToast();
  const navigate = useNavigate();

  const onSubmit: SubmitHandler<{
    password: string;
    confirmPassword: string;
  }> = async (data) => {
    startLoading();
    try {
      await resetPasswordApi(api, email, data.password);
      showToast("Password reset successfully!", false);
      stopLoading();
      setTimeout(
        () =>
          navigate("/login", {
            state: {
              from: "/resetPassword",
            },
          }),
        1000,
      );
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
      stopLoading();
    }
  };

  return (
    <div className="h-full flex flex-col justify-center items-center xs:border-y-1 border-black overflow-y-auto bg-primary py-4">
      <img src="/favicon.svg" alt="JustJio Logo" className="w-32 h-32" />

      <h1 className="text-4xl font-bold text-secondary mt-3">Reset Password</h1>

      <form
        onSubmit={handleSubmit(onSubmit)}
        id="reset-password-form"
        className="flex flex-col items-center gap-3 p-2 w-[70%] mt-2"
      >
        <InputField
          label="Password"
          name="password"
          type="password"
          placeholder="Enter your password"
          register={register}
          errors={errors}
          validation={{
            required: "Password is required",
            minLength: {
              value: 6,
              message: "Password must be at least 6 characters",
            },
          }}
        />

        <InputField
          label="Confirm Password"
          name="confirmPassword"
          type="password"
          placeholder="Confirm your password"
          register={register}
          errors={errors}
          validation={{
            required: "Confirm Password is required",
            validate: (value: string) =>
              value === password || "Passwords do not match",
          }}
        />

        {errorStates[0] && (
          <p className="text-error text-md font-semibold text-wrap text-center leading-tight">
            {errorStates[0]}
          </p>
        )}

        <button
          className={`bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full w-3/5 mt-1`}
          form="reset-password-form"
        >
          {loadingStates[0] ? (
            <Spinner
              spinnerColor="border-white"
              spinnerSize={{ width: "w-6", height: "h-6" }}
            />
          ) : (
            "Reset Password"
          )}
        </button>
      </form>
    </div>
  );
};

export default ResetPasswordPage;
