import { useForm, SubmitHandler } from "react-hook-form";
import Spinner from "../components/Spinner";
import { Link, useNavigate } from "react-router-dom";
import useLoadingAndError from "../hooks/useLoadingAndError";
import InputField from "../components/InputField";
import { signUpApi } from "../api/auth";
import { api } from "../api";

type SignUpFormData = {
	username: string;
	// phoneNumber: string;
	email: string;
	password: string;
	confirmPassword: string;
};

const SignUpPage = () => {
	const { loading, startLoading, stopLoading, error, setErrorMsg } =
		useLoadingAndError();
	const navigate = useNavigate();
	const {
		register,
		handleSubmit,
		formState: { errors },
		watch,
	} = useForm<SignUpFormData>();

	const onSubmit: SubmitHandler<SignUpFormData> = async (data) => {
		startLoading();
		try {
			console.log("[SignUpPage] Form data: ", data);
			const res = await signUpApi(
				api,
				data.username,
				// data.phoneNumber,
				data.email,
				data.password
			);
			console.log("[SignUpPage] Response: ", res);

			if (res.status !== 200) {
				switch (res.status) {
					case 400:
						setErrorMsg("Bad request. Please check request body.");
						break;
					case 409:
						setErrorMsg("Username or email already exists.");
						break;
					case 500:
					default:
						setErrorMsg("An error occurred. Please try again later.");
						break;
				}
				stopLoading();
				return;
			}

			stopLoading();
			navigate("/login");
		} catch (error) {
			console.error(error);
			setErrorMsg("An error occurred. Please try again later.");
			stopLoading();
		}
	};

	const password = watch("password");

	return (
		<div className="h-full flex flex-col justify-center items-center xs:border-y-1 border-black overflow-y-auto bg-primary py-4">
			<img src="/favicon.svg" alt="JustJio Logo" className="w-36 h-36" />

			<form
				onSubmit={handleSubmit(onSubmit)}
				id="signup-form"
				className="flex flex-col items-center gap-3 p-2 w-[70%]"
			>
				<InputField
					label="Username"
					name="username"
					type="text"
					placeholder="Enter your username"
					register={register}
					errors={errors}
					validation={{
						required: "Username is required",
						minLength: {
							value: 3,
							message: "Username must be at least 3 characters",
						},
					}}
				/>

				{/* <InputField
					label="Phone Number"
					name="phoneNumber"
					type="text"
					placeholder="Enter your phone number"
					register={register}
					errors={errors}
					validation={{
						required: "Phone number is required",
						pattern: {
							value: /^[0-9]{8}$/,
							message: "Phone number must be 8 digits",
						},
					}}
				/> */}

				<InputField
					label="Email"
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
						"Sign Up"
					)}
				</button>

				<p className="text-secondary text-sm text-center">
					Already have an account?{" "}
					<Link className="underline cursor-pointer" to="/login">
						Login
					</Link>
				</p>
			</form>
		</div>
	);
};

export default SignUpPage;
