import { useForm } from "react-hook-form";
import ModalWrapper, { ModalWrapperProps } from "../ModalWrapper";
import SearchableDropdown from "../SearchableDropdown";
import { IUser } from "../../types/user";
import { useEffect, useState } from "react";
import {
  getUninvitedFriendsForRoomApi,
  inviteUsersToRoomApi,
} from "../../api/room";
import { api } from "../../api";
import { useToast } from "../../context/toast";
import { AxiosError } from "axios";
import { QrCodeIcon } from "@heroicons/react/24/solid";
import { LinkIcon } from "@heroicons/react/24/outline";
import QRCodeModal from "./QRCodeModal";

interface InviteAttendeesFormData {
  invitees: string;
}

interface InviteAttendeesModalProps {
  roomId: string;
  url: string;
}

const InviteAttendeesModalContent: React.FC<
  InviteAttendeesModalProps & ModalWrapperProps
> = ({ roomId }) => {
  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<InviteAttendeesFormData>();
  const [uninvitedFriends, SetUninvitedFriends] = useState<IUser[]>([]);
  const [isQRCodeModalVisible, setIsQRCodeModalVisible] = useState(false);
  const { showToast } = useToast();

  useEffect(() => {
    const fetchUninvitedFriends = async (roomId: string) => {
      return await getUninvitedFriendsForRoomApi(api, roomId);
    };

    fetchUninvitedFriends(roomId).then((res) => {
      SetUninvitedFriends(res.data.data);
    });
  }, [roomId]);

  const handleInviteUsers = async (data: InviteAttendeesFormData) => {
    const invitees = data.invitees.split(",");
    try {
      await inviteUsersToRoomApi(api, roomId, invitees);

      showToast("Users invited successfully!", false);
      SetUninvitedFriends((prev) =>
        prev.filter((friend) => !invitees.includes(friend.id.toString())),
      );
    } catch (error) {
      console.error("Error inviting users to room", error);
      switch ((error as AxiosError).response?.status) {
        case 400:
          showToast("Bad request, please check request body.", true);
          break;
        case 401:
          showToast("Only hosts can invite other users", true);
          break;
        case 404:
          showToast("User not found, please try again later.", true);
          break;
        case 500:
        default:
          showToast("An error occurred, please try again later.", true);
          break;
      }
    }
  };

  return (
    <>
      <div className="w-full flex flex-col gap-3">
        <div className="w-full flex flex-col items-center gap-2">
          <h2 className="text-3xl font-bold text-secondary">Invite Friend</h2>
          <form
            onSubmit={handleSubmit(handleInviteUsers)}
            className="w-full flex flex-col items-center justify-center gap-3"
            id="invite-people-form"
          >
            <SearchableDropdown
              name="invitees"
              errors={errors}
              register={register}
              onSelect={(selected) => {
                setValue(
                  "invitees",
                  selected.map((option) => option.value).join(","),
                );
              }}
              options={uninvitedFriends.map((friend) => ({
                label: friend.username,
                value: friend.id,
              }))}
              validation={{ required: "Invitees are required" }}
            />

            <button
              className="bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full mt-2 w-2/5"
              form="invite-people-form"
            >
              Submit
            </button>
          </form>
        </div>
        <hr className="text-black" />
        <div className="w-full flex justify-center items-center gap-6 hover:cursor-pointer">
          <div
            className="w-14 flex flex-col items-center hover:cursor-pointer"
            onClick={() => setIsQRCodeModalVisible(true)}
          >
            <QrCodeIcon className="h-12 w-12 p-2 text-secondary bg-white rounded-full hover:scale-105" />
            <p className="text-secondary text-center text-sm leading-4">
              Generate QR
            </p>
          </div>
          <div className="w-14 flex flex-col items-center hover:cursor-pointer">
            <LinkIcon className="h-12 w-12 p-2 text-secondary bg-white rounded-full hover:scale-105" />
            <p className="text-secondary text-center text-sm leading-4">
              Copy Link
            </p>
          </div>
        </div>
      </div>

      <QRCodeModal
        url={window.location.href + "?join=true"}
        isVisible={isQRCodeModalVisible}
        closeModal={() => setIsQRCodeModalVisible(false)}
      />
    </>
  );
};

const InviteAttendeesModal = ModalWrapper(InviteAttendeesModalContent);

export default InviteAttendeesModal;
