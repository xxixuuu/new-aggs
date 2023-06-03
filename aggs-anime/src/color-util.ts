import Color from "color";

const cache: Record<string, string> = {};

export function darken(base: string, value: number) {
  const key = `darken:${base}:${value}`;
  const cached = cache[key];
  if (cached) return cached;

  const c = Color(base).darken(value).toString();
  cache[key] = c;
  return c;
}

export function lighten(base: string, value: number) {
  const key = `lighten:${base}:${value}`;
  const cached = cache[key];
  if (cached) return cached;

  const c = Color(base).lighten(value).toString();
  cache[key] = c;
  return c;
}

export function desaturate(base: string, value: number) {
  const key = `desaturate:${base}:${value}`;
  const cached = cache[key];
  if (cached) return cached;

  const c = Color(base).desaturate(value).toString();
  cache[key] = c;
  return c;
}
