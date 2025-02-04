import { useForm, SubmitHandler } from "react-hook-form";
import Spinner from "../components/Spinner";
import { Link, useNavigate } from "react-router-dom";
import InputField from "../components/InputField";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { useAuth } from "../context/auth";

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

		console.log(data);
		const res = await login(data.username, data.password);
		console.log(res);

		if (!res.isSuccessResponse) {
			console.error(res.error);
			setErrorMsg("An error occurred. Please try again later.");
			stopLoading();
			return;
		}

		stopLoading();
		navigate("/");
	};

	if (loading) {
		return <Spinner bgClass="bg-justjio-primary" />;
	}

	return (
		<div className="h-full flex flex-col justify-center items-center xs:border-y-1 border-black overflow-y-auto bg-justjio-primary">
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

				<button
					className="bg-justjio-secondary hover:bg-purple-900 text-white font-bold py-2 px-4 rounded-full mt-3 w-3/5"
					form="login-form"
				>
					Login
				</button>

				<p className="text-justjio-secondary text-sm text-center">
					Don't have an account?{" "}
					<Link to="/signup" className="underline cursor-pointer">
						Sign Up
					</Link>
				</p>

				{error && <p className="text-red-500 text-wrap text-center">{error}</p>}
			</form>
		</div>
	);
};

export default LoginPage;
