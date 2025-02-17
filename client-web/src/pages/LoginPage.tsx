import { useForm, SubmitHandler } from "react-hook-form";
import Spinner from "../components/Spinner";
import { Link, useNavigate } from "react-router-dom";
import InputField from "../components/InputField";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { useAuth } from "../context/auth";
import { getRedirectPath } from "../utils/redirect";

type LoginFormData = {
	username: string;
	password: string;
};

const LoginPage = () => {
	const { loading, error, startLoading, stopLoading, setErrorMsg } =
		useLoadingAndError();
	const { login } = useAuth();
	const navigate = useNavigate();
	const {
		register,
		handleSubmit,
		formState: { errors },
	} = useForm<LoginFormData>();

	const onSubmit: SubmitHandler<LoginFormData> = async (data) => {
		startLoading();

		console.log("[LoginPage] Form data: ", data);
		const res = await login(data.username, data.password);
		console.log("[LoginPage] Response: ", res);

		if (!res.isSuccessResponse) {
			switch (res.error?.response?.status) {
				case 400:
					setErrorMsg("Bad request, please check request body.");
					break;
				case 401:
				case 404:
					setErrorMsg("Invalid username or password.");
					break;
				case 500:
				default:
					setErrorMsg("An error occurred, please try again later.");
					break;
			}
			stopLoading();
			return;
		}

		stopLoading();
		const redirectPath = getRedirectPath() || "/";
		navigate(redirectPath);
	};

	return (
		<div className="h-full flex flex-col justify-center items-center xs:border-y-1 border-black overflow-y-auto bg-primary">
			<img src="/favicon.svg" alt="JustJio Logo" className="w-36 h-36" />

			<form
				onSubmit={handleSubmit(onSubmit)}
				id="login-form"
				className="flex flex-col items-center gap-3 p-2 w-[70%]"
			>
				<InputField
					label="Username"
					name="username"
					type="text"
					placeholder="Enter your username"
					register={register}
					errors={errors}
					validation={{ required: "Username is required" }}
				/>

				<InputField
					label="Password"
					name="password"
					type="password"
					placeholder="Enter your password"
					register={register}
					errors={errors}
					validation={{ required: "Password is required" }}
				/>

				{error && (
					<p className="text-error text-md font-semibold text-wrap text-center">
						{error}
					</p>
				)}

				<button
					className={`bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full w-3/5 ${
						error ? "" : "mt-3"
					}`}
					form="login-form"
				>
					{loading ? (
						<Spinner
							spinnerColor="border-white"
							spinnerSize={{ width: "w-6", height: "h-6" }}
						/>
					) : (
						"Login"
					)}
				</button>

				<p className="text-secondary text-sm text-center">
					Don't have an account?{" "}
					<Link to="/signup" className="underline cursor-pointer">
						Sign Up
					</Link>
				</p>
			</form>
		</div>
	);
};

export default LoginPage;
