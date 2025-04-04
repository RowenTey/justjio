// import { render, screen, fireEvent, waitFor } from "@testing-library/react";
// import { BrowserRouter } from "react-router-dom";
// import LoginPage from "../pages/LoginPage";
// import { AuthContext } from "../context/auth";
// import { LoadingAndErrorContext } from "../hooks/useLoadingAndError";
// import jest from "jest";

// // Mock the Spinner component
// jest.mock("../components/Spinner", () => () => <div>Loading...</div>);

// // Mock InputField component
// jest.mock(
// 	"../components/InputField",
// 	() =>
// 		({ label, name, type, placeholder, register, errors, validation }: any) =>
// 			(
// 				<div>
// 					<label htmlFor={name}>{label}</label>
// 					<input
// 						id={name}
// 						type={type}
// 						placeholder={placeholder}
// 						{...register(name, validation)}
// 					/>
// 					{errors[name] && <span>{errors[name].message}</span>}
// 				</div>
// 			)
// );

// // Mock useAuth hook
// jest.mock("../context/auth", () => ({
// 	useAuth: () => ({
// 		login: jest.fn().mockResolvedValue({ isSuccessResponse: true }),
// 	}),
// }));

// // Mock useLoadingAndError hook
// jest.mock("../hooks/useLoadingAndError", () => ({
// 	default: () => ({
// 		loading: false,
// 		error: null,
// 		startLoading: jest.fn(),
// 		stopLoading: jest.fn(),
// 		setErrorMsg: jest.fn(),
// 	}),
// }));

// describe("LoginPage", () => {
// 	test("renders LoginPage and submits form", async () => {
// 		const { login } = require("../context/auth").useAuth();
// 		const { startLoading, stopLoading, setErrorMsg } =
// 			require("../hooks/useLoadingAndError").default();

// 		render(
// 			<AuthContext.Provider value={{ login }}>
// 				<LoadingAndErrorContext.Provider
// 					value={{ startLoading, stopLoading, setErrorMsg }}
// 				>
// 					<BrowserRouter>
// 						<LoginPage />
// 					</BrowserRouter>
// 				</LoadingAndErrorContext.Provider>
// 			</AuthContext.Provider>
// 		);

// 		// Verify that the component renders correctly
// 		expect(screen.getByText(/Login/i)).toBeInTheDocument();
// 		expect(screen.getByLabelText(/Username/i)).toBeInTheDocument();
// 		expect(screen.getByLabelText(/Password/i)).toBeInTheDocument();

// 		// Simulate user input
// 		fireEvent.change(screen.getByLabelText(/Username/i), {
// 			target: { value: "testuser" },
// 		});
// 		fireEvent.change(screen.getByLabelText(/Password/i), {
// 			target: { value: "password123" },
// 		});

// 		// Simulate form submission
// 		fireEvent.click(screen.getByText(/Submit/i));

// 		// Wait for the async operations
// 		await waitFor(() => expect(startLoading).toHaveBeenCalled());
// 		await waitFor(() =>
// 			expect(login).toHaveBeenCalledWith("testuser", "password123")
// 		);
// 		await waitFor(() => expect(stopLoading).toHaveBeenCalled());
// 	});

// 	test("shows error message on failed login", async () => {
// 		const { login } = require("../context/auth").useAuth();
// 		login.mockResolvedValue({
// 			isSuccessResponse: false,
// 			error: "Invalid credentials",
// 		});

// 		const { startLoading, stopLoading, setErrorMsg } =
// 			require("../hooks/useLoadingAndError").default();

// 		render(
// 			<AuthContext.Provider value={{ login }}>
// 				<LoadingAndErrorContext.Provider
// 					value={{ startLoading, stopLoading, setErrorMsg }}
// 				>
// 					<BrowserRouter>
// 						<LoginPage />
// 					</BrowserRouter>
// 				</LoadingAndErrorContext.Provider>
// 			</AuthContext.Provider>
// 		);

// 		// Simulate user input
// 		fireEvent.change(screen.getByLabelText(/Username/i), {
// 			target: { value: "testuser" },
// 		});
// 		fireEvent.change(screen.getByLabelText(/Password/i), {
// 			target: { value: "password123" },
// 		});

// 		// Simulate form submission
// 		fireEvent.click(screen.getByText(/Submit/i));

// 		// Wait for the async operations
// 		await waitFor(() =>
// 			expect(setErrorMsg).toHaveBeenCalledWith(
// 				"An error occurred. Please try again later."
// 			)
// 		);
// 	});
// });
