import svelte from "@astrojs/svelte";
import tailwind from "@astrojs/tailwind";
import { defineConfig } from "astro/config";

export default defineConfig({
  integrations: [svelte(), tailwind()],
  scopedStyleStrategy: "class",
});
