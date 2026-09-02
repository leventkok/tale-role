# image-worker

Theme-scoped scene images for the Storyteller **side panel**. Never inlined in chat. Does not block the turn loop.

F5 ships a **stub compositor**: SVG wash per one of the six themes, plus a redacted visual prompt. No paid image APIs (no DALL·E, Midjourney, Stability). Real local weights land later the same way as the LLM adapters — disk path, never git.

The game API embeds this library in-process and paints after the engine + stub narrator return. The turn JSON has no `image_svg`.
