import { LoaderIcon } from "./icons";
import type { Attachment } from "~/lib/nanobot";

export const PreviewAttachment = ({
  attachment: { uri },
  isUploading = false,
}: {
  attachment: Attachment;
  isUploading?: boolean;
}) => {
  return (
    <div data-testid="input-attachment-preview" className="flex flex-col gap-2">
      <div className="w-20 h-16 aspect-video bg-muted rounded-md relative flex flex-col items-center justify-center">
        <img
          key={uri}
          src={uri}
          alt={"An image attachment"}
          className="rounded-md size-full object-cover"
        />

        {(isUploading || !uri) && (
          <div
            data-testid="input-attachment-loader"
            className="animate-spin absolute text-zinc-500"
          >
            <LoaderIcon />
          </div>
        )}
      </div>
    </div>
  );
};
