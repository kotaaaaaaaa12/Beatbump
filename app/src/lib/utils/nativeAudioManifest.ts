/** Native HLS is used only for the same site's matching signed audio ticket. */
export function nativeAudioManifest(file: string, manifest: string | undefined, supported: boolean, origin: string): string | undefined {
 if (!supported || !manifest) return;
 try {
  const audio = new URL(file, origin);
  const playlist = new URL(manifest, origin);
  if (audio.origin !== origin || playlist.origin !== origin || audio.pathname !== "/api/v1/media" ||
   playlist.pathname !== "/api/v1/audio.m3u8" || !audio.searchParams.get("ticket") ||
   audio.searchParams.get("ticket") !== playlist.searchParams.get("ticket")) return;
  return playlist.href;
 } catch { return; }
}
