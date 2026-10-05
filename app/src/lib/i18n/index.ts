import { derived, get, writable } from "svelte/store";
import en from "./en.json";
import ja from "./ja.json";

export type LanguagePreference = "auto" | "ja" | "en";
export type Locale = "ja" | "en";
export type MessageId = keyof typeof en;
export type Parameters = Record<string, string | number>;
export const LANGUAGE_KEY = "beatbump-language";
const inBrowser = typeof window !== "undefined" && typeof document !== "undefined";
const validPreference = (value: unknown): value is LanguagePreference =>
	value === "auto" || value === "ja" || value === "en";

export function browserLocale(languages: readonly string[]): Locale {
	// The primary preferred language decides; secondary Japanese does not override English.
	return /^ja(?:-|$)/i.test(languages[0] ?? "") ? "ja" : "en";
}

function readPreference(): LanguagePreference {
	try {
		const value = inBrowser ? localStorage.getItem(LANGUAGE_KEY) : null;
		return validPreference(value) ? value : "auto";
	} catch { return "auto"; }
}
const preference = writable<LanguagePreference>(readPreference());
const detected = writable<Locale>(inBrowser ? browserLocale(navigator.languages?.length ? navigator.languages : [navigator.language]) : "en");
export const languagePreference = { subscribe: preference.subscribe };
export const locale = derived([preference, detected], ([choice, automatic]) => choice === "auto" ? automatic : choice);

export function setLanguage(value: LanguagePreference): void {
	if (!validPreference(value)) return;
	preference.set(value);
	if (inBrowser) {
		try { localStorage.setItem(LANGUAGE_KEY, value); } catch { /* Private browsing may deny storage. */ }
	}
}

export function translateFor(language: Locale, message: string, params: Parameters = {}): string {
	const catalog: Record<string, string> = language === "ja" ? ja : en;
	if (language === "en" && message === "{count} songs" && Number(params.count) === 1) message = "{count} song";
	const template = Object.prototype.hasOwnProperty.call(catalog, message) ? catalog[message] : message;
	return template.replace(/\{(\w+)\}/g, (match, key: string) => params[key] === undefined ? match : String(params[key]));
}
export const t = derived(locale, language => (message: string | undefined, params: Parameters = {}) => translateFor(language, message ?? "", params));
export function translateMetadataFor(language: Locale, value: unknown): string {
	const text = String(value ?? "");
	if (language === "en") return text;
	const units: Record<string, string> = { song: "songs", track: "tracks", view: "views", subscriber: "subscribers", play: "plays", hour: "hours", minute: "minutes", min: "minutes", second: "seconds", sec: "seconds" };
	return text.split(/(\s*[•·]\s*)/).map(part => translateFor(language, part)).join("").replace(/([\d,.]+[KMB]?)\s+(songs?|tracks?|views?|subscribers?|plays?|hours?|minutes?|mins?|seconds?|secs?)\b/gi, (_, count: string, unit: string) => {
		const singular = unit.toLowerCase().replace(/s$/, "");
		return translateFor(language, "{count} " + units[singular], { count });
	});
}
// Only structural metadata uses this formatter; song/artist names never do.
export const metadata = derived(locale, language => (value: unknown) => translateMetadataFor(language, value));
export const translate = (message: string, params: Parameters = {}) => translateFor(get(locale), message, params);

// Worker notifications arrive on the main thread as English messages. Match catalog
// templates without changing the worker's protocol or persisting translated data.
const messageTemplates = Object.keys(en).filter(key => /\{\w+\}/.test(key)).map(key => {
	const names: string[] = [];
	const escaped = key.split(/(\{\w+\})/).map(part => {
		if (/^\{\w+\}$/.test(part)) { names.push(part.slice(1, -1)); return "(.+?)"; }
		return part.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
	}).join("");
	return { key, names, pattern: new RegExp("^" + escaped + "$") };
});
export function translateMessage(message: unknown): string {
	const text = String(message ?? "");
	if (Object.prototype.hasOwnProperty.call(en, text)) return translate(text);
	for (const template of messageTemplates) {
		const match = template.pattern.exec(text);
		if (match) return translate(template.key, Object.fromEntries(template.names.map((name, index) => [name, name === "reason" ? errorMessage(match[index + 1]) : match[index + 1]])));
	}
	return text;
}
export const message = derived(locale, () => (value: unknown, isError = false) => isError ? errorMessage(value) : translateMessage(value));

export function errorMessage(error: unknown): string {
	const text = error instanceof Error ? error.message : String(error ?? "");
	const localized = translateMessage(text);
	if (localized !== text || Object.prototype.hasOwnProperty.call(en, text)) return localized;
	return translate(/network|fetch|connection|timeout/i.test(text) ? "Network error. Please try again." : "An error occurred. Please try again.");
}

if (inBrowser) {
	preference.subscribe(value => {
		try { document.cookie = LANGUAGE_KEY + "=" + value + "; Path=/; Max-Age=31536000; SameSite=Lax" + (location.protocol === "https:" ? "; Secure" : ""); } catch { /* The startup page also reads localStorage. */ }
	});
	locale.subscribe(language => { document.documentElement.lang = language; });
	window.addEventListener("languagechange", () => {
		detected.set(browserLocale(navigator.languages?.length ? navigator.languages : [navigator.language]));
	});
	window.addEventListener("storage", event => {
		if (event.key === LANGUAGE_KEY || event.key === null) preference.set(readPreference());
	});
}
