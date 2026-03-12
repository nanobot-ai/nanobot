declare module '@novnc/novnc/lib/rfb.js' {
	interface RFBSecurityFailureDetail {
		status: number | string;
	}

	interface RFBEventMap {
		connect: Event;
		disconnect: Event;
		credentialsrequired: Event;
		clipboard: CustomEvent<{ text: string }>;
		securityfailure: CustomEvent<RFBSecurityFailureDetail>;
	}

	class RFB extends EventTarget {
		constructor(target: Element, url: string);

		scaleViewport: boolean;
		resizeSession: boolean;
		dragViewport: boolean;
		clipViewport: boolean;

		focus(options?: FocusOptions): void;
		clipboardPasteFrom(text: string): void;
		sendKey(keysym: number, code: string | null, down?: boolean): void;
		disconnect(): void;

		addEventListener<K extends keyof RFBEventMap>(
			type: K,
			listener: (this: RFB, event: RFBEventMap[K]) => void,
			options?: boolean | AddEventListenerOptions
		): void;
	}

	export default RFB;
}
