import type { Route } from "./+types/chat";
import {
  getContext,
  chat,
  setAgent,
  deleteChat,
  setVisibility,
} from "~/lib/nanobot";
import Chat from "~/components/chat";
import { useRouteLoaderData } from "react-router";

export async function action({ request, params }: Route.ActionArgs) {
  if (request.method === "DELETE") {
    if (!params.id) {
      throw new Error("Chat ID is required for deletion.");
    }
    return await deleteChat(getContext(request), params.id);
  }

  const formData = await request.formData();

  const agent = formData.get("agent") as string;
  if (agent) {
    return await setAgent(getContext(request), params.id || "", agent);
  }

  const prompt = formData.get("prompt") as string;
  if (prompt) {
    await chat(getContext(request), params.id || "", prompt);
    return;
  }

  const visibility = formData.get("visibility") as string;
  if (visibility) {
    if (visibility !== "public" && visibility !== "private") {
      throw new Error(
        "Invalid visibility type. Must be 'public' or 'private'.",
      );
    }
    return await setVisibility(
      getContext(request),
      params.id || "",
      visibility,
    );
  }
}

export default function Page() {
  const data = useRouteLoaderData("routes/layout");
  return <Chat chat={data.chat} user={data.user} />;
}
