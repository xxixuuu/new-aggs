export function groupBy<K extends string | number, T>(
  array: T[],
  fn: (v: T) => K
): Map<K, T[]> {
  const res = new Map<K, T[]>();
  for (const e of array) {
    const key = fn(e);
    if (res.has(key)) {
      res.get(key)!.push(e);
    } else {
      res.set(key, [e]);
    }
  }
  return res;
}

export function range(from: number, to: number, step: number): number[] {
  const res: number[] = [];
  for (let i = from; i < to; i += step) res.push(i);
  return res;
}

export function throwError(message?: string, options?: ErrorOptions): never {
  throw new Error(message, options);
}

export function compare(a: number, b: number) {
  return a < b ? -1 : a > b ? 1 : 0;
}

export function isNotNull<T>(v: T | null | undefined): v is T {
  return v != null;
}

export function formatSec(sec: number) {
  return new Date(sec * 1000).toLocaleString("ja-JP", {
    hour: "numeric",
    minute: "numeric",
    second: "numeric",
    fractionalSecondDigits: 3,
  });
}
