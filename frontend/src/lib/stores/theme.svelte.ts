const STORAGE_KEY = 'cp_theme';

class ThemeStore {
	dark = $state(false);

	init() {
		const saved = localStorage.getItem(STORAGE_KEY);
		this.dark = saved ? saved === 'dark' : window.matchMedia('(prefers-color-scheme: dark)').matches;
		this._apply();
	}

	toggle() {
		this.dark = !this.dark;
		localStorage.setItem(STORAGE_KEY, this.dark ? 'dark' : 'light');
		this._apply();
	}

	private _apply() {
		document.documentElement.setAttribute('data-theme', this.dark ? 'dark' : 'light');
	}
}

export const theme = new ThemeStore();
