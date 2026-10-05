import type { GlobalIntersectionObserver } from "./observer";

export const CALLBACK_MAP = {
	images: (
		thisArg: GlobalIntersectionObserver,
		entry: IntersectionObserverEntry,
	) => {
		const target = entry.target as HTMLImageElement;
		if (entry.isIntersecting) {
			thisArg.unobserve(target);
			if (!target.dataset.src) return;
			// An absent placeholder must not prevent loading the actual thumbnail.
			target.src = target.dataset.src;
			// The image's onerror handler supplies its visible fallback.
			void target.decode().catch(() => {});
		}
	},
} as const;
