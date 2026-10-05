import { writable } from "svelte/store";

// Compatibility exports: do not read old machine translations or contact a service.
export const DISCOVERY_CACHE_KEY = "beatbump-discovery-ja-v1";
export const discoveryTranslations = writable<Record<string, string>>({});
export function remoteDiscoveryTranslation(_text: string): undefined {
    return undefined;
}
