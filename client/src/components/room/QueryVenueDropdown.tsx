/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useState } from "react";
import { queryVenueApi } from "../../api/room";
import { api } from "../../api";
import { useDebounce } from "../../hooks/useDebounce";
import { IVenue } from "../../types/room";
import Spinner from "../Spinner";

interface SearchableVenueDropdownProps {
  value: string;
  onChange: (venue: IVenue) => void;
  errors: any;
  register: any;
}

const QueryVenueDropdown = ({
  value,
  onChange,
  errors,
  register,
}: SearchableVenueDropdownProps) => {
  const [venues, setVenues] = useState<IVenue[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState(value || "");
  const debouncedSearchTerm = useDebounce(searchTerm, 500); // 500ms debounce

  useEffect(() => {
    if (debouncedSearchTerm.length > 2 && debouncedSearchTerm !== value) {
      fetchVenues(debouncedSearchTerm);
    } else {
      setVenues([]);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedSearchTerm]);

  const fetchVenues = async (query: string) => {
    setIsLoading(true);
    try {
      const { data: response } = await queryVenueApi(api, query);
      setVenues(response.data);
    } catch (error) {
      console.error("Error fetching venues:", error);
      setVenues([]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-1 w-full">
      <label className="font-semibold text-secondary">Venue</label>
      <div className="w-full relative">
        <input
          type="text"
          id="venue"
          placeholder="Search for venue..."
          value={searchTerm}
          onChange={(e) => {
            setSearchTerm(e.target.value);
            if (e.target.value.length - searchTerm.length > 1) {
              onChange(venues.find((v) => v.name === e.target.value)!);
            }
          }}
          className="peer bg-white placeholder-gray-500 text-black px-2 py-1 rounded-lg shadow-lg w-full focus:outline-none focus:border-secondary focus:border-2"
          list="venue-suggestions"
        />
        {isLoading && (
          <div className="absolute right-2 top-2">
            <Spinner
              spinnerColor="border-gray-500"
              spinnerSize={{ width: "w-4", height: "h-4" }}
            />
          </div>
        )}
        <datalist id="venue-suggestions">
          {venues.map((venue, index) => (
            <option key={index} value={venue.name} />
          ))}
        </datalist>
      </div>
      <input
        type="hidden"
        {...register("venue", { required: "Venue is required" })}
      />
      {errors.venue && (
        <span className="ml-2 text-error text-wrap">
          {errors.venue?.message?.toString()}
        </span>
      )}
    </div>
  );
};

export default QueryVenueDropdown;
