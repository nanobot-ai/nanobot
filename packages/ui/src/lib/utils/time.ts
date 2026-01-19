export interface TimeAgoResult {
	relativeTime: string;
	fullDate: string;
}

/**
 * Formats a timestamp into a relative time description ("2 hours ago") and a localized full date string
 * @param timestamp ISO string date or undefined
 * @param granularity return a relative time with this granularity
 * @returns Object containing relativeTime and fullDate strings
 */
export function formatTimeAgo(timestamp: string | undefined, granularity?: string): TimeAgoResult {
	if (!timestamp) return { relativeTime: '', fullDate: '' };

	const now = new Date();
	const date = new Date(timestamp);
	const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

	// Format the full date for the tooltip
	const options: Intl.DateTimeFormatOptions = {
		weekday: 'long',
		year: 'numeric',
		month: 'long',
		day: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		hour12: true
	};
	const fullDate = date.toLocaleString(undefined, options);

	// Relative time calculation
	let relativeTime = '';
	let interval = Math.floor(seconds / 31536000);
	if (interval >= 1) {
		relativeTime = interval === 1 ? '1 year ago' : `${interval} years ago`;
	} else if (granularity === 'year') {
		relativeTime = 'This year';
	} else {
		interval = Math.floor(seconds / 2592000);
		if (interval >= 1) {
			relativeTime = interval === 1 ? '1 month ago' : `${interval} months ago`;
		} else if (granularity === 'month') {
			relativeTime = 'This month';
		} else {
			interval = Math.floor(seconds / 86400);
			if (interval >= 1) {
				relativeTime = interval === 1 ? '1 day ago' : `${interval} days ago`;
			} else if (granularity === 'day') {
				relativeTime = 'Today';
			} else {
				interval = Math.floor(seconds / 3600);
				if (interval >= 1) {
					relativeTime = interval === 1 ? '1 hour ago' : `${interval} hours ago`;
				} else if (granularity === 'hour') {
					relativeTime = 'In the last hour';
				} else {
					interval = Math.floor(seconds / 60);
					if (interval >= 1) {
						relativeTime = interval === 1 ? '1 minute ago' : `${interval} minutes ago`;
					} else if (granularity === 'minute') {
						relativeTime = 'In the last minute';
					} else {
						if (seconds < 10) return { relativeTime: 'just now', fullDate };
						relativeTime = `${Math.floor(seconds)} seconds ago`;
					}
				}
			}
		}
	}

	return { relativeTime, fullDate };
}
