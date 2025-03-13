/* eslint-disable @typescript-eslint/no-explicit-any */
import { ChevronDownIcon } from "@heroicons/react/24/outline";
import React, { useState, useEffect, useRef } from "react";
import { FieldErrors, UseFormRegister } from "react-hook-form";

type Option = {
  label: string;
  value: string | number;
};

type SearchableDropdownProps = {
  options: Option[];
  onSelect: (selectedOptions: Option[]) => void;
  label?: string;
  name: string;
  register: UseFormRegister<any>;
  errors: FieldErrors<any>;
  validation?: any;
};

const SearchableDropdown: React.FC<SearchableDropdownProps> = ({
  options,
  onSelect,
  label,
  name,
  register,
  validation,
  errors,
}) => {
  const [searchTerm, setSearchTerm] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [filteredOptions, setFilteredOptions] = useState<Option[]>([]);
  const [selectedOptions, setSelectedOptions] = useState<Option[]>([]);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // filter options based on search term
  useEffect(() => {
    setFilteredOptions(
      options.filter((option) =>
        option.label.toLowerCase().includes(searchTerm.toLowerCase()),
      ),
    );
  }, [searchTerm, options]);

  // close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [dropdownRef]);

  const handleSelect = (option: Option) => {
    const isSelected = selectedOptions.some(
      (selectedOption) => selectedOption.value === option.value,
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
    <div ref={dropdownRef} className="relative w-full inline-block text-black">
      {label && (
        <label htmlFor={name} className="font-semibold text-secondary mb-1">
          {label}
        </label>
      )}

      <div
        className={`w-full cursor-pointer relative bg-white shadow-lg  ${
          label && "mt-1"
        } ${errors[name] && !isOpen && "mb-1"} ${
          isOpen ? "rounded-t-lg rounded-b-none" : "rounded-lg"
        }`}
        onClick={() => setIsOpen(!isOpen)}
      >
        <button
          type="button"
          className={`w-[87.5%] px-3 py-1 text-left text-gray-500 whitespace-nowrap overflow-x-auto`}
        >
          {selectedOptions.length > 0
            ? selectedOptions.map((opt) => opt.label).join(", ")
            : "Select options"}
        </button>
        <ChevronDownIcon className="absolute w-6 h-6 right-2 top-[0.35rem] text-secondary" />
      </div>

      {isOpen && (
        <div className="absolute z-10 w-full bg-white border border-gray-300 shadow-lg">
          <input
            type="text"
            className="w-full pl-4 py-1 bg-gray-100 text-black placeholder-black focus:outline-none"
            placeholder="Search..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
          <ul className="max-h-[7.5rem] overflow-y-auto">
            {filteredOptions.map((option) => (
              <li
                key={option.value}
                className={`pl-4 pr-2 py-1 cursor-pointer hover:bg-gray-400 ${
                  selectedOptions.some(
                    (selectedOption) => selectedOption.value === option.value,
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
        <p className="text-error text-wrap">
          {errors[name]?.message?.toString()}
        </p>
      )}
    </div>
  );
};

export default SearchableDropdown;
