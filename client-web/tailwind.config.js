/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: "#E9D7FD",
        secondary: "#4E1164",
        tertiary: "#400e52",
        success: "#4F8A10",
        "success-bg": "#DFF2BF",
        error: "#D8000C",
        "error-bg": "#FFBABA",
      },
    },
    screens: {
      xs: "435px",
      sm: "640px",
    },
  },
  plugins: [],
};
