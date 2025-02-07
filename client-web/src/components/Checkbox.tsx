import { CheckIcon } from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";

interface CheckboxProps {
	label: string;
	isChecked: boolean;
	onChange: (checked: boolean) => void;
}

const Checkbox: React.FC<CheckboxProps> = ({ label, isChecked, onChange }) => {
	const [checked, setChecked] = useState(isChecked);

	useEffect(() => {
		setChecked(isChecked);
	}, [isChecked]);

	const handleChange = () => {
		setChecked(!checked);
		onChange(!checked);
	};

	return (
		<label className="flex items-center cursor-pointer">
			<input
				type="checkbox"
				className="hidden"
				checked={checked}
				onChange={handleChange}
			/>

			<span
				className={`w-5 h-5 border-2 rounded-lg flex items-center justify-center mr-2 transition-colors ${
					checked
						? "bg-secondary border-secondary"
						: "border-secondary hover:bg-gray-300"
				}`}
			>
				{checked && <CheckIcon className="w-4 h-4 text-white" />}
			</span>
			<span className="text-md text-secondary">{label}</span>
		</label>
	);
};

export default Checkbox;
