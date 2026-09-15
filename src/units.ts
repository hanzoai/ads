// Money and time as the platform counts them: cents and unix seconds.

/** Cents as a plain amount, "1,250.00". The platform names no currency. */
export const money = (cents: number): string =>
  (cents / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })

/** An amount typed by a person, as cents. Anything unreadable is 0. */
export const cents = (typed: string): number => {
  const n = Number.parseFloat(typed.replace(/,/g, ''))
  return Number.isFinite(n) && n > 0 ? Math.round(n * 100) : 0
}

/** Distance from now in the coarsest useful unit: "3 hours ago". */
export function since(seconds: number): string {
  if (!seconds) return ''
  const size = Math.abs(Math.floor(Date.now() / 1000) - seconds)
  const [count, unit] =
    size < 3600
      ? [Math.max(1, Math.round(size / 60)), 'minute']
      : size < 86400
        ? [Math.round(size / 3600), 'hour']
        : size < 86400 * 30
          ? [Math.round(size / 86400), 'day']
          : [Math.round(size / (86400 * 30)), 'month']
  return `${count} ${unit}${count === 1 ? '' : 's'} ago`
}
