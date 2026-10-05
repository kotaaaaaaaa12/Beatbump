export interface MediaRequestTiming {
 startDelayMs: number;
 firstByteMs: number;
 transferMs: number;
 upstreamMs?: number;
 transport?: "companion" | "direct";
}

/** Summarize completed requests without exposing signed URLs or headers. */
export function summarizeMediaRequests(entries: readonly PerformanceResourceTiming[], source: string, sourceAt: number): MediaRequestTiming[] {
 return entries.filter(entry => entry.name === source && entry.startTime >= sourceAt - 10 && entry.responseStart > 0)
  .sort((a, b) => a.startTime - b.startTime).slice(-12).map(entry => {
   const result: MediaRequestTiming = {
    startDelayMs: Math.max(0, Math.round(entry.startTime - sourceAt)),
    firstByteMs: Math.max(0, Math.round(entry.responseStart - entry.startTime)),
    transferMs: Math.max(0, Math.round(entry.responseEnd - entry.responseStart)),
   };
   const upstream = entry.serverTiming?.find(timing => ["media_companion", "media_direct"].includes(timing.name));
   if (upstream && Number.isFinite(upstream.duration) && upstream.duration >= 0) {
    result.upstreamMs = Math.round(upstream.duration);
    result.transport = upstream.name === "media_direct" ? "direct" : "companion";
   }
   return result;
  });
}
