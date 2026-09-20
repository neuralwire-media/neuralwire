/**
 * Normalizes whitespace (collapsing newlines, tabs, and multiple spaces into a single space)
 * and optionally truncates cleanly at word boundary without exceeding maxLen characters.
 */
export function cleanMetaDescription(text: string, maxLen = 160): string {
	if (!text) return '';
	const cleaned = text.trim().replace(/\s+/g, ' ');
	if (maxLen <= 0 || cleaned.length <= maxLen) {
		return cleaned;
	}
	const target = maxLen - 3;
	if (target <= 0) {
		return cleaned.slice(0, maxLen);
	}
	let sub = cleaned.slice(0, target);
	const lastSpace = sub.lastIndexOf(' ');
	if (lastSpace > 0) {
		sub = sub.slice(0, lastSpace);
	}
	sub = sub.replace(/[ ,;:.\-–—]+$/, '');
	return `${sub}...`;
}
