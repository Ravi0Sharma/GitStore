import { formatDistanceToNow } from 'date-fns';

/**
 * Safely formats a date value to a relative time string.
 * Handles Date, string, number, null, and undefined values.
 * Returns "Unknown" for invalid dates.
 */
export function safeDistanceToNow(value: unknown): string {
  if (value == null) {
    return 'Unknown';
  }

  let date: Date;

  if (value instanceof Date) {
    date = value;
  } else if (typeof value === 'string') {
    date = new Date(value);
  } else if (typeof value === 'number') {
    date = new Date(value);
  } else {
    return 'Unknown';
  }

  // Check if date is valid
  if (Number.isNaN(date.getTime())) {
    return 'Unknown';
  }

  try {
    return formatDistanceToNow(date, { addSuffix: true });
  } catch (error) {
    return 'Unknown';
  }
}


