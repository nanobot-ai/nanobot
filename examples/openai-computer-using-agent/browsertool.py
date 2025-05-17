import os
from typing import Union

from fastmcp import FastMCP, Context, Image
from mcp import SamplingMessage
from playwright.async_api import async_playwright, Page

CUA_KEY_TO_PLAYWRIGHT_KEY = {
    "/": "Divide",
    "\\": "Backslash",
    "alt": "Alt",
    "arrowdown": "ArrowDown",
    "arrowleft": "ArrowLeft",
    "arrowright": "ArrowRight",
    "arrowup": "ArrowUp",
    "backspace": "Backspace",
    "capslock": "CapsLock",
    "cmd": "Meta",
    "ctrl": "Control",
    "delete": "Delete",
    "end": "End",
    "enter": "Enter",
    "esc": "Escape",
    "home": "Home",
    "insert": "Insert",
    "option": "Alt",
    "pagedown": "PageDown",
    "pageup": "PageUp",
    "shift": "Shift",
    "space": " ",
    "super": "Meta",
    "tab": "Tab",
    "win": "Meta",
}

page: Union[Page | None] = None
mcp = FastMCP("openai-computer-using-agent")


async def screenshot_message() -> SamplingMessage:
    return SamplingMessage(
        role="user",
        content=(await screenshot()).to_image_content(),
    )


async def screenshot() -> Image:
    """Capture only the viewport (not full_page)."""
    png_bytes = await page.screenshot(full_page=False)
    return Image(data=png_bytes, format="png")


@mcp.tool(description="Opens the browser to a specific URL. Must be a valid https:// URL.")
async def open_url(url: str) -> str:
    global page
    await page.goto(url)
    return "Opened URL: " + url


@mcp.tool()
async def browser(
        type: str,
        x: int | None = None,
        y: int | None = None,
        scroll_x: int | None = None,
        scroll_y: int | None = None,
        button: str | None = None,
        path: list[tuple[int, int]] | None = None,
        keys: list[str] | None = None,
        text: str | None = None
) -> Image:
    match type:
        case "click":
            if button not in ("left", "right", "middle"):
                button = "left"
            await page.mouse.click(x, y, button=button)
        case "double_click":
            await page.mouse.dblclick(x, y)
        case "drag":
            if path:
                await page.mouse.move(path[0][0], path[0][1])
                await page.mouse.down()
                for px, py in path[1:]:
                    await page.mouse.move(px, py)
                await page.mouse.up()
        case "keypress":
            mapped_keys = [CUA_KEY_TO_PLAYWRIGHT_KEY.get(key.lower(), key) for
                           key in keys]
            for key in mapped_keys:
                await page.keyboard.down(key)
            for key in reversed(mapped_keys):
                await page.keyboard.up(key)
        case "move":
            await page.mouse.move(x, y)
        case "screenshot":
            pass
        case "scroll":
            await page.mouse.move(x, y)
            await page.evaluate(f"window.scrollBy({scroll_x}, {scroll_y})")
        case "type":
            await page.keyboard.type(text)
        case "wait":
            await asyncio.sleep(1)
    return await screenshot()


@mcp.tool(description="This agent is a computer use agent that can browse the "
                      "web and interact with web pages. It is designed to assist"
                      " users in finding information and performing tasks online.",
          annotations={
              "prompt": "The instructions for the browser agent.",
          })
async def browser_agent(prompt: str, ctx: Context) -> list[str | Image]:
    result = await ctx.sample(prompt)
    return [result.text.strip(), await screenshot()]


async def main():
    global page
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=os.environ.get("HEADLESS", "false") == "true")
        page = await browser.new_page()
        await page.goto("https://www.google.com")
        await page.set_viewport_size({"width": 1024, "height": 768})
        await mcp.run_async()


if __name__ == "__main__":
    import asyncio

    asyncio.run(main())
