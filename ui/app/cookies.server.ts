import { createCookie } from "react-router";

export const sidebarCookie = createCookie("sidebar:state", {
  maxAge: 604_800, // one week
});
