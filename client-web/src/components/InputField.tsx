/* eslint-disable @typescript-eslint/no-explicit-any */
import React from "react";
import { UseFormRegister, FieldErrors } from "react-hook-form";

interface InputFieldProps {
	label: string;
	name: string;
	type: string;
	isFloat?: boolean;
	placeholder: string;
	register: UseFormRegister<any>;
	errors: FieldErrors<any>;
	validation?: any;
}

const InputField: React.FC<InputFieldProps> = ({
	label,
	name,
	type,
	isFloat,
	placeholder,
	register,
	errors,
	validation,
}) => {
	return (
		<div className="flex flex-col gap-1 w-full">
			<label htmlFor={name} className="font-semibold text-justjio-secondary">
				{label}
			</label>
			<input
				type={type}
				step={isFloat ? "0.01" : "1"}
				id={name}
				placeholder={placeholder}
				className="bg-white text-black px-2 py-1 rounded-lg shadow-lg"
				{...register(name, validation)}
			/>
			{errors[name] && (
				<span className="text-red-600 text-wrap">
					{errors[name]?.message?.toString()}
				</span>
			)}
		</div>
	);
};

export default InputField;
