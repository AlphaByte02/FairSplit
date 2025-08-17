import { defineConfig } from "unocss";
import presetWind4 from "@unocss/preset-wind4";

export default defineConfig({
    cli: {
        entry: {
            patterns: ["./web/**/*.templ"],
            // outFile: "./web/assets/css/uno.css",
        },
    },
    presets: [
        presetWind4({
            preflights: {
                reset: true,
            },
        }),
    ],
});
