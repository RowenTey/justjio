/* eslint-disable @typescript-eslint/no-explicit-any */
import { CalendarDaysIcon, ClockIcon } from "@heroicons/react/24/outline";
import React from "react";
import { UseFormRegister, FieldErrors } from "react-hook-form";

interface InputFieldProps {
	label?: string;
	name: string;
	type: string;
	placeholder: string;
	register: UseFormRegister<any>;
	errors: FieldErrors<any>;
	isFloat?: boolean;
	min?: string;
	max?: string;
	pattern?: string;
	defaultValue?: string;
	validation?: any;
}

const InputField: React.FC<InputFieldProps> = ({
	label,
	name,
	type,
	placeholder,
	register,
	errors,
	isFloat,
	min,
	max,
	pattern,
	defaultValue,
	validation,
}) => {
	return (
		<div className="flex flex-col gap-1 w-full">
			{label && (
				<label htmlFor={name} className="font-semibold text-secondary">
					{label}
				</label>
			)}
			<div className="w-full relative">
				<input
					type={type}
					step={isFloat ? "0.01" : undefined}
					id={name}
					placeholder={placeholder}
					min={min}
					max={max}
					pattern={pattern}
					defaultValue={defaultValue}
					className="peer bg-white placeholder-gray-500 text-black px-2 py-1 rounded-lg shadow-lg w-full focus:outline-none focus:border-secondary focus:border-2"
					{...register(name, validation)}
				/>
				{type === "date" && (
					<CalendarDaysIcon className="absolute right-3 pointer-events-none h-6 w-6 text-secondary peer-focus:top-[0.35rem] top-1" />
				)}
				{type === "time" && (
					<ClockIcon className="absolute right-3 pointer-events-none h-6 w-6 text-secondary peer-focus:top-[0.35rem] top-1" />
				)}
			</div>
			{errors[name] && (
				<span className="text-error text-wrap">
					{errors[name]?.message?.toString()}
				</span>
			)}
		</div>
	);
};

export default InputField;
