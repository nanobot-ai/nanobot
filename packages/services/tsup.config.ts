import { defineConfig } from "tsup";

export default defineConfig({
	entry: ["server.ts"],
	format: ["esm"],
	dts: true,
	clean: true,
	// minify: true,
	sourcemap: true,
	splitting: false,
	treeshake: true,
	// noExternal: [/.*/],
	outDir: "dist",
	target: "esnext",
	platform: "node",
});
