import { get, writable } from "svelte/store";
import { SERVER_DOMAIN } from "../../env";
import ja from "./ja.json";

const inBrowser = typeof window !== "undefined" && typeof document !== "undefined";
export const DISCOVERY_CACHE_KEY = "beatbump-discovery-ja-v1";
const CACHE_AGE = 30 * 24 * 60 * 60 * 1000;
const validTranslation = (value: unknown): value is string =>
    typeof value === "string" && value.length <= 320 && /[\u3040-\u30ff\u3400-\u9fff]/u.test(value) && !/[<>\u0000-\u001f]/u.test(value);
const own = (record: Record<string, unknown>, key: string) => Object.prototype.hasOwnProperty.call(record, key) ? record[key] : undefined;
type Cached = { text: string; savedAt: number };
const cache = new Map<string, Cached>();

function readCache(): Record<string, string> {
    const values: Record<string, string> = Object.create(null);
    if (!inBrowser) return values;
    try {
        const raw = localStorage.getItem(DISCOVERY_CACHE_KEY);
        if (!raw || raw.length > 256000) return values;
        const entries = JSON.parse(raw);
        if (!Array.isArray(entries)) return values;
        for (const entry of entries.slice(-512)) {
            if (!Array.isArray(entry) || entry.length !== 2) continue;
            const [key, value] = entry;
            if (typeof key !== "string" || key.length > 144 || !validTranslation(value?.text) ||
                !Number.isFinite(value?.savedAt) || value.savedAt > Date.now() || Date.now() - value.savedAt > CACHE_AGE) continue;
            cache.set(key, value);
            values[key] = value.text;
        }
    } catch { /* Private browsing and invalid cached data must not block music. */ }
    return values;
}

export const discoveryTranslations = writable<Record<string, string>>(readCache());
const waiting = new Set<string>();
const queued = new Set<string>();
const attempts = new Map<string, number>();
const retryAt = new Map<string, number>();
const suppressedUntil = new Map<string, number>();
let timer: ReturnType<typeof setTimeout> | undefined;
let running = false;

// Callers restrict this function to headings and official curated playlists.
export function remoteDiscoveryTranslation(text: string): string | undefined {
    const translated = own(get(discoveryTranslations), text);
    if (typeof translated === "string") return translated;
    if (!inBrowser || !/[a-z]/i.test(text) || /[\u3040-\u30ff\u3400-\u9fff]/u.test(text) || text.length > 144 || /[<>\u0000-\u001f]/u.test(text)) return undefined;
    if ((suppressedUntil.get(text) ?? 0) > Date.now()) return undefined;
    if (retryAt.has(text)) return ja["Translation pending…"];
    if (!waiting.has(text)) {
        attempts.delete(text);
        waiting.add(text);
        queued.add(text);
        schedule();
    }
    return ja["Translating…"];
}

function schedule(): void {
    if (running) return;
    if (timer) clearTimeout(timer);
    const next = queued.size ? Date.now() + 80 : Math.min(...retryAt.values());
    if (!Number.isFinite(next)) { timer = undefined; return; }
    timer = setTimeout(flush, Math.max(0, next - Date.now()));
}

function failed(text: string, retryable: boolean, delay: number): void {
    if (retryable && (attempts.get(text) ?? 0) < 3) {
        retryAt.set(text, Date.now() + delay);
    } else {
        retryAt.delete(text);
        // Avoid repeatedly spending inference budget on an unavailable label.
        suppressedUntil.set(text, Date.now() + 5 * 60 * 1000);
    }
}

function saveCache(translations: Record<string, string>): void {
    for (const [key, text] of Object.entries(translations)) {
        cache.delete(key);
        cache.set(key, { text, savedAt: Date.now() });
    }
    while (cache.size > 512) {
        const oldest = cache.keys().next().value;
        if (typeof oldest !== "string") break;
        cache.delete(oldest);
    }
    try { localStorage.setItem(DISCOVERY_CACHE_KEY, JSON.stringify([...cache])); }
    catch { /* The shared server cache remains available when device storage is denied. */ }
}

async function flush(): Promise<void> {
    timer = undefined;
    if (running) return;
    for (const [text, due] of retryAt) {
        if (due > Date.now()) continue;
        retryAt.delete(text);
        waiting.add(text);
        queued.add(text);
    }
    if (!queued.size) { schedule(); return; }
    running = true;
    const labels: string[] = [];
    let size = 0;
    for (const text of queued) {
        if (labels.length === 16 || size + text.length > 1536) break;
        labels.push(text);
        queued.delete(text);
        attempts.set(text, (attempts.get(text) ?? 0) + 1);
        size += text.length;
    }
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 25000);
    const maxAttempt = Math.max(...labels.map(text => attempts.get(text) ?? 1));
    let retryDelay = 5000 * 2 ** (maxAttempt - 1);
    try {
        const response = await fetch(`${SERVER_DOMAIN}/api/v1/localize`, {
            method: "POST", headers: { "Content-Type": "application/json" },
            credentials: "same-origin", signal: controller.signal,
            body: JSON.stringify({ locale: "ja", labels }),
        });
        const suggested = Number(response.headers.get("Retry-After"));
        if (Number.isFinite(suggested) && suggested > 0) retryDelay = Math.max(retryDelay, Math.min(60000, suggested * 1000));
        if (!response.ok) {
            labels.forEach(text => failed(text, response.status === 429 || response.status >= 500, retryDelay));
        } else {
            const result = await response.json();
            const translations: Record<string, string> = Object.create(null);
            const retryable = new Set(Array.isArray(result.retryable) ? result.retryable.filter((value: unknown) => typeof value === "string") : []);
            for (const text of labels) {
                const value = result.translations && own(result.translations, text);
                if (validTranslation(value)) {
                    translations[text] = value;
                    attempts.delete(text);
                    suppressedUntil.delete(text);
                } else failed(text, retryable.has(text), retryDelay);
            }
            saveCache(translations);
            discoveryTranslations.update(current => ({ ...current, ...translations }));
        }
    } catch {
        labels.forEach(text => failed(text, true, retryDelay));
    } finally {
        clearTimeout(timeout);
        labels.forEach(text => waiting.delete(text));
        running = false;
        // Publish retry/fallback status without storing failures on the device.
        discoveryTranslations.update(current => ({ ...current }));
        schedule();
    }
}
