import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        background: "var(--background)",
        foreground: "var(--foreground)",
        mint: {
          400: "#34d399",
          500: "#10b981",
          600: "#059669",
        },
        sand: {
          50: "#faf9f7",
          100: "#f3f1ed",
          200: "#e8e4dd",
          300: "#d5cfc4",
          400: "#b0a896",
          500: "#8c8272",
          600: "#6b6356",
          700: "#504a40",
          800: "#363229",
          900: "#1c1a16",
          950: "#0e0d0b",
        },
      },
      fontFamily: {
        mono: [
          "GeistMono",
          "ui-monospace",
          "SFMono-Regular",
          "Menlo",
          "monospace",
        ],
      },
    },
  },
  plugins: [],
};
export default config;
