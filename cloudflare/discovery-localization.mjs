const MODEL = '@cf/meta/m2m100-1.2b';
const MAX_ITEMS = 16;
const DAILY_LIMIT = 512;
const japanese = /[\u3040-\u30ff\u3400-\u9fff]/u;
const cacheHeaders = { 'Cache-Control': 'no-store', 'X-Content-Type-Options': 'nosniff' };

// Only public discovery labels reach this service, never listening history or tokens.
export class DiscoveryLocalizer {
  constructor(storage, ai, context) {
    this.storage = storage;
    this.ai = ai;
    this.pending = new Map();
    this.context = context;
    this.active = 0;
    this.queue = [];
  }

  async handle(request) {
    const url = new URL(request.url);
    if (request.method !== 'POST') return Response.json({ error: 'method_not_allowed' }, { status: 405, headers: cacheHeaders });
    if (request.headers.get('Origin') !== url.origin || request.headers.get('Sec-Fetch-Site') === 'cross-site') {
      return Response.json({ error: 'origin_not_allowed' }, { status: 403, headers: cacheHeaders });
    }
    if (!request.headers.get('Content-Type')?.startsWith('application/json')) {
      return Response.json({ error: 'invalid_content_type' }, { status: 415, headers: cacheHeaders });
    }
    let body;
    try {
      const reader = request.body?.getReader();
      if (!reader) throw new Error('missing_body');
      let bytes = 0;
      const chunks = [];
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        bytes += value.byteLength;
        if (bytes > 8192) { await reader.cancel(); throw new Error('body_too_large'); }
        chunks.push(value);
      }
      const joined = new Uint8Array(bytes);
      let offset = 0;
      for (const chunk of chunks) { joined.set(chunk, offset); offset += chunk.byteLength; }
      body = JSON.parse(new TextDecoder().decode(joined));
    } catch {
      return Response.json({ error: 'invalid_body' }, { status: 400, headers: cacheHeaders });
    }
    if (body?.locale !== 'ja' || !Array.isArray(body.labels) || !body.labels.length || body.labels.length > MAX_ITEMS ||
        body.labels.some(text => typeof text !== 'string' || !text.trim() || text.length > 144 || /[\u0000-\u001f<>]/u.test(text)) ||
        body.labels.reduce((size, text) => size + text.length, 0) > 1536) {
      return Response.json({ error: 'invalid_labels' }, { status: 400, headers: cacheHeaders });
    }
    const translations = Object.create(null);
    const missing = [];
    const retryable = [];
    const outstanding = new Set([...new Set(body.labels)]);
    // Cached labels and unrelated visitors do not wait behind a whole AI batch.
    const tasks = [...outstanding].map(async text => {
      const result = await this.resolve(text);
      if (result.text) translations[text] = result.text;
      else {
        missing.push(text);
        if (result.retryable) retryable.push(text);
      }
      outstanding.delete(text);
    });
    const work = Promise.all(tasks);
    this.context?.waitUntil(work);
    let timer;
    try {
      await Promise.race([work, new Promise(resolve => { timer = setTimeout(resolve, 20000); })]);
    } finally { clearTimeout(timer); }
    // Inferences can finish and populate the shared cache after this response.
    const unfinished = [...outstanding];
    return Response.json({
      translations, missing: [...missing, ...unfinished], retryable: [...retryable, ...unfinished],
    }, { headers: { ...cacheHeaders, ...(missing.length || unfinished.length ? { 'Retry-After': '5' } : {}) } });
  }

  acquire() {
    if (this.active < 4) {
      this.active++;
      return Promise.resolve(() => this.release());
    }
    if (this.queue.length >= 16) return Promise.resolve(null);
    return new Promise(resolve => this.queue.push(resolve));
  }

  release() {
    const next = this.queue.shift();
    if (next) next(() => this.release());
    else this.active--;
  }

  async resolve(text) {
    try { return await this.translate(text); }
    catch { return { retryable: true }; }
  }

  async translate(text) {
    const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(text));
    const key = 'discovery-ja-v1:' + [...new Uint8Array(digest)].map(byte => byte.toString(16).padStart(2, '0')).join('');
    const cached = await this.storage.get(key);
    if (typeof cached === "string" && japanese.test(cached)) return { text: cached };
    if (!this.ai?.run) return { retryable: false };
    if (this.pending.has(key)) return this.pending.get(key);
    const promise = (async () => {
      const release = await this.acquire();
      if (!release) return { retryable: true };
      let timer;
      try {
        const day = new Date().toISOString().slice(0, 10);
        const allowed = await this.storage.transaction(async storage => {
          const prior = await storage.get('discovery-ai-budget');
          const count = prior?.day === day ? prior.count : 0;
          if (count >= DAILY_LIMIT) return false;
          await storage.put('discovery-ai-budget', { day, count: count + 1 });
          return true;
        });
        if (!allowed) return { retryable: false };
        const controller = new AbortController();
        timer = setTimeout(() => controller.abort(), 8000);
        const result = await this.ai.run(MODEL, { text, source_lang: 'en', target_lang: 'ja' }, { signal: controller.signal });
        const translated = result?.translated_text?.trim();
        if (typeof translated !== 'string' || !translated || translated.length > 320 || !japanese.test(translated) || /[<>\u0000-\u0008]/u.test(translated)) return { retryable: false };
        await this.storage.put(key, translated);
        return { text: translated };
      } catch (error) {
        console.warn('Discovery translation unavailable:', error?.message || String(error));
        return { retryable: true };
      } finally {
        clearTimeout(timer);
        release();
      }
    })();
    this.pending.set(key, promise);
    try { return await promise; } finally { this.pending.delete(key); }
  }
}
