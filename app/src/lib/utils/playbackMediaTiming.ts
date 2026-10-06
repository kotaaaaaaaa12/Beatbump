export interface MediaRequestTiming {
 startDelayMs: number;
 firstByteMs: number;
 transferMs: number;
 upstreamMs?: number;
 transport?: "companion" | "direct" | "cache";
 responseStatus?: number;
 responseBytes?: number;
 rangeStart?: number;
 rangeEnd?: number;
 totalBytes?: number;
 method?: "GET" | "HEAD";
}

/** Summarize completed requests without exposing signed URLs or headers. */
export function summarizeMediaRequests(entries: readonly PerformanceResourceTiming[], source: string, sourceAt: number): MediaRequestTiming[] {
 const audioSource = source.includes("/api/v1/audio.m3u8?") ? source.replace("/api/v1/audio.m3u8?", "/api/v1/media?") : source;
 const playlistSource = source.includes("/api/v1/media?") ? source.replace("/api/v1/media?", "/api/v1/audio.m3u8?") : source;
 return entries.filter(entry => (entry.name === source || entry.name === audioSource || entry.name === playlistSource) && entry.startTime >= sourceAt - 10 && entry.responseStart > 0)
  .sort((a, b) => a.startTime - b.startTime).slice(-12).map(entry => {
   const result: MediaRequestTiming = {
    startDelayMs: Math.max(0, Math.round(entry.startTime - sourceAt)),
    firstByteMs: Math.max(0, Math.round(entry.responseStart - entry.startTime)),
    transferMs: Math.max(0, Math.round(entry.responseEnd - entry.responseStart)),
   };
   const upstream = entry.serverTiming?.find(timing => ["media_companion", "media_direct", "media_cache"].includes(timing.name));
   if (upstream && Number.isFinite(upstream.duration) && upstream.duration >= 0) {
    result.upstreamMs = Math.round(upstream.duration);
    result.transport = upstream.name === "media_cache" ? "cache" : upstream.name === "media_direct" ? "direct" : "companion";
   }
   const fields = { media_status: "responseStatus", media_bytes: "responseBytes", media_range_start: "rangeStart", media_range_end: "rangeEnd", media_total_bytes: "totalBytes" } as const;
   for (const [name, field] of Object.entries(fields)) {
    const metric = entry.serverTiming?.find(timing => timing.name === name);
    if (metric && Number.isSafeInteger(metric.duration) && metric.duration >= 0) {
     result[field as typeof fields[keyof typeof fields]] = metric.duration;
    }
   }
   if (result.responseStatus !== undefined) {
    result.method = entry.serverTiming?.some(timing => timing.name === "media_head" && timing.duration === 1) ? "HEAD" : "GET";
   }
   return result;
  });
}

export interface PlaybackMediaEvent {
 event: string;
 atMs: number;
 readyState: number;
 networkState: number;
 paused: boolean;
 positionMs: number;
 bufferedAheadMs: number;
 errorCode?: number;
}

/** Keep only numeric media state; never include source URLs or error messages. */
export function captureMediaEvent(audio: HTMLMediaElement, event: string, atMs: number): PlaybackMediaEvent {
 let bufferedAheadMs = 0;
 const position = Number.isFinite(audio.currentTime) ? audio.currentTime : 0;
 for (let index = 0; index < audio.buffered.length; index++) {
  if (audio.buffered.start(index) <= position && audio.buffered.end(index) >= position) {
   bufferedAheadMs = Math.round((audio.buffered.end(index) - position) * 1000);
   break;
  }
 }
 return { event, atMs: Math.max(0, Math.round(atMs)), readyState: audio.readyState,
  networkState: audio.networkState, paused: audio.paused, positionMs: Math.round(position * 1000),
  bufferedAheadMs, ...(audio.error ? { errorCode: audio.error.code } : {}) };
}
