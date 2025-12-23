export interface Resource extends BaseMetadata, Icons {
	uri: string;
	description?: string;
	mimeType?: string;
	annotations?: {
		audience?: string[];
		priority?: number;
		lastModified?: string;
	};
	size?: number;
	_meta?: { [key: string]: unknown };
}

export interface Resources {
	resources: Resource[];
}

export interface BaseMetadata {
	name: string;
	title?: string;
}

export interface ResourceContents {
	uri: string;
	mimeType?: string;
	text?: string;
	blob?: string;
	_meta?: { [key: string]: unknown };
}

export interface InitializationResult {
	capabilities?: {
		experimental?: {
			"ai.nanobot"?: {
				session?: {
					ui?: boolean;
					workspace?: {
						id?: string;
						baseUri?: string;
						base?: string;
						supported?: boolean;
					};
				};
			};
		};
	};
}

export interface Icons {
	/**
	 * Optional set of sized icons that the client can display in a user interface.
	 *
	 * Clients that support rendering icons MUST support at least the following MIME types:
	 * - `image/png` - PNG images (safe, universal compatibility)
	 * - `image/jpeg` (and `image/jpg`) - JPEG images (safe, universal compatibility)
	 *
	 * Clients that support rendering icons SHOULD also support:
	 * - `image/svg+xml` - SVG images (scalable but requires security precautions)
	 * - `image/webp` - WebP images (modern, efficient format)
	 */
	icons?: Icon[];
}

export interface Icon {
	/**
	 * A standard URI pointing to an icon resource. May be an HTTP/HTTPS URL or a
	 * `data:` URI with Base64-encoded image data.
	 *
	 * Consumers SHOULD takes steps to ensure URLs serving icons are from the
	 * same domain as the client/server or a trusted domain.
	 *
	 * Consumers SHOULD take appropriate precautions when consuming SVGs as they can contain
	 * executable JavaScript.
	 *
	 * @format uri
	 */
	src: string;

	/**
	 * Optional MIME type override if the source MIME type is missing or generic.
	 * For example: `"image/png"`, `"image/jpeg"`, or `"image/svg+xml"`.
	 */
	mimeType?: string;

	/**
	 * Optional array of strings that specify sizes at which the icon can be used.
	 * Each string should be in WxH format (e.g., `"48x48"`, `"96x96"`) or `"any"` for scalable formats like SVG.
	 *
	 * If not provided, the client should assume that the icon can be used at any size.
	 */
	sizes?: string[];

	/**
	 * Optional specifier for the theme this icon is designed for. `light` indicates
	 * the icon is designed to be used with a light background, and `dark` indicates
	 * the icon is designed to be used with a dark background.
	 *
	 * If not provided, the client should assume the icon can be used with any theme.
	 */
	theme?: "light" | "dark";
}

export const UIPath = "/mcp?ui";
