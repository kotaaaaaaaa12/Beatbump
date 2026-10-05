import { SERVER_DOMAIN } from "../env";
import { get } from "svelte/store";
import { locale, type Locale } from "$lib/i18n";

export function localizedApiURL(path: string, language: Locale = get(locale)): string {
    if (!path.startsWith("/api/v1/")) return `${SERVER_DOMAIN}${path}`;
    const address = new URL(path, "https://beatbump.invalid");
    address.searchParams.set("lang", language);
    return `${SERVER_DOMAIN}${address.pathname}${address.search}${address.hash}`;
}

export const APIClient = {
    fetch: (url: string): Promise<any> => {
        const headers: Record<string, string> = {}

        // add the headers to the options
        let uri = localizedApiURL(url);
        return fetch(uri, { headers: headers, credentials: 'same-origin' })
    },
    post: (url: string, body?: any): Promise<any> => {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json'
        }

        let uri = localizedApiURL(url);
        return fetch(uri, {
            method: 'POST',
            headers: headers,
            credentials: 'same-origin',
            body: JSON.stringify(body)
        })
    },
    del: (url: string): Promise<any> => {
        const headers: Record<string, string> = {}
        let uri = localizedApiURL(url);
        return fetch(uri, {
            method: 'DELETE',
            headers: headers,
            credentials: 'same-origin'
        })
    }
};
