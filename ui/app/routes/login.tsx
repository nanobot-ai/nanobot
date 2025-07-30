import { Form, redirect } from "react-router";
import { authenticator, sessionStorage } from "~/services/auth.server";

// Import this from correct place for your route
import type { Route } from "./+types/login";

// Second, we need to export an action function, here we will use the
// `authenticator.authenticate` method
export async function action({ request }: Route.ActionArgs) {
  try {
    // we call the method with the name of the strategy we want to use and the
    // request object
    const user = await authenticator.authenticate("user-pass", request);

    const session = await sessionStorage.getSession(
      request.headers.get("cookie"),
    );

    session.set("user", user);

    // Redirect to the home page after successful login
    redirect("/", {
      headers: {
        "Set-Cookie": await sessionStorage.commitSession(session),
      },
    });
    return {};
  } catch (error) {
    // Return validation errors or authentication errors
    if (error instanceof Error) {
      return { error: error.message };
    }

    // Re-throw any other errors (including redirects)
    throw error;
  }
}

// Finally, we need to export a loader function to check if the user is already
// authenticated and redirect them to the dashboard
export async function loader({ request }: Route.LoaderArgs) {
  const session = await sessionStorage.getSession(
    request.headers.get("cookie"),
  );
  const user = session.get("user");

  // If the user is already authenticated redirect to the dashboard
  if (user) return redirect("/");

  // Otherwise return null to render the login page
  return {};
}

// First we create our UI with the form doing a POST and the inputs with
// the names we are going to use in the strategy
export default function Login({ actionData }: Route.ComponentProps) {
  return (
    <div>
      <h1>Login</h1>

      {actionData?.error ? (
        <div className="error">{actionData.error}</div>
      ) : null}

      <Form method="post" target="/login">
        <div>
          <label htmlFor="email">Email</label>
          <input type="email" name="email" id="email" required />
        </div>

        <div>
          <label htmlFor="password">Password</label>
          <input
            type="password"
            name="password"
            id="password"
            autoComplete="current-password"
            required
          />
        </div>

        <button type="submit">Sign In</button>
      </Form>
    </div>
  );
}
