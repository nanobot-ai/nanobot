import type { Route } from "./+types/create-resource";
import { createResource, getContext } from "~/lib/nanobot";

export async function action({ request, params }: Route.ActionArgs) {
  const formData = await request.formData();
  const file = formData.get("file");
  const mimeType = formData.get("mimeType");
  if (!(file instanceof File)) {
    return { error: "Invalid file upload." };
  }
  if (!mimeType || typeof mimeType !== "string") {
    return { error: "Invalid or missing mimeType." };
  }

  try {
    const buffer = await file.arrayBuffer();
    const base64String = Buffer.from(buffer).toString("base64");
    return createResource(getContext(request), params.id, base64String, {
      mimeType,
    });
  } catch (error) {
    return { error: `Failed to read file. ${error}` };
  }
}
