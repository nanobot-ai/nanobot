import { browser } from '$app/environment';

const STORAGE_KEY = 'sidebar-collapsed';

export class SidebarStore {
	isCollapsed = $state(false);
	isMobileOpen = $state(false);

	constructor() {
		// Load sidebar state from localStorage on initialization (desktop only)
		if (browser && window.innerWidth >= 1024) {
			const saved = localStorage.getItem(STORAGE_KEY);
			if (saved !== null) {
				this.isCollapsed = JSON.parse(saved);
			}
		}
	}

	toggle(): void {
		if (browser && window.innerWidth >= 1024) {
			this.isCollapsed = !this.isCollapsed;
			localStorage.setItem(STORAGE_KEY, JSON.stringify(this.isCollapsed));
		}
	}

	collapse(): void {
		if (browser && window.innerWidth >= 1024) {
			this.isCollapsed = true;
			localStorage.setItem(STORAGE_KEY, JSON.stringify(this.isCollapsed));
		}
	}

	expand(): void {
		if (browser && window.innerWidth >= 1024) {
			this.isCollapsed = false;
			localStorage.setItem(STORAGE_KEY, JSON.stringify(this.isCollapsed));
		}
	}

	toggleMobile(): void {
		this.isMobileOpen = !this.isMobileOpen;
	}

	openMobile(): void {
		this.isMobileOpen = true;
	}

	closeMobile(): void {
		this.isMobileOpen = false;
	}
}
