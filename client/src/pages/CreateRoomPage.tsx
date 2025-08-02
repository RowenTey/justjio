import { SubmitHandler, useForm } from "react-hook-form";
import InputField from "../components/InputField";
import RoomTopBar from "../components/top-bar/TopBarWithBackArrow";
import { useRoomCtx } from "../context/room";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { useNavigate } from "react-router-dom";
import Spinner from "../components/Spinner";
import SearchableDropdown from "../components/SearchableDropdown";
import { useUserCtx } from "../context/user";
import { useToast } from "../context/toast";
import { useEffect, useState } from "react";
import { IRoom, IVenue } from "../types/room";
import QueryVenueDropdown from "../components/room/QueryVenueDropdown";
import ToggleSwitch from "../components/ToggleSwitch";

enum Image {
  BIRTHDAY = "/imgs/birthday.png",
  GATHERING = "/imgs/gathering.png",
  PARTY = "/imgs/party.png",
  SPORTS = "/imgs/sports.png",
  TRAVEL = "/imgs/travel.png",
  PETS = "/imgs/pets.png",
}

type CreateRoomFormData = {
  name: string;
  date: string;
  venue: string;
  time: string;
  image: Image;
  invitees: string;
  isPrivate: boolean;
  description?: string;
};

const CreateRoomPage = () => {
  const { loadingStates, startLoading, stopLoading } = useLoadingAndError();
  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
    watch,
  } = useForm<CreateRoomFormData>({
    defaultValues: {
      name: "",
      date: "",
      venue: "",
      time: "",
      image: Image.BIRTHDAY,
      invitees: "",
      isPrivate: false,
      description: "",
    },
  });
  const { user, friends, fetchFriends } = useUserCtx();
  const { createRoom } = useRoomCtx();
  const { showToast } = useToast();
  const [selectedVenue, setSelectedVenue] = useState<IVenue | undefined>(
    undefined,
  );
  const navigate = useNavigate();

  useEffect(() => {
    fetchFriends(user.id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  const onSubmit: SubmitHandler<CreateRoomFormData> = async (data) => {
    if (!selectedVenue) {
      showToast("Please select a venue.", true);
      return;
    }

    startLoading();

    console.log("[CreateRoomPage] Submitted data: ", data);
    const roomData: Partial<IRoom> = {
      name: data.name,
      date: new Date(data.date).toISOString(),
      venue: data.venue,
      venuePlaceId: selectedVenue.googleMapsPlaceId,
      time: data.time,
      imageUrl: data.image,
      isPrivate: data.isPrivate,
      description: data.description,
    };
    const res = await createRoom(
      roomData,
      data.invitees ? data.invitees.split(",") : [],
    );

    if (!res.isSuccessResponse) {
      switch (res.error?.response?.status) {
        case 400:
          showToast("Bad request, please check request body.", true);
          break;
        case 404:
          showToast("User not found, please try again later.", true);
          break;
        case 500:
        default:
          showToast("An error occurred, please try again later.", true);
          break;
      }
      stopLoading();
      return;
    }

    stopLoading();
    showToast("Room created successfully!", false);
    navigate("/", { state: { from: "/rooms/create" } });
  };

  // Get current selected image from form
  const selectedImage = watch("image");

  return (
    <div className="h-full flex flex-col items-center bg-gray-200">
      <RoomTopBar title="Create Room" />

      <form
        onSubmit={handleSubmit(onSubmit)}
        id="create-room-form"
        className={`flex flex-col items-center p-2 pr-3 gap-2 w-[85%] h-[92%] overflow-y-auto ${
          Object.keys(errors).length > 0 && "sm:pt-[15%]"
        }`}
      >
        <div className="flex flex-col justify-center w-full">
          <div className="flex justify-between items-center pb-1">
            <label className="font-semibold text-secondary">Room Image</label>
            <ToggleSwitch
              isOn={watch("isPrivate") || false}
              onToggle={() => setValue("isPrivate", !watch("isPrivate"))}
              option1="Private"
              option2="Public"
            />
          </div>
          <div className="flex gap-2 w-full overflow-x-auto px-1 pt-1 pb-3">
            {Object.values(Image).map((image) => (
              <button
                key={image}
                type="button"
                onClick={() => setValue("image", image)}
                className={`flex-shrink-0 w-[160px] h-[110px] overflow-hidden rounded-xl ${
                  selectedImage === image ? "ring-2 ring-secondary" : ""
                }`}
              >
                <img
                  src={image}
                  className="w-full h-full object-cover"
                  loading="lazy"
                />
              </button>
            ))}
          </div>
          {errors.image && (
            <span className="ml-2 text-error text-wrap">
              {errors.image?.message?.toString()}
            </span>
          )}
        </div>

        <InputField
          name="name"
          type="text"
          label="Room Name"
          placeholder="Enter room name"
          errors={errors}
          register={register}
          validation={{ required: "Room Name is required" }}
        />

        <QueryVenueDropdown
          value={watch("venue", "")}
          onChange={(value) => {
            setSelectedVenue(value);
            setValue("venue", value.name);
          }}
          errors={errors}
          register={register}
          validation={{ required: "Venue is required" }}
        />

        <InputField
          name="date"
          type="date"
          label="Date"
          placeholder="Enter date"
          errors={errors}
          register={register}
          validation={{
            required: "Date is required",
            min: {
              value: new Date().toISOString().split("T")[0],
              message: "Date must be in the future",
            },
          }}
          min={new Date().toISOString().split("T")[0]}
          defaultValue={new Date().toISOString().split("T")[0]}
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
          validation={{
            required: "Time is required",
          }}
        />

        <SearchableDropdown
          label="Invitees"
          name="invitees"
          errors={errors}
          register={register}
          onSelect={(selected) => {
            setValue(
              "invitees",
              selected.map((option) => option.value).join(","),
            );
          }}
          options={friends.map((friend) => ({
            label: friend.username,
            value: friend.id,
          }))}
          validation={{}}
        />

        <div className="flex flex-col gap-1 w-full">
          <label className="font-semibold text-secondary">Description</label>
          <textarea
            className="w-full h-24 bg-white placeholder-gray-500 text-black px-2 py-1 rounded-lg shadow-lg focus:outline-none focus:border-secondary focus:border-2"
            placeholder="Enter room description (optional)"
            {...register("description")}
          ></textarea>
        </div>

        <button
          className="bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full mt-4 w-2/5"
          form="create-room-form"
        >
          {loadingStates[0] ? (
            <Spinner
              spinnerColor="border-white"
              spinnerSize={{ width: "w-6", height: "h-6" }}
            />
          ) : (
            "Submit"
          )}
        </button>
      </form>
    </div>
  );
};

export default CreateRoomPage;
