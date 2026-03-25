/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/templates/**/*.templ",
    "./web/templates/**/*_templ.go",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["light", "dark"],
  },
}
