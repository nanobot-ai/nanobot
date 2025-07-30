import type { Route } from "./+types/create-resource";
import { deleteCustomAgent, getContext } from "~/lib/nanobot";
import { redirect } from "react-router";

export async function action({ request, params }: Route.ActionArgs) {
  const formData = await request.formData();
  const agentId = formData.get("agentId") as string;

  if (request.method !== "DELETE") {
    throw new Error("Invalid request method. Use DELETE to remove an agent.");
  }

  if (!agentId) {
    throw new Error("Agent ID is required");
  }

  // Use "new" as the session ID since we're not in a specific chat context
  await deleteCustomAgent(getContext(request), params.id || "new", agentId);

  return redirect("/");
}
