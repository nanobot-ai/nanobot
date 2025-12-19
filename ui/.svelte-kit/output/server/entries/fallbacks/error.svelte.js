import { p as push, e as escape_html, a as pop } from "../../chunks/index.js";
import "clsx";
import { p as page } from "../../chunks/index2.js";
function Error($$payload, $$props) {
  push();
  $$payload.out.push(`<h1>${escape_html(page.status)}</h1> <p>${escape_html(page.error?.message)}</p>`);
  pop();
}
export {
  Error as default
};
