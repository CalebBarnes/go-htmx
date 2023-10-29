/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.html", "./components/**/*.html"],
  theme: {
    extend: {},
  },
  plugins: [require('@tailwindcss/forms')],
};
