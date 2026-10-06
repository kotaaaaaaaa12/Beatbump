<script lang="ts">
 import { onDestroy, tick } from "svelte";
 import { t } from "$lib/i18n";
 import { AudioPlayer } from "$lib/player";
 import type { Song } from "$lib/types";
 import { activeLyricIndex, parseSyncedLyrics, type LyricsResponse } from "$lib/utils/lyrics";
 export let track: Song | undefined;
 export let position = 0;
 export let duration = 0;
 export let enabled = false;
 let result: LyricsResponse | undefined;
 let loading = false;
 let requestKey = "";
 let controller: AbortController | undefined;
 let scroller: HTMLDivElement;
 let userScrollUntil = 0;
 let lastScroll = -2;
 let generation = 0;
 $: artist = track?.artistInfo?.artist?.find(item => item.text)?.text ?? "";
 $: key = enabled && track?.videoId && duration > 0 ? `${track.videoId}:${Math.round(duration)}` : "";
 $: if (key !== requestKey) { requestKey = key; void load(); }
 $: lines = parseSyncedLyrics(result?.record?.syncedLyrics ?? "");
 $: active = activeLyricIndex(lines, position);
 $: if (enabled && active !== lastScroll && lines.length) void followLine(active);
 async function load() {
  controller?.abort(); const run = ++generation;
  result = undefined; lastScroll = -2; userScrollUntil = 0;
  if (!requestKey || !track) { loading = false; return; }
  if (!artist) { result = { status: "invalid_track" }; loading = false; return; }
  controller = new AbortController(); loading = true;
  const query = new URLSearchParams({title: track.title, artist, duration: String(duration)});
  if (track.album?.title) query.set("album", track.album.title);
  try {
   const response = await fetch(`/api/v1/lyrics.json?${query}`, {signal: controller.signal});
   const value: LyricsResponse = await response.json();
   if (run !== generation) return;
   if (!response.ok || !["found","not_found"].includes(value.status) || (value.status === "found" && !value.record)) throw new Error("Lyrics unavailable");
   result = value;
  } catch (error) {
   if (run === generation && !(error instanceof DOMException && error.name === "AbortError")) result = {status: "unavailable"};
  } finally { if (run === generation) loading = false; }
 }
 async function followLine(index: number) {
  if (Date.now() < userScrollUntil || index < 0) return;
  lastScroll = index; await tick();
  const element = scroller?.querySelector<HTMLElement>(`[data-line="${index}"]`);
  if (!element) return;
  const top = scroller.scrollTop + element.getBoundingClientRect().top - scroller.getBoundingClientRect().top - scroller.clientHeight/3;
  scroller.scrollTo({top: Math.max(0,top), behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth"});
 }
 function seek(time: number) { userScrollUntil = 0; lastScroll = -2; AudioPlayer.seek(time); }
 onDestroy(() => { ++generation; controller?.abort(); });
</script>

<section class="lyrics" aria-label={$t("Lyrics")}>
 <header><h2>{$t("Lyrics")}</h2>{#if lines.length}<span>{$t("Tap a line to seek")}</span>{/if}</header>
 <div class="lyrics-scroll" bind:this={scroller} on:touchstart|stopPropagation={() => userScrollUntil = Date.now()+8000} on:wheel={() => userScrollUntil = Date.now()+8000}>
  {#if loading || (enabled && !key)}
   <p class="message" role="status">{$t("Loading lyrics…")}</p>
  {:else if result?.status === "unavailable"}
   <div class="message" role="status"><p>{$t("Could not connect to the lyrics service.")}</p><button class="retry" on:click={load}>{$t("Retry")}</button></div>
  {:else if result?.status === "not_found" || result?.status === "invalid_track"}
   <p class="message">{$t("Lyrics are not available for this track.")}</p>
  {:else if result?.record?.instrumental}
   <p class="message">{$t("Instrumental track")}</p>
  {:else if lines.length}
   <div class="lines">{#each lines as line, index}<button data-line={index} class:current={index === active} class:past={index < active} aria-current={index === active ? "true" : undefined} aria-label={$t("Seek to {time}", {time: `${Math.floor(line.time/60)}:${String(Math.floor(line.time%60)).padStart(2,"0")}`})} on:click={() => seek(line.time)}>{line.text || "♪"}</button>{/each}</div>
  {:else if result?.record?.plainLyrics}
   <p class="plain">{result.record.plainLyrics}</p>
  {/if}
 </div>
 {#if result?.record}<footer>{$t("Lyrics source")}: <a href={`https://lrclib.net/api/get/${result.record.id}`} target="_blank" rel="noopener noreferrer">LRCLIB</a></footer>{/if}
</section>

<style>
 .lyrics { display:flex; flex-direction:column; width:100%; height:100%; min-height:0; text-align:left; border-radius:16px; background:rgba(20,23,30,.94); color:#fff; font-family:system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif; overflow:hidden; }
 header { display:flex; align-items:center; justify-content:space-between; gap:12px; padding:16px 20px 10px; flex-wrap:wrap; }
 h2 { color:#fff; font-family:inherit; font-size:1.1rem; margin:0; } header span, footer { color:#c3c7d0; font-size:.75rem; }
 .lyrics-scroll { flex:1; min-height:0; overflow-y:auto; overscroll-behavior:contain; touch-action:pan-y; padding:12px 20px 32px; scroll-behavior:smooth; }
 .lines { display:flex; flex-direction:column; gap:8px; padding-bottom:35%; }
 .lines button { appearance:none; border:0; background:transparent !important; color:#999fab !important; box-shadow:none !important; white-space:normal; text-transform:none; justify-content:flex-start; padding:8px 0; font:inherit; font-size:clamp(1.15rem,2.3vw,2rem); font-weight:700; text-align:left; line-height:1.5; overflow-wrap:anywhere; cursor:pointer; transition:color .18s; }
 .lines button.current { color:#fff !important; } .lines button.past { color:#c0c5ce !important; }
 button:focus-visible { outline:2px solid #b9c8ff; outline-offset:3px; border-radius:4px; }
 .plain { white-space:pre-wrap; font-size:1.15rem; line-height:1.8; overflow-wrap:anywhere; margin:0; }
 .message { color:#c3c7d0; line-height:1.6; }
 .retry { background:#fff !important; color:#18202e !important; border:0; border-radius:20px; padding:10px 20px; font:inherit; cursor:pointer; }
 footer { padding:10px 20px; } footer a { color:inherit; text-decoration:underline; }
 @media (prefers-reduced-motion: reduce) { .lyrics-scroll {scroll-behavior:auto;} .lines button {transition:none;} }
</style>
