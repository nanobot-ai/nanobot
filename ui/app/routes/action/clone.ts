import type { Route } from "./+types/clone";
import { getContext, clone } from "~/lib/nanobot";
import { redirect } from "react-router";

export async function action({ request, params }: Route.ActionArgs) {
  const id = await clone(getContext(request), params.id || "");
  return redirect(`/chat/${id}`);
}
