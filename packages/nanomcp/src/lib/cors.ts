import type { Middleware } from "@remix-run/fetch-router";

export interface CorsOptions {
	/**
	 * Configures the `Access-Control-Allow-Origin` header.
	 * - `string`: Sets a specific origin (e.g., "https://example.com")
	 * - `string[]`: Array of allowed origins
	 * - `*`: Allows all origins
	 * - `function`: Dynamic origin validation function
	 * Defaults to `*`.
	 */
	origin?: string | string[] | "*" | ((origin: string) => boolean);

	/**
	 * Configures the `Access-Control-Allow-Methods` header.
	 * Defaults to `["GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"]`.
	 */
	methods?: string[];

	/**
	 * Configures the `Access-Control-Allow-Headers` header.
	 * Defaults to the value from the request's `Access-Control-Request-Headers` header.
	 */
	allowedHeaders?: string[];

	/**
	 * Configures the `Access-Control-Expose-Headers` header.
	 */
	exposedHeaders?: string[];

	/**
	 * Configures the `Access-Control-Allow-Credentials` header.
	 * Set to `true` to pass the header, otherwise omitted.
	 * Defaults to `false`.
	 */
	credentials?: boolean;

	/**
	 * Configures the `Access-Control-Max-Age` header in seconds.
	 * Defaults to `86400` (24 hours).
	 */
	maxAge?: number;

	/**
	 * Whether to pass the CORS preflight response to the next handler.
	 * Set to `true` to pass OPTIONS requests to the next handler.
	 * Defaults to `false` (handles OPTIONS immediately).
	 */
	preflightContinue?: boolean;

	/**
	 * Provides a status code to use for successful OPTIONS requests.
	 * Defaults to `204`.
	 */
	optionsSuccessStatus?: number;
}

/**
 * Creates a CORS middleware handler that adds appropriate CORS headers to responses.
 *
 * @example
 * ```typescript
 * const router = createRouter({
 *   middleware: [
 *     cors({
 *       origin: "https://example.com",
 *       credentials: true,
 *     }),
 *   ],
 * });
 * ```
 */
export function cors(options: CorsOptions = {}): Middleware {
	const {
		origin = "*",
		methods = ["GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"],
		allowedHeaders,
		exposedHeaders,
		credentials = false,
		maxAge = 86400,
		preflightContinue = false,
		optionsSuccessStatus = 204,
	} = options;

	return async (context, next) => {
		const requestOrigin =
			context.request.headers.get("Origin") ||
			context.request.headers.get("origin") ||
			"";

		// Determine if the origin is allowed
		let allowedOrigin: string | null = null;

		if (origin === "*") {
			allowedOrigin = "*";
		} else if (typeof origin === "string") {
			allowedOrigin = origin;
		} else if (Array.isArray(origin)) {
			if (origin.includes(requestOrigin)) {
				allowedOrigin = requestOrigin;
			}
		} else if (typeof origin === "function") {
			if (origin(requestOrigin)) {
				allowedOrigin = requestOrigin;
			}
		}

		// Handle preflight OPTIONS request
		if (context.method === "OPTIONS") {
			const headers = new Headers();

			if (allowedOrigin) {
				headers.set("Access-Control-Allow-Origin", allowedOrigin);
			}

			if (credentials) {
				headers.set("Access-Control-Allow-Credentials", "true");
			}

			// Use requested headers if not specified
			const requestedHeaders = context.request.headers.get(
				"Access-Control-Request-Headers",
			);
			if (allowedHeaders) {
				headers.set("Access-Control-Allow-Headers", allowedHeaders.join(", "));
			} else if (requestedHeaders) {
				headers.set("Access-Control-Allow-Headers", requestedHeaders);
			}

			headers.set("Access-Control-Allow-Methods", methods.join(", "));
			headers.set("Access-Control-Max-Age", maxAge.toString());

			if (exposedHeaders) {
				headers.set("Access-Control-Expose-Headers", exposedHeaders.join(", "));
			}

			// Vary header to indicate response varies based on Origin
			headers.set("Vary", "Origin");

			if (preflightContinue) {
				// Pass preflight to next handler
				const response = await next();
				// Merge CORS headers with response headers
				for (const [key, value] of headers.entries()) {
					response.headers.set(key, value);
				}
				return response;
			}

			// Return preflight response immediately
			return new Response(null, {
				status: optionsSuccessStatus,
				headers,
			});
		}

		// For non-OPTIONS requests, add CORS headers to response
		const response = await next();

		if (allowedOrigin) {
			response.headers.set("Access-Control-Allow-Origin", allowedOrigin);
		}

		if (credentials) {
			response.headers.set("Access-Control-Allow-Credentials", "true");
		}

		if (exposedHeaders) {
			response.headers.set(
				"Access-Control-Expose-Headers",
				exposedHeaders.join(", "),
			);
		}

		// Vary header to indicate response varies based on Origin
		response.headers.set("Vary", "Origin");

		return response;
	};
}
