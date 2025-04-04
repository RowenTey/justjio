import { SubmitHandler, useForm } from "react-hook-form";
import Spinner from "../components/Spinner";
import TopBarWithBackArrow from "../components/top-bar/TopBarWithBackArrow";
import { AxiosError } from "axios";
import useLoadingAndError from "../hooks/useLoadingAndError";
import InputField from "../components/InputField";
import { useUserCtx } from "../context/user";
import { updateUserApi } from "../api/user";
import { api } from "../api";
import { useNavigate } from "react-router-dom";
import { useToast } from "../context/toast";

const EditProfilePage = () => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<{ username: string }>();
  const { loadingStates, startLoading, stopLoading, errorStates, setErrorMsg } =
    useLoadingAndError();
  const { user, setUser } = useUserCtx();
  const { showToast } = useToast();
  const navigate = useNavigate();

  const onSubmit: SubmitHandler<{ username: string }> = async (data) => {
    startLoading();
    try {
      await updateUserApi(api, user.id, "username", data.username);
      setUser({ ...user, username: data.username });
      showToast(`Updated username successfully!`, false);
      stopLoading();
      setTimeout(() => navigate(-1), 1000);
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
    <div className="h-full flex flex-col items-center gap-4 bg-gray-200">
      <TopBarWithBackArrow title="Edit Profile" />

      <form
        onSubmit={handleSubmit(onSubmit)}
        id="edit-profile-form"
        className="h-full flex flex-col justify-center items-center gap-3 p-2 w-[70%] mt-2"
      >
        <InputField
          label="Username"
          name="username"
          type="text"
          placeholder="Enter your new username"
          register={register}
          errors={errors}
          validation={{
            required: "Username is required",
            minLength: {
              value: 3,
              message: "Username must be at least 3 characters",
            },
            validate: (value: string) => {
              return value === user.username
                ? "Username must be different from current username"
                : true;
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
          form="edit-profile-form"
        >
          {loadingStates[0] ? (
            <Spinner
              spinnerColor="border-white"
              spinnerSize={{ width: "w-6", height: "h-6" }}
            />
          ) : (
            "Update"
          )}
        </button>
      </form>
    </div>
  );
};

export default EditProfilePage;
