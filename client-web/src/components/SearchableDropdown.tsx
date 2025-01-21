/* eslint-disable no-mixed-spaces-and-tabs */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { ChevronDownIcon } from "@heroicons/react/24/outline";
import React, { useState, useEffect } from "react";
import { FieldErrors, UseFormRegister } from "react-hook-form";

type Option = {
	label: string;
	value: string;
};

type SearchableDropdownProps = {
	options: Option[];
	onSelect: (selectedOptions: Option[]) => void;
	name: string;
	register: UseFormRegister<any>;
	errors: FieldErrors<any>;
	validation?: any;
};

const SearchableDropdown: React.FC<SearchableDropdownProps> = ({
	options,
	onSelect,
	name,
	register,
	validation,
	errors,
}) => {
	const [searchTerm, setSearchTerm] = useState("");
	const [isOpen, setIsOpen] = useState(false);
	const [filteredOptions, setFilteredOptions] = useState<Option[]>([]);
	const [selectedOptions, setSelectedOptions] = useState<Option[]>([]);

	// filter options based on search term
	useEffect(() => {
		setFilteredOptions(
			options.filter((option) =>
				option.label.toLowerCase().includes(searchTerm.toLowerCase())
			)
		);
	}, [searchTerm, options]);

	const handleSelect = (option: Option) => {
		const isSelected = selectedOptions.some(
			(selectedOption) => selectedOption.value === option.value
		);

		let newSelectedOptions: Option[] = [];
		newSelectedOptions = isSelected
			? selectedOptions.filter((selectedOption) => {
					return selectedOption.value !== option.value;
			  })
			: [...selectedOptions, option];

		setSelectedOptions(newSelectedOptions);
		onSelect(newSelectedOptions);
		setSearchTerm("");
	};

	return (
		<div className="relative inline-block w-full mt-4 text-black">
			<div className="w-full cursor-pointer" onClick={() => setIsOpen(!isOpen)}>
				<button
					type="button"
					className="w-full px-3 py-2 text-left bg-white border border-gray-300 rounded-md shadow-sm overflow-x-auto"
				>
					Select options
				</button>
				<ChevronDownIcon className="absolute w-6 h-6 text-black right-3 top-3" />
			</div>

			{isOpen && (
				<div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg">
					<input
						type="text"
						className="w-full pl-4 py-2 bg-slate-400 text-black placeholder-black focus:border-1 focus:border-black focus:outline-none"
						placeholder="Search..."
						value={searchTerm}
						onChange={(e) => setSearchTerm(e.target.value)}
					/>
					<ul className="max-h-[7.5rem] overflow-y-auto">
						{filteredOptions.map((option) => (
							<li
								key={option.value}
								className={`px-4 py-2 cursor-pointer hover:bg-gray-300 ${
									selectedOptions.some(
										(selectedOption) => selectedOption.value === option.value
									)
										? "bg-gray-200"
										: ""
								}`}
								onClick={() => handleSelect(option)}
							>
								{option.label}
							</li>
						))}
					</ul>
				</div>
			)}
			<input type="hidden" {...register(name, validation)} />
			{errors[name] && (
				<p className="mt-2 ml-2 text-sm text-red-600">
					{errors[name]?.message?.toString()}
				</p>
			)}
		</div>
	);
};

export default SearchableDropdown;
