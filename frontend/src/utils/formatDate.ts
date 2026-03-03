const MONTHS = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];

/**
 * Format a date string as "Mar 3 2026, 16:25"
 */
export function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return dateStr;
  const mon = MONTHS[d.getMonth()];
  const day = d.getDate();
  const yyyy = d.getFullYear();
  const hh = String(d.getHours()).padStart(2, '0');
  const min = String(d.getMinutes()).padStart(2, '0');
  return `${mon} ${day} ${yyyy}, ${hh}:${min}`;
}

/**
 * Format a date string as "Mar 3" (short, for chart axes).
 */
export function formatDateShort(dateStr: string): string {
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return dateStr;
  return `${MONTHS[d.getMonth()]} ${d.getDate()}`;
}
