import { W as current_component, p as push, m as spread_props, a as pop, U as attr_class, k as clsx, c as attr, d as stringify, e as escape_html, b as ensure_array_like, X as bind_props, Y as maybe_selected } from "./index.js";
import "clsx";
import { Marked } from "marked";
import { markedHighlight } from "marked-highlight";
import hljs from "highlight.js";
import { I as Icon, X, T as Triangle_alert, a as Copy } from "./chat.svelte.js";
import "@mcp-ui/client/ui-resource-renderer.wc.js";
import React from "react";
import "react-dom/client";
import { isUIResource } from "@mcp-ui/client";
function html(value) {
  var html2 = String(value ?? "");
  var open = "<!---->";
  return open + html2 + "<!---->";
}
function onDestroy(fn) {
  var context = (
    /** @type {Component} */
    current_component
  );
  (context.d ??= []).push(fn);
}
function External_link($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    ["path", { "d": "M15 3h6v6" }],
    ["path", { "d": "M10 14 21 3" }],
    [
      "path",
      {
        "d": "M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
      }
    ]
  ];
  Icon($$payload, spread_props([
    { name: "external-link" },
    /**
     * @component @name ExternalLink
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cGF0aCBkPSJNMTUgM2g2djYiIC8+CiAgPHBhdGggZD0iTTEwIDE0IDIxIDMiIC8+CiAgPHBhdGggZD0iTTE4IDEzdjZhMiAyIDAgMCAxLTIgMkg1YTIgMiAwIDAgMS0yLTJWOGEyIDIgMCAwIDEgMi0yaDYiIC8+Cjwvc3ZnPgo=) - https://lucide.dev/icons/external-link
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
function Library($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    ["path", { "d": "m16 6 4 14" }],
    ["path", { "d": "M12 6v14" }],
    ["path", { "d": "M8 8v12" }],
    ["path", { "d": "M4 4v16" }]
  ];
  Icon($$payload, spread_props([
    { name: "library" },
    /**
     * @component @name Library
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cGF0aCBkPSJtMTYgNiA0IDE0IiAvPgogIDxwYXRoIGQ9Ik0xMiA2djE0IiAvPgogIDxwYXRoIGQ9Ik04IDh2MTIiIC8+CiAgPHBhdGggZD0iTTQgNHYxNiIgLz4KPC9zdmc+Cg==) - https://lucide.dev/icons/library
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
function Lightbulb($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    [
      "path",
      {
        "d": "M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.7 1.3 1.5 1.5 2.5"
      }
    ],
    ["path", { "d": "M9 18h6" }],
    ["path", { "d": "M10 22h4" }]
  ];
  Icon($$payload, spread_props([
    { name: "lightbulb" },
    /**
     * @component @name Lightbulb
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cGF0aCBkPSJNMTUgMTRjLjItMSAuNy0xLjcgMS41LTIuNSAxLS45IDEuNS0yLjIgMS41LTMuNUE2IDYgMCAwIDAgNiA4YzAgMSAuMiAyLjIgMS41IDMuNS43LjcgMS4zIDEuNSAxLjUgMi41IiAvPgogIDxwYXRoIGQ9Ik05IDE4aDYiIC8+CiAgPHBhdGggZD0iTTEwIDIyaDQiIC8+Cjwvc3ZnPgo=) - https://lucide.dev/icons/lightbulb
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
function Paperclip($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    [
      "path",
      {
        "d": "m16 6-8.414 8.586a2 2 0 0 0 2.829 2.829l8.414-8.586a4 4 0 1 0-5.657-5.657l-8.379 8.551a6 6 0 1 0 8.485 8.485l8.379-8.551"
      }
    ]
  ];
  Icon($$payload, spread_props([
    { name: "paperclip" },
    /**
     * @component @name Paperclip
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cGF0aCBkPSJtMTYgNi04LjQxNCA4LjU4NmEyIDIgMCAwIDAgMi44MjkgMi44MjlsOC40MTQtOC41ODZhNCA0IDAgMSAwLTUuNjU3LTUuNjU3bC04LjM3OSA4LjU1MWE2IDYgMCAxIDAgOC40ODUgOC40ODVsOC4zNzktOC41NTEiIC8+Cjwvc3ZnPgo=) - https://lucide.dev/icons/paperclip
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
function Send($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    [
      "path",
      {
        "d": "M14.536 21.686a.5.5 0 0 0 .937-.024l6.5-19a.496.496 0 0 0-.635-.635l-19 6.5a.5.5 0 0 0-.024.937l7.93 3.18a2 2 0 0 1 1.112 1.11z"
      }
    ],
    ["path", { "d": "m21.854 2.147-10.94 10.939" }]
  ];
  Icon($$payload, spread_props([
    { name: "send" },
    /**
     * @component @name Send
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cGF0aCBkPSJNMTQuNTM2IDIxLjY4NmEuNS41IDAgMCAwIC45MzctLjAyNGw2LjUtMTlhLjQ5Ni40OTYgMCAwIDAtLjYzNS0uNjM1bC0xOSA2LjVhLjUuNSAwIDAgMC0uMDI0LjkzN2w3LjkzIDMuMThhMiAyIDAgMCAxIDEuMTEyIDEuMTF6IiAvPgogIDxwYXRoIGQ9Im0yMS44NTQgMi4xNDctMTAuOTQgMTAuOTM5IiAvPgo8L3N2Zz4K) - https://lucide.dev/icons/send
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
function Settings($$payload, $$props) {
  push();
  /**
   * @license @lucide/svelte v0.540.0 - ISC
   *
   * ISC License
   *
   * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2023 as part of Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2025.
   *
   * Permission to use, copy, modify, and/or distribute this software for any
   * purpose with or without fee is hereby granted, provided that the above
   * copyright notice and this permission notice appear in all copies.
   *
   * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
   *
   * ---
   *
   * The MIT License (MIT) (for portions derived from Feather)
   *
   * Copyright (c) 2013-2023 Cole Bemis
   *
   * Permission is hereby granted, free of charge, to any person obtaining a copy
   * of this software and associated documentation files (the "Software"), to deal
   * in the Software without restriction, including without limitation the rights
   * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   * copies of the Software, and to permit persons to whom the Software is
   * furnished to do so, subject to the following conditions:
   *
   * The above copyright notice and this permission notice shall be included in all
   * copies or substantial portions of the Software.
   *
   * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   * SOFTWARE.
   *
   */
  let { $$slots, $$events, ...props } = $$props;
  const iconNode = [
    [
      "path",
      {
        "d": "M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915"
      }
    ],
    ["circle", { "cx": "12", "cy": "12", "r": "3" }]
  ];
  Icon($$payload, spread_props([
    { name: "settings" },
    /**
     * @component @name Settings
     * @description Lucide SVG icon component, renders SVG Element with children.
     *
     * @preview ![img](data:image/svg+xml;base64,PHN2ZyAgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIgogIHdpZHRoPSIyNCIKICBoZWlnaHQ9IjI0IgogIHZpZXdCb3g9IjAgMCAyNCAyNCIKICBmaWxsPSJub25lIgogIHN0cm9rZT0iIzAwMCIgc3R5bGU9ImJhY2tncm91bmQtY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDJweCIKICBzdHJva2Utd2lkdGg9IjIiCiAgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIgogIHN0cm9rZS1saW5lam9pbj0icm91bmQiCj4KICA8cGF0aCBkPSJNOS42NzEgNC4xMzZhMi4zNCAyLjM0IDAgMCAxIDQuNjU5IDAgMi4zNCAyLjM0IDAgMCAwIDMuMzE5IDEuOTE1IDIuMzQgMi4zNCAwIDAgMSAyLjMzIDQuMDMzIDIuMzQgMi4zNCAwIDAgMCAwIDMuODMxIDIuMzQgMi4zNCAwIDAgMS0yLjMzIDQuMDMzIDIuMzQgMi4zNCAwIDAgMC0zLjMxOSAxLjkxNSAyLjM0IDIuMzQgMCAwIDEtNC42NTkgMCAyLjM0IDIuMzQgMCAwIDAtMy4zMi0xLjkxNSAyLjM0IDIuMzQgMCAwIDEtMi4zMy00LjAzMyAyLjM0IDIuMzQgMCAwIDAgMC0zLjgzMUEyLjM0IDIuMzQgMCAwIDEgNi4zNSA2LjA1MWEyLjM0IDIuMzQgMCAwIDAgMy4zMTktMS45MTUiIC8+CiAgPGNpcmNsZSBjeD0iMTIiIGN5PSIxMiIgcj0iMyIgLz4KPC9zdmc+Cg==) - https://lucide.dev/icons/settings
     * @see https://lucide.dev/guide/packages/lucide-svelte - Documentation
     *
     * @param {Object} props - Lucide icons props and any valid SVG attribute
     * @returns {FunctionalComponent} Svelte component
     *
     */
    props,
    {
      iconNode,
      children: ($$payload2) => {
        props.children?.($$payload2);
        $$payload2.out.push(`<!---->`);
      },
      $$slots: { default: true }
    }
  ]));
  pop();
}
const marked = new Marked(
  markedHighlight({
    emptyLangClass: "hljs",
    langPrefix: "hljs language-",
    highlight(code, lang) {
      const language = hljs.getLanguage(lang) ? lang : "plaintext";
      return hljs.highlight(code, { language }).value;
    }
  })
);
marked.setOptions({
  breaks: true,
  gfm: true
});
marked.use({
  renderer: {
    link(arg) {
      const base = marked.Renderer.prototype.link.call(this, arg);
      return base.replace("<a", '<a target="_blank" rel="noopener noreferrer"');
    }
  }
});
function renderMarkdown(content) {
  if (!content) return "";
  return marked.parse(content);
}
function MessageItemText($$payload, $$props) {
  push();
  let { item: item2, role } = $$props;
  const renderedContent = role === "assistant" ? renderMarkdown(item2.text) : item2.text;
  $$payload.out.push(`<div${attr_class(clsx([
    "prose w-full max-w-none rounded-box p-2 text-base-content",
    {
      "mb-3": role === "assistant",
      "p-4": role === "assistant",
      "bg-base-200": role === "user"
    }
  ]))}>${html(renderedContent)}</div>`);
  pop();
}
function MessageItemImage($$payload, $$props) {
  push();
  let { item: item2 } = $$props;
  $$payload.out.push(`<div class="mb-3 rounded-lg bg-base-200 p-3"><img${attr("src", `data:${stringify(item2.mimeType)};base64,${stringify(item2.data)}`)} alt="" class="max-w-full rounded"/></div>`);
  pop();
}
function MessageItemAudio($$payload, $$props) {
  push();
  let { item: item2 } = $$props;
  $$payload.out.push(`<div class="mb-3 rounded-lg bg-base-200 p-3"><audio controls class="w-full"><source${attr("src", `data:${stringify(item2.mimeType)};base64,${stringify(item2.data)}`)}${attr("type", item2.mimeType)}/> Your browser does not support the audio element.</audio></div>`);
  pop();
}
function MessageItemResourceLink($$payload, $$props) {
  push();
  let { item: item2 } = $$props;
  $$payload.out.push(`<div class="mb-3 rounded-lg bg-base-200 p-3"><div class="flex items-center gap-2 text-sm">`);
  External_link($$payload, { class: "h-4 w-4 text-primary" });
  $$payload.out.push(`<!----> <span>${escape_html(item2.name || item2.uri)}</span></div> `);
  if (item2.description) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<p class="mt-1 text-xs opacity-70">${escape_html(item2.description)}</p>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
function item($$payload, label, type, loading, name, id, onClick) {
  $$payload.out.push(`<div class="flex items-center gap-2 rounded-xl bg-base-200 px-3 py-2 text-sm">`);
  if (loading) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="loading loading-xs loading-spinner"></span>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<span>${escape_html(getFileIcon(type))}</span>`);
  }
  $$payload.out.push(`<!--]--> <span class="max-w-32 truncate">${escape_html(name)}</span> <button type="button" class="btn h-5 w-5 rounded-full p-0 btn-ghost btn-xs"${attr("aria-label", label)}>`);
  X($$payload, { class: "h-3 w-3" });
  $$payload.out.push(`<!----></button></div>`);
}
function getFileIcon(type) {
  if (type?.startsWith("image/")) {
    return "🖼️";
  } else if (type === "application/pdf") {
    return "📄";
  } else if (type?.includes("text/") || type?.includes("json")) {
    return "📝";
  } else if (type?.includes("spreadsheet") || type?.includes("csv")) {
    return "📊";
  } else if (type?.includes("document")) {
    return "📋";
  }
  return "📎";
}
function MessageAttachments($$payload, $$props) {
  push();
  let {
    uploadingFiles = [],
    uploadedFiles = [],
    selectedResources = []
  } = $$props;
  if (uploadedFiles.length > 0 || uploadingFiles.length > 0 || selectedResources.length > 0) {
    $$payload.out.push("<!--[-->");
    const each_array = ensure_array_like(uploadingFiles);
    const each_array_1 = ensure_array_like(uploadedFiles);
    const each_array_2 = ensure_array_like(selectedResources);
    $$payload.out.push(`<div class="flex flex-wrap gap-2"><!--[-->`);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let uploadingFile = each_array[$$index];
      item($$payload, "Cancel upload", "", true, uploadingFile.file.name, uploadingFile.id);
    }
    $$payload.out.push(`<!--]--> <!--[-->`);
    for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
      let uploadedFile = each_array_1[$$index_1];
      item($$payload, "Remove file", uploadedFile.file.type, false, uploadedFile.file.name, uploadedFile.id);
    }
    $$payload.out.push(`<!--]--> <!--[-->`);
    for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
      let resource = each_array_2[$$index_2];
      item($$payload, "Remove resource", resource.mimeType || "", false, resource.title || resource.name);
    }
    $$payload.out.push(`<!--]--></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function MessageItemResource($$payload, $$props) {
  push();
  let { item: item2 } = $$props;
  const isError = item2.resource.mimeType === "application/vnd.nanobot.error+json";
  function isTextType(mimeType) {
    return mimeType.startsWith("text/") || mimeType.includes("json") || mimeType.includes("xml") || mimeType === "application/javascript" || mimeType === "application/typescript";
  }
  function isPdfType(mimeType) {
    return mimeType === "application/pdf";
  }
  function getResourceDisplayName() {
    if (item2.resource.title) return item2.resource.title;
    if (item2.resource.name) return item2.resource.name;
    const mimeType = item2.resource.mimeType;
    if (mimeType === "application/json") return "JSON Data";
    if (mimeType === "application/xml") return "XML Document";
    if (mimeType === "application/pdf") return "PDF Document";
    if (mimeType.startsWith("text/")) return "Text Document";
    if (mimeType.startsWith("image/")) return "Image";
    if (mimeType.includes("json")) return "JSON Resource";
    if (mimeType.includes("html")) return "HTML Document";
    if (mimeType.includes("csv")) return "CSV Data";
    if (mimeType.includes("markdown")) return "Markdown";
    return mimeType;
  }
  function formatFileSize(bytes) {
    if (!bytes) return "";
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
  function getDecodedText() {
    if (!item2.resource.blob) return "";
    try {
      const binaryString = atob(item2.resource.blob);
      const bytes = Uint8Array.from(binaryString, (c) => c.charCodeAt(0));
      const str = new TextDecoder("utf-8").decode(bytes);
      try {
        return JSON.stringify(JSON.parse(str), null, 2);
      } catch {
        return str;
      }
    } catch {
      return "Error decoding content";
    }
  }
  if (isError) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="mb-3 rounded-lg border border-error/20 bg-error/10 p-3"><div class="mb-2 flex items-center gap-2 text-sm">`);
    Triangle_alert($$payload, { class: "h-4 w-4 text-error" });
    $$payload.out.push(`<!----> <span class="font-medium text-error">Error</span></div> `);
    if (item2.resource.text) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<pre class="mt-2 rounded bg-base-100 p-2 text-xs break-all whitespace-pre-wrap text-error">${escape_html(item2.resource.text)}</pre>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<div class="card-compact card max-w-sm border border-base-200/50 shadow-md"><div class="card-body"><div class="flex items-start gap-3"><div class="flex-shrink-0"><div class="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-2xl">${escape_html(getFileIcon(item2.resource.mimeType))}</div></div> <div class="min-w-0 flex-1"><h4 class="truncate text-sm font-semibold text-base-content">${escape_html(getResourceDisplayName())}</h4> `);
    if (item2.resource.description) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<p class="mt-1 text-xs text-base-content/60" style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;">${escape_html(item2.resource.description)}</p>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> <div class="mt-2 flex items-center gap-2"><span class="badge badge-ghost badge-xs">${escape_html(item2.resource.mimeType)}</span> `);
    if (item2.resource.size) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="text-xs text-base-content/50">${escape_html(formatFileSize(item2.resource.size))}</span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div></div></div> <div class="mt-3 card-actions justify-end"><button type="button" class="btn btn-sm btn-primary">View Content</button></div></div></div> <dialog class="modal modal-bottom sm:modal-middle"><div class="modal-box max-h-[80vh] max-w-4xl overflow-hidden"><div class="mb-4 flex items-center justify-between"><h3 class="flex items-center gap-3 text-lg font-bold"><div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10"><span class="text-xl">${escape_html(getFileIcon(item2.resource.mimeType))}</span></div> <div><div class="text-base-content">${escape_html(getResourceDisplayName())}</div> `);
    if (item2.resource.description) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="text-sm font-normal text-base-content/60">${escape_html(item2.resource.description)}</div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div></h3></div> <div class="mb-4 flex items-center gap-2"><span class="badge badge-sm badge-primary">${escape_html(item2.resource.mimeType)}</span> `);
    if (item2.resource.size) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="badge badge-ghost badge-sm">${escape_html(formatFileSize(item2.resource.size))}</span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    if (item2.resource.annotations?.lastModified) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="text-xs text-base-content/50">Modified: ${escape_html(new Date(item2.resource.annotations.lastModified).toLocaleDateString())}</span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div> <div class="max-h-96 overflow-auto">`);
    if (isTextType(item2.resource.mimeType) && item2.resource.blob) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="mockup-code"><pre><code>${escape_html(getDecodedText())}</code></pre></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
      if (isPdfType(item2.resource.mimeType) && item2.resource.blob) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<div class="w-full"><iframe${attr("src", `data:application/pdf;base64,${stringify(item2.resource.blob)}`)} class="h-96 w-full rounded border border-base-300" title="PDF Viewer"></iframe></div>`);
      } else {
        $$payload.out.push("<!--[!-->");
        $$payload.out.push(`<div class="py-8 text-center"><div class="mb-4 text-6xl">${escape_html(getFileIcon(item2.resource.mimeType))}</div> <p class="text-base-content/60">Preview not available for this resource type</p> `);
        if (item2.resource.blob) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<p class="mt-2 text-sm text-base-content/40">Resource data is available but cannot be previewed</p>`);
        } else {
          $$payload.out.push("<!--[!-->");
          $$payload.out.push(`<p class="mt-2 text-sm text-base-content/40">No resource data available</p>`);
        }
        $$payload.out.push(`<!--]--></div>`);
      }
      $$payload.out.push(`<!--]-->`);
    }
    $$payload.out.push(`<!--]--></div> <div class="modal-action"><form method="dialog"><button class="btn">Close</button></form></div></div></dialog>`);
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function MessageItemReasoning($$payload, $$props) {
  push();
  let { item: item2 } = $$props;
  $$payload.out.push(`<div class="mb-3 rounded-lg border-l-4 border-neutral bg-neutral/20 p-3"><div class="mb-2 flex items-center gap-2 text-sm">`);
  Lightbulb($$payload, { class: "h-4 w-4 text-neutral" });
  $$payload.out.push(`<!----> <span class="font-medium text-neutral">Reasoning</span></div> `);
  if (item2.summary) {
    $$payload.out.push("<!--[-->");
    const each_array = ensure_array_like(item2.summary);
    $$payload.out.push(`<!--[-->`);
    for (let index = 0, $$length = each_array.length; index < $$length; index++) {
      let summaryItem = each_array[index];
      $$payload.out.push(`<p class="text-sm opacity-80">${escape_html(summaryItem.text)}</p>`);
    }
    $$payload.out.push(`<!--]-->`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
function MessageItemUI($$payload, $$props) {
  push();
  let { item: item2, onSend, style = {} } = $$props;
  React.createRef();
  $$payload.out.push(`<div class="contents"></div>`);
  pop();
}
function MessageItemTool($$payload, $$props) {
  push();
  let { item: item2, onSend } = $$props;
  let singleUIResource = item2.output?.content && item2.output?.content?.filter((i) => {
    return isUIResource(i) && !i.resource?._meta?.["ai.nanobo.meta/workspace"];
  }).length === 1;
  function parseToolInput(input) {
    try {
      const parsed = JSON.parse(input);
      if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
        return { success: true, data: parsed };
      }
    } catch {
    }
    return { success: false, data: input };
  }
  function getStyle(item3, singleUIResource2 = false) {
    if (singleUIResource2) {
      return {};
    }
    if (isUIResource(item3) && item3.resource._meta?.["mcpui.dev/ui-preferred-frame-size"]) {
      const coords = item3.resource._meta["mcpui.dev/ui-preferred-frame-size"];
      if (Array.isArray(coords) && coords[0] && coords[1]) {
        return { width: `${coords[0]}`, height: `${coords[1]}` };
      } else if (coords && typeof coords === "object" && "height" in coords && "width" in coords) {
        return { width: `${coords.width}`, height: `${coords.height}` };
      }
    }
    return { width: "300px", height: "400px" };
  }
  function parseToolOutput(output) {
    try {
      const parsed = JSON.parse(output);
      const formattedJson = JSON.stringify(parsed, null, 2);
      const highlightedJson = renderMarkdown("```json\n" + formattedJson + "\n```");
      return { success: true, data: highlightedJson };
    } catch {
    }
    return { success: false, data: output };
  }
  $$payload.out.push(`<div class="text collapse mt-3 mb-2 w-full border border-base-100 bg-base-100 hover:collapse-arrow hover:border-base-300"><input type="checkbox"/> <div class="collapse-title"><div class="flex items-center gap-2">`);
  if (item2.output) {
    $$payload.out.push("<!--[-->");
    Settings($$payload, { class: "h-4 w-4 text-primary/60" });
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<span class="loading loading-xs loading-spinner"></span>`);
  }
  $$payload.out.push(`<!--]--> <span class="text-sm font-medium text-primary/60">Tool call: ${escape_html(item2.name)}</span></div></div> <div class="collapse-content"><div class="space-y-3 pt-2">`);
  if (item2.arguments) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="grid"><div class="mb-1 text-xs font-medium text-base-content/70">Input:</div> `);
    if (parseToolInput(item2.arguments).success) {
      $$payload.out.push("<!--[-->");
      const each_array = ensure_array_like(Object.entries(parseToolInput(item2.arguments).data));
      $$payload.out.push(`<div class="overflow-x-auto rounded bg-base-200 p-3"><table class="table w-full table-zebra table-xs"><thead><tr><th class="text-xs">Key</th><th class="text-xs">Value</th></tr></thead><tbody><!--[-->`);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let [key, value] = each_array[$$index];
        $$payload.out.push(`<tr><td class="font-mono text-xs">${escape_html(key)}</td><td class="font-mono text-xs break-all">${escape_html(typeof value === "object" ? JSON.stringify(value) : String(value))}</td></tr>`);
      }
      $$payload.out.push(`<!--]-->`);
      if (Object.keys(parseToolInput(item2.arguments).data).length === 0) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<tr><td class="font-mono text-xs">No arguments</td></tr>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></tbody></table></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
      $$payload.out.push(`<div class="overflow-x-auto rounded bg-base-200 p-3 font-mono text-sm">${escape_html(item2.arguments)}</div>`);
    }
    $$payload.out.push(`<!--]--></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (item2.output) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex flex-col"><div class="mb-1 text-xs font-medium text-base-content/70">Output:</div> `);
    if (item2.output.isError) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="alert alert-error"><span>Tool execution failed</span></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    if (item2.output.structuredContent) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="prose overflow-x-auto rounded border border-success/20 bg-success/10 p-3 prose-invert">${html(parseToolOutput(JSON.stringify(item2.output.structuredContent)).data)}</div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    if (item2.output.content) {
      $$payload.out.push("<!--[-->");
      const each_array_1 = ensure_array_like(item2.output.content);
      $$payload.out.push(`<!--[-->`);
      for (let i = 0, $$length = each_array_1.length; i < $$length; i++) {
        let contentItem = each_array_1[i];
        $$payload.out.push(`<div class="prose overflow-x-auto rounded border border-success/20 bg-success/10 p-3 prose-invert">`);
        if (contentItem.type === "text" && "text" in contentItem && parseToolOutput(contentItem.text).success) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`${html(parseToolOutput(contentItem.text).data)}`);
        } else {
          $$payload.out.push("<!--[!-->");
          if (contentItem.type === "text" && "text" in contentItem && contentItem.text) {
            $$payload.out.push("<!--[-->");
            $$payload.out.push(`${html(renderMarkdown(contentItem.text))}`);
          } else {
            $$payload.out.push("<!--[!-->");
            $$payload.out.push(`${html(parseToolOutput(JSON.stringify(contentItem)).data)}`);
          }
          $$payload.out.push(`<!--]-->`);
        }
        $$payload.out.push(`<!--]--></div>`);
      }
      $$payload.out.push(`<!--]-->`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<div class="flex items-center gap-2 text-xs text-base-content/50 italic"><span class="loading loading-xs loading-spinner"></span> Running...</div>`);
  }
  $$payload.out.push(`<!--]--></div></div></div> <div class="flex w-full flex-wrap items-start justify-start gap-2 p-2">`);
  if (item2.output && item2.output.content) {
    $$payload.out.push("<!--[-->");
    const each_array_2 = ensure_array_like(item2.output.content);
    $$payload.out.push(`<!--[-->`);
    for (let i = 0, $$length = each_array_2.length; i < $$length; i++) {
      let contentItem = each_array_2[i];
      if (contentItem.type === "resource" && contentItem.resource && isUIResource(contentItem) && !contentItem.resource._meta?.["ai.nanobot.meta/workspace"]) {
        $$payload.out.push("<!--[-->");
        MessageItemUI($$payload, {
          item: contentItem,
          onSend,
          style: getStyle(contentItem, singleUIResource)
        });
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--> `);
      if (contentItem.type === "image") {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<img${attr("src", `data:${stringify(contentItem.mimeType)};base64,${stringify(contentItem.data)}`)} alt="Tool output" class="max-w-full rounded"/>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]-->`);
    }
    $$payload.out.push(`<!--]-->`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
function MessageItem($$payload, $$props) {
  push();
  let { item: item2, role, onSend } = $$props;
  if (item2.type === "text") {
    $$payload.out.push("<!--[-->");
    MessageItemText($$payload, { item: item2, role });
  } else {
    $$payload.out.push("<!--[!-->");
    if (item2.type === "image") {
      $$payload.out.push("<!--[-->");
      MessageItemImage($$payload, { item: item2 });
    } else {
      $$payload.out.push("<!--[!-->");
      if (item2.type === "audio") {
        $$payload.out.push("<!--[-->");
        MessageItemAudio($$payload, { item: item2 });
      } else {
        $$payload.out.push("<!--[!-->");
        if (item2.type === "resource_link") {
          $$payload.out.push("<!--[-->");
          MessageItemResourceLink($$payload, { item: item2 });
        } else {
          $$payload.out.push("<!--[!-->");
          if (item2.type === "resource") {
            $$payload.out.push("<!--[-->");
            MessageItemResource($$payload, { item: item2 });
          } else {
            $$payload.out.push("<!--[!-->");
            if (item2.type === "reasoning") {
              $$payload.out.push("<!--[-->");
              MessageItemReasoning($$payload, { item: item2 });
            } else {
              $$payload.out.push("<!--[!-->");
              if (item2.type === "tool") {
                $$payload.out.push("<!--[-->");
                MessageItemTool($$payload, { item: item2, onSend });
              } else {
                $$payload.out.push("<!--[!-->");
              }
              $$payload.out.push(`<!--]-->`);
            }
            $$payload.out.push(`<!--]-->`);
          }
          $$payload.out.push(`<!--]-->`);
        }
        $$payload.out.push(`<!--]-->`);
      }
      $$payload.out.push(`<!--]-->`);
    }
    $$payload.out.push(`<!--]-->`);
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function Message($$payload, $$props) {
  push();
  let { message, timestamp, onSend } = $$props;
  const displayTime = timestamp || (message.created ? new Date(message.created) : /* @__PURE__ */ new Date());
  const toolCall = (() => {
    try {
      const item2 = message.items?.[0];
      return message.role === "user" && item2?.type === "text" ? JSON.parse(item2.text) : null;
    } catch {
      return null;
    }
  })();
  if (message.role === "user" && toolCall?.type === "prompt" && toolCall.payload?.prompt) {
    $$payload.out.push("<!--[-->");
    MessageItemText($$payload, {
      item: {
        id: crypto.randomUUID(),
        type: "text",
        text: toolCall.payload?.prompt
      },
      role: "user"
    });
  } else {
    $$payload.out.push("<!--[!-->");
    if (message.role === "user" && toolCall?.type === "tool" && toolCall.payload?.toolName) {
      $$payload.out.push("<!--[-->");
    } else {
      $$payload.out.push("<!--[!-->");
      if (message.role === "user") {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<div class="group flex w-full justify-end"><div class="max-w-md"><div class="flex flex-col items-end"><div class="rounded-box bg-base-200 p-2">`);
        if (message.items && message.items.length > 0) {
          $$payload.out.push("<!--[-->");
          const each_array = ensure_array_like(message.items);
          $$payload.out.push(`<!--[-->`);
          for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
            let item2 = each_array[$$index];
            MessageItem($$payload, { item: item2, role: message.role });
          }
          $$payload.out.push(`<!--]-->`);
        } else {
          $$payload.out.push("<!--[!-->");
          $$payload.out.push(`<p>No content</p>`);
        }
        $$payload.out.push(`<!--]--></div> <div class="transition-duration-500 mb-1 text-sm font-medium opacity-0 transition-opacity group-hover:opacity-100"><time class="ml-2 text-xs opacity-50">${escape_html(displayTime.toLocaleTimeString())}</time></div></div></div></div>`);
      } else {
        $$payload.out.push("<!--[!-->");
        $$payload.out.push(`<div class="flex w-full items-start gap-3"><div class="flex min-w-0 flex-1 flex-col items-start">`);
        if (message.items && message.items.length > 0) {
          $$payload.out.push("<!--[-->");
          const each_array_1 = ensure_array_like(message.items);
          $$payload.out.push(`<!--[-->`);
          for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
            let item2 = each_array_1[$$index_1];
            MessageItem($$payload, { item: item2, role: message.role, onSend });
          }
          $$payload.out.push(`<!--]-->`);
        } else {
          $$payload.out.push("<!--[!-->");
          $$payload.out.push(`<div class="prose w-full max-w-full rounded-lg bg-base-200 p-3 prose-invert"><p>No content</p></div>`);
        }
        $$payload.out.push(`<!--]--></div></div>`);
      }
      $$payload.out.push(`<!--]-->`);
    }
    $$payload.out.push(`<!--]-->`);
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function AgentHeader($$payload, $$props) {
  push();
  let { agent } = $$props;
  $$payload.out.push(`<div class="flex flex-col items-center p-8 pt-20">`);
  if (agent?.name) {
    $$payload.out.push("<!--[-->");
    if (agent.icon) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="mb-6"><img${attr("src", agent.icon)}${attr("alt", agent.name)} class="h-16"/></div> <div class="mb-8 text-center"><p class="max-w-md text-base-content/70">${escape_html(agent.description || "")}</p></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
      $$payload.out.push(`<div class="mb-6"><div class="flex h-20 w-20 items-center justify-center rounded-full bg-primary/20"><svg class="h-10 w-10 text-primary" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-6-3a2 2 0 11-4 0 2 2 0 014 0zm-2 4a5 5 0 00-4.546 2.916A5.986 5.986 0 0010 16a5.986 5.986 0 004.546-2.084A5 5 0 0010 11z" clip-rule="evenodd"></path></svg></div></div> <div class="mb-8 text-center"><h2 class="mb-2 text-2xl font-semibold text-base-content">${escape_html(agent.name)}</h2> <p class="max-w-md text-base-content/70">${escape_html(agent.description || "")}</p></div>`);
    }
    $$payload.out.push(`<!--]-->`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (agent) {
    $$payload.out.push("<!--[-->");
    const each_array = ensure_array_like(agent.starterMessages || []);
    $$payload.out.push(`<div${attr_class(clsx([
      "grid w-full max-w-2xl grid-cols-1 gap-3",
      {
        "grid-cols-2": agent.starterMessages?.length === 2,
        "grid-cols-3": agent.starterMessages?.length ?? 0 > 2
      }
    ]))}><!--[-->`);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let message = each_array[$$index];
      $$payload.out.push(`<button class="card-compact card cursor-pointer bg-base-200 shadow-sm transition-colors hover:bg-base-300"><span class="card-body text-sm">${escape_html(message)}</span></button>`);
    }
    $$payload.out.push(`<!--]--></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
function Messages($$payload, $$props) {
  push();
  let { messages, onSend, isLoading = false, agent } = $$props;
  let messageGroups = (() => {
    return messages.reduce(
      (acc, message) => {
        if (message.role === "user" || acc.length === 0) {
          acc.push([message]);
        } else {
          acc[acc.length - 1].push(message);
        }
        return acc;
      },
      []
    );
  })();
  let hasMessageContent = messageGroups[messageGroups.length - 1]?.some((message) => message.role === "assistant" && message.items && message.items.some((item2) => item2.type === "tool" || item2.type === "text" && item2.text && item2.text.trim().length > 0));
  let showLoadingIndicator = isLoading && !hasMessageContent;
  $$payload.out.push(`<div id="message-groups" class="flex flex-col space-y-4 pt-4">`);
  if (messages.length === 0) {
    $$payload.out.push("<!--[-->");
    AgentHeader($$payload, { agent });
  } else {
    $$payload.out.push("<!--[!-->");
    const lastIndex = messageGroups.length - 1;
    const each_array = ensure_array_like(messageGroups);
    $$payload.out.push(`<!--[-->`);
    for (let i = 0, $$length = each_array.length; i < $$length; i++) {
      let messageGroup = each_array[i];
      const isLast = i === lastIndex;
      const each_array_1 = ensure_array_like(messageGroup);
      $$payload.out.push(`<div${attr("id", `group-${messageGroup[0]?.id}`)}${attr_class(clsx({ "min-h-[calc(100vh-2rem)]": isLast, contents: !isLast }))}${attr("data-message-id", messageGroup[0]?.id)}><!--[-->`);
      for (let i2 = 0, $$length2 = each_array_1.length; i2 < $$length2; i2++) {
        let message = each_array_1[i2];
        Message($$payload, { message, onSend });
      }
      $$payload.out.push(`<!--]--> `);
      if (isLast) {
        $$payload.out.push("<!--[-->");
        if (showLoadingIndicator) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<div class="flex w-full items-start gap-3"><div class="flex min-w-0 flex-1 flex-col items-start"><div class="flex items-center justify-center p-8"><span class="loading loading-lg loading-spinner text-base-content/30"></span></div></div></div>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]--> <div class="h-59"></div>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></div>`);
    }
    $$payload.out.push(`<!--]-->`);
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
function MessageSlashPrompts($$payload, $$props) {
  push();
  let { prompts, onPrompt, message } = $$props;
  let showSlashCommands = message.trim().startsWith("/");
  let filteredPrompts = (() => {
    if (!showSlashCommands) return [];
    const query = message.trim().slice(1).toLowerCase();
    return prompts.filter((prompt) => prompt.name.toLowerCase().includes(query) || prompt.title?.toLowerCase().includes(query) || prompt.description?.toLowerCase().includes(query));
  })();
  let selectedCommandIndex = 0;
  let slashQuery = message.trim().slice(1).toLowerCase();
  function handleKeydown(e) {
    if (showSlashCommands) {
      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          selectedCommandIndex = Math.min(selectedCommandIndex + 1, filteredPrompts.length - 1);
          return true;
        case "ArrowUp":
          e.preventDefault();
          selectedCommandIndex = Math.max(selectedCommandIndex - 1, 0);
          return true;
        case "Enter":
          e.preventDefault();
          if (filteredPrompts[selectedCommandIndex]) {
            executeSlashCommand(filteredPrompts[selectedCommandIndex]);
          }
          return true;
      }
    }
    return false;
  }
  function executeSlashCommand(prompt) {
    onPrompt?.(prompt.name);
  }
  if (showSlashCommands) {
    $$payload.out.push("<!--[-->");
    const each_array = ensure_array_like(filteredPrompts);
    $$payload.out.push(`<div class="z-50 max-h-60 w-full overflow-y-auto rounded-lg border border-base-300 bg-base-100 shadow-lg" style="top: calc(100% + 0.5rem);"><!--[-->`);
    for (let index = 0, $$length = each_array.length; index < $$length; index++) {
      let prompt = each_array[index];
      $$payload.out.push(`<button type="button"${attr_class(`w-full px-4 py-2 text-left transition-colors hover:bg-base-200 ${stringify(index === selectedCommandIndex ? "bg-primary/10" : "")}`)}><div class="flex items-center space-x-2"><span class="font-mono text-sm text-primary">/${escape_html(prompt.name)}</span> `);
      if (prompt.title && prompt.title !== prompt.name) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<span class="text-sm font-medium">${escape_html(prompt.title)}</span>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></div> `);
      if (prompt.description) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<div class="mt-1 text-xs text-base-content/60">${escape_html(prompt.description)}</div>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></button>`);
    }
    $$payload.out.push(`<!--]--> `);
    if (filteredPrompts.length === 0) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="px-4 py-2 text-sm text-base-content/50">No commands found for "${escape_html(slashQuery)}"</div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  bind_props($$props, { handleKeydown });
  pop();
}
function MessageResources($$payload, $$props) {
  push();
  let {
    disabled = false,
    resources = [],
    messages = [],
    selectedResources
  } = $$props;
  let allResources = (() => {
    const embeddedResources = [];
    for (const message of messages || []) {
      if (message.role !== "assistant") continue;
      for (const item2 of message.items || []) {
        if (item2.type !== "tool") continue;
        for (const content of item2.output?.content || []) {
          if (content.type === "resource") {
            embeddedResources.push({
              uri: content.resource.uri,
              name: content.resource.name || content.resource.uri.split("/").pop() || content.resource.uri,
              description: content.resource.description,
              title: content.resource.title,
              mimeType: content.resource.mimeType,
              size: content.resource.size,
              annotations: content.resource.annotations,
              _meta: content.resource._meta
            });
          } else if (content.type === "resource_link") {
            embeddedResources.push({
              uri: content.uri,
              name: content.name || content.uri.split("/").pop() || content.uri,
              title: content.name || content.uri.split("/").pop() || content.uri,
              description: content.description
            });
          }
        }
      }
    }
    return [...resources, ...embeddedResources].filter((resource, index, self) => index === self.findIndex((r) => r.uri === resource.uri) && !resource.uri.startsWith("ui:") && !resource.uri.startsWith("chat:"));
  })();
  if (
    // No need for click handler with DaisyUI dropdown
    allResources.length > 0
  ) {
    $$payload.out.push("<!--[-->");
    const each_array = ensure_array_like(allResources);
    $$payload.out.push(`<div class="dropdown dropdown-end dropdown-top"><button class="btn h-9 w-9 rounded-full p-0 btn-ghost btn-sm"${attr("disabled", disabled, true)} aria-label="Select resources">`);
    Library($$payload, { class: "h-4 w-4" });
    $$payload.out.push(`<!----></button> <ul class="dropdown-content menu z-50 max-h-[50vh] w-64 overflow-y-auto rounded-box border border-base-300 bg-base-100 p-2 shadow-lg"><li class="menu-title"><span>Available Resources</span></li> <!--[-->`);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let resource = each_array[$$index];
      $$payload.out.push(`<li><button type="button"${attr_class(`flex items-center space-x-2 ${stringify(selectedResources.some((r) => r.uri === resource.uri) ? "active" : "")}`)}><span class="text-base">${escape_html(getFileIcon(resource.mimeType))}</span> <span class="flex-1 overflow-hidden"><span class="block truncate text-sm font-medium">${escape_html(resource.title || resource.name)}</span> `);
      if (resource.description) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<span class="block truncate text-xs opacity-60">${escape_html(resource.description)}</span>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></span> `);
      if (selectedResources.some((r) => r.uri === resource.uri)) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<span class="inline-block h-2 w-2 rounded-full bg-primary"></span>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></button></li>`);
    }
    $$payload.out.push(`<!--]--></ul></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function MessageInput($$payload, $$props) {
  push();
  let {
    onPrompt,
    placeholder = "Type a message...",
    disabled = false,
    uploadingFiles = [],
    uploadedFiles = [],
    prompts = [],
    resources = [],
    messages = [],
    supportedMimeTypes = [
      "image/*",
      "text/plain",
      "application/pdf",
      "application/json",
      "text/csv"
    ]
  } = $$props;
  let message = "";
  let isUploading = false;
  let selectedResources = [];
  $$payload.out.push(`<div class="p-0 md:p-4">`);
  MessageSlashPrompts($$payload, {
    prompts,
    message,
    onPrompt: (
      // Auto-resize when message changes
      (p) => {
        message = "";
        onPrompt?.(p);
      }
    )
  });
  $$payload.out.push(`<!----> <input type="file"${attr("accept", supportedMimeTypes.join(","))} class="hidden" aria-label="File upload"/> <form><div class="space-y-3 rounded-t-2xl border-2 border-base-200 bg-base-100 p-3 transition-colors focus-within:border-primary md:rounded-2xl"><textarea${attr("placeholder", placeholder)} class="max-h-32 min-h-[2.5rem] w-full resize-none bg-transparent p-1 text-sm leading-6 outline-none placeholder:text-base-content/50" rows="1">`);
  const $$body = escape_html(message);
  if ($$body) {
    $$payload.out.push(`${$$body}`);
  }
  $$payload.out.push(`</textarea> <div${attr_class(`flex items-center ${stringify(uploadedFiles.length > 0 || uploadingFiles.length > 0 || selectedResources.length > 0 ? "justify-between" : "justify-end")}`)}><select class="select hidden w-48 select-ghost select-sm"${attr("disabled", disabled || isUploading, true)}><option value="gpt-4"${maybe_selected($$payload, "gpt-4")}>GPT-4</option><option value="gpt-3.5-turbo"${maybe_selected($$payload, "gpt-3.5-turbo")}>GPT-3.5 Turbo</option><option value="claude-3-opus"${maybe_selected($$payload, "claude-3-opus")}>Claude 3 Opus</option><option value="claude-3-sonnet"${maybe_selected($$payload, "claude-3-sonnet")}>Claude 3 Sonnet</option><option value="gemini-pro"${maybe_selected($$payload, "gemini-pro")}>Gemini Pro</option></select> `);
  MessageAttachments($$payload, {
    selectedResources,
    uploadedFiles,
    uploadingFiles
  });
  $$payload.out.push(`<!----> <div class="flex gap-2"><button type="button" class="btn h-9 w-9 rounded-full p-0 btn-ghost btn-sm"${attr("disabled", disabled || isUploading, true)} aria-label="Attach file">`);
  {
    $$payload.out.push("<!--[!-->");
    Paperclip($$payload, { class: "h-4 w-4" });
  }
  $$payload.out.push(`<!--]--></button> `);
  MessageResources($$payload, {
    disabled,
    resources,
    selectedResources,
    messages
  });
  $$payload.out.push(`<!----> <button type="submit" class="btn h-9 w-9 rounded-full p-0 btn-sm btn-primary"${attr("disabled", disabled || isUploading || !message.trim(), true)} aria-label="Send message">`);
  if (disabled && !isUploading) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="loading loading-xs loading-spinner"></span>`);
  } else {
    $$payload.out.push("<!--[!-->");
    Send($$payload, { class: "h-4 w-4" });
  }
  $$payload.out.push(`<!--]--></button></div></div></div></form></div>`);
  pop();
}
function Elicitation($$payload, $$props) {
  push();
  let { elicitation, open = false } = $$props;
  let formData = {};
  function isRequired(key) {
    return elicitation.requestedSchema.required?.includes(key) ?? false;
  }
  function getFieldTitle(key, schema) {
    return schema.title || key;
  }
  function validateForm() {
    if (!elicitation.requestedSchema.required) return true;
    for (const requiredField of elicitation.requestedSchema.required) {
      const value = formData[requiredField];
      if (value === void 0 || value === "" || value === null) {
        return false;
      }
    }
    return true;
  }
  function isOAuthElicitation() {
    return Boolean(elicitation._meta?.["ai.nanobot.meta/oauth-url"]);
  }
  function getOAuthUrl() {
    return elicitation._meta?.["ai.nanobot.meta/oauth-url"];
  }
  function getServerName() {
    return elicitation._meta?.["ai.nanobot.meta/server-name"] || "MCP Server";
  }
  if (open) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<dialog class="modal-open modal"><div class="modal-box w-full max-w-2xl"><form method="dialog"><button class="btn absolute top-2 right-2 btn-circle btn-ghost btn-sm">✕</button></form> `);
    if (isOAuthElicitation()) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<h3 class="mb-4 text-lg font-bold">Authentication Required</h3> <div class="mb-6"><p class="mb-4 text-base-content/80">The <strong>${escape_html(getServerName())}</strong> server requires authentication to continue.</p> <p class="mb-4 text-base-content/80">Please click the link below to authenticate:</p> <div class="group relative mb-4 rounded-lg bg-base-200 p-4"><p class="pr-8 font-mono text-sm break-all text-base-content/90">${escape_html(getOAuthUrl())}</p> <button type="button" class="btn absolute top-2 right-2 opacity-60 btn-ghost transition-opacity btn-xs hover:opacity-100" title="Copy to clipboard">`);
      Copy($$payload, { class: "h-4 w-4" });
      $$payload.out.push(`<!----></button> `);
      {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></div></div> <div class="modal-action"><button type="button" class="btn btn-error">Decline</button> <button type="button" class="btn btn-success">Authenticate</button></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
      const each_array = ensure_array_like(Object.entries(elicitation.requestedSchema.properties));
      $$payload.out.push(`<h3 class="mb-4 text-lg font-bold">Information Request</h3> <div class="mb-6"><p class="whitespace-pre-wrap text-base-content/80">${escape_html(elicitation.message)}</p></div> <form class="space-y-4"><!--[-->`);
      for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
        let [key, schema] = each_array[$$index_1];
        $$payload.out.push(`<div class="form-control"><label class="label"${attr("for", key)}><span class="label-text font-medium">${escape_html(getFieldTitle(key, schema))} `);
        if (isRequired(key)) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<span class="text-error">*</span>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]--></span></label> `);
        if (schema.description) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<div class="label"><span class="label-text-alt text-base-content/60">${escape_html(schema.description)}</span></div>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]--> `);
        if (schema.type === "string" && "enum" in schema) {
          $$payload.out.push("<!--[-->");
          const each_array_1 = ensure_array_like(schema.enum);
          $$payload.out.push(`<select${attr("id", key)} class="select-bordered select w-full"${attr("required", isRequired(key), true)}>`);
          $$payload.select_value = formData[key];
          $$payload.out.push(`<!--[-->`);
          for (let i = 0, $$length2 = each_array_1.length; i < $$length2; i++) {
            let option = each_array_1[i];
            $$payload.out.push(`<option${attr("value", option)}${maybe_selected($$payload, option)}>${escape_html(schema.enumNames?.[i] || option)}</option>`);
          }
          $$payload.out.push(`<!--]-->`);
          $$payload.select_value = void 0;
          $$payload.out.push(`</select>`);
        } else {
          $$payload.out.push("<!--[!-->");
          if (schema.type === "boolean") {
            $$payload.out.push("<!--[-->");
            $$payload.out.push(`<div class="form-control"><label class="label cursor-pointer justify-start gap-3"><input${attr("id", key)} type="checkbox"${attr("checked", Boolean(formData[key]), true)} class="checkbox"/> <span class="label-text">Enable</span></label></div>`);
          } else {
            $$payload.out.push("<!--[!-->");
            if (schema.type === "number" || schema.type === "integer") {
              $$payload.out.push("<!--[-->");
              $$payload.out.push(`<input${attr("id", key)} type="number"${attr("value", formData[key])} class="input-bordered input w-full"${attr("required", isRequired(key), true)}${attr("min", schema.minimum)}${attr("max", schema.maximum)}${attr("step", schema.type === "integer" ? "1" : "any")}/>`);
            } else {
              $$payload.out.push("<!--[!-->");
              if (schema.type === "string") {
                $$payload.out.push("<!--[-->");
                if (schema.format === "email") {
                  $$payload.out.push("<!--[-->");
                  $$payload.out.push(`<input${attr("id", key)} type="email"${attr("value", formData[key])} class="input-bordered input w-full"${attr("required", isRequired(key), true)}${attr("minlength", schema.minLength)}${attr("maxlength", schema.maxLength)}/>`);
                } else {
                  $$payload.out.push("<!--[!-->");
                  if (schema.format === "uri") {
                    $$payload.out.push("<!--[-->");
                    $$payload.out.push(`<input${attr("id", key)} type="url"${attr("value", formData[key])} class="input-bordered input w-full"${attr("required", isRequired(key), true)}${attr("minlength", schema.minLength)}${attr("maxlength", schema.maxLength)}/>`);
                  } else {
                    $$payload.out.push("<!--[!-->");
                    if (schema.format === "date") {
                      $$payload.out.push("<!--[-->");
                      $$payload.out.push(`<input${attr("id", key)} type="date"${attr("value", formData[key])} class="input-bordered input w-full"${attr("required", isRequired(key), true)}/>`);
                    } else {
                      $$payload.out.push("<!--[!-->");
                      if (schema.format === "date-time") {
                        $$payload.out.push("<!--[-->");
                        $$payload.out.push(`<input${attr("id", key)} type="datetime-local"${attr("value", formData[key])} class="input-bordered input w-full"${attr("required", isRequired(key), true)}/>`);
                      } else {
                        $$payload.out.push("<!--[!-->");
                        $$payload.out.push(`<input${attr("id", key)} type="text"${attr("value", formData[key])} class="input-bordered input w-full"${attr("required", isRequired(key), true)}${attr("minlength", schema.minLength)}${attr("maxlength", schema.maxLength)}/>`);
                      }
                      $$payload.out.push(`<!--]-->`);
                    }
                    $$payload.out.push(`<!--]-->`);
                  }
                  $$payload.out.push(`<!--]-->`);
                }
                $$payload.out.push(`<!--]-->`);
              } else {
                $$payload.out.push("<!--[!-->");
              }
              $$payload.out.push(`<!--]-->`);
            }
            $$payload.out.push(`<!--]-->`);
          }
          $$payload.out.push(`<!--]-->`);
        }
        $$payload.out.push(`<!--]--></div>`);
      }
      $$payload.out.push(`<!--]--></form> <div class="modal-action"><button type="button" class="btn btn-error">Decline</button> <button type="button" class="btn btn-primary"${attr("disabled", !validateForm(), true)}>Accept</button></div>`);
    }
    $$payload.out.push(`<!--]--></div></dialog>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function Prompt($$payload, $$props) {
  push();
  let { prompt, open = false } = $$props;
  let showDialog = open;
  let formData = {};
  function isRequired(arg) {
    return arg.required ?? false;
  }
  function validateForm() {
    if (!prompt.arguments) return true;
    for (const arg of prompt.arguments) {
      if (isRequired(arg) && (!formData[arg.name] || formData[arg.name].trim() === "")) {
        return false;
      }
    }
    return true;
  }
  if (!open) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="card cursor-pointer bg-base-100 shadow-md transition-shadow hover:shadow-lg" role="button" tabindex="0"><div class="card-body"><h3 class="card-title text-lg">${escape_html(prompt.title || prompt.name)}</h3> `);
    if (prompt.description) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<p class="text-sm text-base-content/70">${escape_html(prompt.description)}</p>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    if (prompt.arguments && prompt.arguments.length > 0) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="badge badge-sm badge-primary">${escape_html(prompt.arguments.length)} argument${escape_html(prompt.arguments.length === 1 ? "" : "s")}</div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (showDialog) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<dialog class="modal-open modal"><div class="modal-box w-full max-w-2xl"><form method="dialog"><button class="btn absolute top-2 right-2 btn-circle btn-ghost btn-sm">✕</button></form> <h3 class="mb-4 text-lg font-bold">${escape_html(prompt.title || prompt.name)}</h3> `);
    if (prompt.description) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="mb-6"><p class="text-base-content/80">${escape_html(prompt.description)}</p></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> <form class="space-y-4">`);
    if (prompt.arguments) {
      $$payload.out.push("<!--[-->");
      const each_array = ensure_array_like(prompt.arguments);
      $$payload.out.push(`<!--[-->`);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let arg = each_array[$$index];
        $$payload.out.push(`<div class="form-control"><label class="label"${attr("for", arg.name)}><span class="label-text font-medium">${escape_html(arg.title || arg.name)} `);
        if (isRequired(arg)) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<span class="text-error">*</span>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]--></span></label> `);
        if (arg.description) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<div class="label"><span class="label-text-alt text-base-content/60">${escape_html(arg.description)}</span></div>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]--> <input${attr("id", arg.name)} type="text"${attr("value", formData[arg.name])} class="input-bordered input w-full"${attr("required", isRequired(arg), true)}${attr("placeholder", arg.description || `Enter ${arg.name}`)}/></div>`);
      }
      $$payload.out.push(`<!--]-->`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></form> <div class="modal-action"><button type="button" class="btn btn-ghost">Cancel</button> <button type="button" class="btn btn-primary"${attr("disabled", !validateForm(), true)}>Execute Prompt</button></div></div> <form method="dialog" class="modal-backdrop"><button>close</button></form></dialog>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function Thread($$payload, $$props) {
  push();
  let {
    messages,
    prompts,
    resources,
    onSendMessage,
    onFileUpload,
    cancelUpload,
    uploadingFiles,
    uploadedFiles,
    elicitations,
    onElicitationResult,
    agent,
    isLoading = false
  } = $$props;
  let hasMessages = messages && messages.length > 0;
  let selectedPrompt = void 0;
  $$payload.out.push(`<div class="flex h-dvh w-full flex-col md:relative peer-[.workspace]:md:w-1/4"><div class="w-full overflow-y-auto"><div class="mx-auto max-w-4xl">`);
  if (prompts && prompts.length > 0) {
    $$payload.out.push("<!--[-->");
    const each_array = ensure_array_like(prompts);
    $$payload.out.push(`<div class="mb-6"><div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3"><!--[-->`);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let prompt = each_array[$$index];
      if (selectedPrompt === prompt.name) {
        $$payload.out.push("<!--[-->");
        Prompt($$payload, {
          prompt,
          open: true
        });
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]-->`);
    }
    $$payload.out.push(`<!--]--></div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  Messages($$payload, { messages, onSend: onSendMessage, isLoading, agent });
  $$payload.out.push(`<!----></div></div> <div${attr_class(`absolute right-0 bottom-0 left-0 flex flex-col transition-all duration-500 ease-in-out ${stringify(hasMessages ? "bg-base-100/80 backdrop-blur-sm" : "md:-translate-y-1/2 [@media(min-height:900px)]:md:top-1/2 [@media(min-height:900px)]:md:bottom-auto")}`)}>`);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <div class="mx-auto w-full max-w-4xl">`);
  MessageInput($$payload, {
    placeholder: `Type your message...${prompts && prompts.length > 0 ? " or / for prompts" : ""}`,
    resources,
    messages,
    onPrompt: (p) => selectedPrompt = p,
    disabled: isLoading,
    prompts,
    uploadingFiles,
    uploadedFiles
  });
  $$payload.out.push(`<!----></div></div> `);
  if (elicitations && elicitations.length > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<!---->`);
    {
      Elicitation($$payload, {
        elicitation: elicitations[0],
        open: true
      });
    }
    $$payload.out.push(`<!---->`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
export {
  MessageItemUI as M,
  Thread as T,
  onDestroy as o
};
