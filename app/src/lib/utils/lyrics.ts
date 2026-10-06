export interface LyricLine { time: number; text: string }
export interface LyricsRecord {
 id: number; trackName: string; artistName: string; albumName: string;
 duration: number; instrumental: boolean; plainLyrics: string; syncedLyrics: string;
}
export interface LyricsResponse { status: "found" | "not_found" | "unavailable" | "invalid_track"; source?: string; record?: LyricsRecord }

// Keep timing tags out of the displayed text; never render provider text as HTML.
export function parseSyncedLyrics(value: string): LyricLine[] {
 const lines: LyricLine[] = [];
 const offset = Number(/\[offset:([+-]?\d+)\]/i.exec(value)?.[1] ?? 0) / 1000;
 for (const raw of value.split(/\r?\n/).slice(0, 2000)) {
  const tags = [...raw.matchAll(/\[(\d{1,3}):([0-5]\d)(?:[.:](\d{1,3}))?\]/g)];
  const text = raw.replace(/\[(\d{1,3}):([0-5]\d)(?:[.:](\d{1,3}))?\]/g, "").replace(/<\d{1,3}:[0-5]\d(?:\.\d{1,3})?>/g, "").trim();
  for (const tag of tags) {
   const time = Number(tag[1])*60 + Number(tag[2]) + Number(`0.${tag[3] ?? "0"}`) + offset;
   if (Number.isFinite(time) && time <= 3600) lines.push({ time: Math.max(0,time), text });
  }
 }
 return lines.sort((a,b) => a.time-b.time);
}
export function activeLyricIndex(lines: readonly LyricLine[], position: number): number {
 let low=0, high=lines.length;
 while (low<high) { const mid=(low+high)>>>1; if (lines[mid].time<=position) low=mid+1; else high=mid; }
 return low-1;
}
