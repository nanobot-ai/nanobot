import type { Route } from "./+types/create-resource";
import { getContext, newCustomAgent } from "~/lib/nanobot";
import { redirect } from "react-router";

export async function action({ request }: Route.ActionArgs) {
  const agent = await newCustomAgent(getContext(request));
  return redirect(`/agent/${agent.id}/edit`);
}
