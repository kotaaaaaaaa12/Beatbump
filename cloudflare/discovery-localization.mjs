// Compatibility shim for deployments that still reference the retired module.
// Localization now uses the frontend catalogs and never calls an AI service.
export class DiscoveryLocalizer {
  async handle() {
    return Response.json({ error: 'automatic_translation_disabled' }, {
      status: 410, headers: { 'Cache-Control': 'no-store', 'X-Content-Type-Options': 'nosniff' },
    });
  }
}
