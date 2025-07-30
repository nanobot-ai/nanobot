import { FormStrategy } from "remix-auth-form";
import { Authenticator } from "remix-auth";
import { createCookieSessionStorage } from "react-router";

// Define your user type
export type User = {
  id: string;
  email: string;
  name: string;
  // ... other user properties
};

// Create a session storage
export const sessionStorage = createCookieSessionStorage({
  cookie: {
    name: "__session",
    httpOnly: true,
    path: "/",
    sameSite: "lax",
    secrets: ["s3cr3t"], // replace this with an actual secret
    secure: process.env.NODE_ENV === "production",
  },
});

// Create an instance of the authenticator, pass a generic with what
// strategies will return
export const authenticator = new Authenticator<User>();

// Your authentication logic (replace with your actual DB/API calls)
async function login(email: string, password: string): Promise<User> {
  // Verify credentials
  // Return user data or throw an error
  return {
    id: "12345", // Replace with actual user ID
    email: email,
    name: "John Doe", // Replace with actual user name
  };
}

// Tell the Authenticator to use the form strategy
authenticator.use(
  new FormStrategy(async ({ form }) => {
    const email = form.get("email") as string;
    const password = form.get("password") as string;

    if (!email || !password) {
      throw new Error("Email and password are required");
    }

    // the type of this user must match the type you pass to the
    // Authenticator the strategy will automatically inherit the type if
    // you instantiate directly inside the `use` method
    return await login(email, password);
  }),
  // each strategy has a name and can be changed to use the same strategy
  // multiple times, especially useful for the OAuth2 strategy.
  "user-pass",
);
