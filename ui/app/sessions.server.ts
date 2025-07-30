import { createThemeSessionResolver } from "remix-themes";
import { createCookieSessionStorage } from "react-router";

export const themeSessionResolver = createThemeSessionResolver(
  createCookieSessionStorage({
    cookie: {
      name: "__remix-themes",
      path: "/",
      httpOnly: true,
      sameSite: "lax",
      secrets: [process.env.COOKIE_SECRET || ""],
      secure: process.env.NODE_ENV === "production",
    },
  }),
);
