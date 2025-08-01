import { SubmitHandler, useForm } from "react-hook-form";
import Spinner from "../components/Spinner";
import TopBarWithBackArrow from "../components/top-bar/TopBarWithBackArrow";
import { AxiosError } from "axios";
import useLoadingAndError from "../hooks/useLoadingAndError";
import InputField from "../components/InputField";
import { api } from "../api";
import { useLocation, useNavigate } from "react-router-dom";
import { useToast } from "../context/toast";
import { useState } from "react";
import { IVenue } from "../types/room";
import QueryVenueDropdown from "../components/room/QueryVenueDropdown";
import { updateRoomApi, UpdateRoomRequest } from "../api/room";

type EditRoomformData = {
  date: string;
  time: string;
  venue: string;
  description: string;
};

const EditRoomPage = () => {
  const navigate = useNavigate();
  const { state } = useLocation();
  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<EditRoomformData>({
    defaultValues: {
      venue: state.room.venue,
      description: state.room.description,
    },
  });
  const { loadingStates, startLoading, stopLoading, errorStates, setErrorMsg } =
    useLoadingAndError();
  const { showToast } = useToast();
  const [selectedVenue, setSelectedVenue] = useState<IVenue>({
    name: state.room.venue,
    googleMapsPlaceId: state.room.venuePlaceId,
    address: "",
  });

  const onSubmit: SubmitHandler<EditRoomformData> = async (data) => {
    startLoading();
    try {
      const payload: UpdateRoomRequest = {
        venue: data.venue,
        placeId: selectedVenue.googleMapsPlaceId,
        date: new Date(data.date).toISOString(),
        time: data.time,
        description: data.description,
      };
      await updateRoomApi(api, state.room.id, payload);
      showToast(`Updated room successfully!`, false);
      setTimeout(() => navigate(-1), 1000);
    } catch (error) {
      console.error(error);
      switch ((error as AxiosError).response?.status) {
        case 400:
          setErrorMsg("Bad request. Please check request body.");
          break;
        case 404:
          setErrorMsg("Room not found.");
          break;
        case 500:
        default:
          setErrorMsg("An error occurred. Please try again later.");
          break;
      }
    } finally {
      stopLoading();
    }
  };

  if (!state.room) {
    navigate(-1);
    return null;
  }

  return (
    <div className="h-full flex flex-col items-center gap-4 bg-gray-200">
      <TopBarWithBackArrow title="Edit Room" />

      <form
        onSubmit={handleSubmit(onSubmit)}
        id="edit-profile-form"
        className="h-full flex flex-col justify-center items-center gap-3 p-2 w-[70%] mt-2"
      >
        <QueryVenueDropdown
          value={watch("venue") || ""}
          onChange={(value) => {
            setSelectedVenue(value);
            setValue("venue", value.name);
          }}
          errors={errors}
          register={register}
        />

        <InputField
          name="date"
          type="date"
          label="Date"
          placeholder="Enter date"
          errors={errors}
          register={register}
          validation={{
            min: {
              value: new Date().toISOString().split("T")[0],
              message: "Date must be in the future",
            },
          }}
          min={new Date().toISOString().split("T")[0]}
          defaultValue={state.room.date.split("T")[0]}
        />

        <InputField
          name="time"
          type="time"
          label="Time"
          placeholder="Enter time"
          min="00:00"
          max="23:59"
          errors={errors}
          register={register}
          defaultValue={state.room.time}
        />

        <div className="flex flex-col gap-1 w-full">
          <label className="font-semibold text-secondary">
            Description
            <textarea
              className="w-full h-24 bg-white placeholder-gray-500 text-black px-2 py-1 rounded-lg shadow-lg focus:outline-none focus:border-secondary focus:border-2"
              defaultValue={state.room.description}
              placeholder="Enter room description (optional)"
              {...register("description")}
            ></textarea>
          </label>
        </div>

        {errorStates[0] && (
          <p className="text-error text-md font-semibold text-wrap text-center leading-tight">
            {errorStates[0]}
          </p>
        )}

        <button
          className={`bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full w-3/5 mt-1`}
          form="edit-profile-form"
        >
          {loadingStates[0] ? (
            <Spinner
              spinnerColor="border-white"
              spinnerSize={{ width: "w-6", height: "h-6" }}
            />
          ) : (
            "Update"
          )}
        </button>
      </form>
    </div>
  );
};

export default EditRoomPage;
