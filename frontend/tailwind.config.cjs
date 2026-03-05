/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}",
    "./node_modules/flowbite-svelte/**/*.{html,js,svelte,ts}",
    "./node_modules/flowbite/**/*.{js,ts}",
  ],
  theme: {
    extend: {
      colors: {
        accent: {
          50: "#fff4ec",
          500: "#ff7a1a",
          700: "#d65f10",
        },
      },
    },
  },
  plugins: [require("flowbite/plugin")],
};
