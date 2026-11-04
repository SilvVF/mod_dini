export function arrayAny<T>(
  value: T[],
  predicate: (value: T, index: number, array: T[]) => boolean,
): boolean {
  for (let i = 0; i < value.length; i++) {
    if (predicate(value[i], i, value)) {
      return true;
    }
  }

  return false;
}

export function firstNotNullOf<T, R>(
  value: T[],
  transform: (value: T, index: number, array: T[]) => R | null | undefined,
): NonNullable<R> {
  for (let i = 0; i < value.length; i++) {
    const result = transform(value[i], i, value);

    if (result != null) {
      return result as NonNullable<R>;
    }
  }

  throw new Error("No element was transformed to a non-null value.");
}

export function groupBy<T, K extends PropertyKey>(
  items: Iterable<T>,
  keySelector: (item: T, index: number) => K,
): Record<K, T[]>;

export function groupBy<T, K extends PropertyKey, V>(
  items: Iterable<T>,
  keySelector: (item: T, index: number) => K,
  valueTransform: (item: T, index: number) => V,
): Record<K, V[]>;

export function groupBy<T, K extends PropertyKey, V>(
  items: Iterable<T>,
  keySelector: (item: T, index: number) => K,
  valueTransform?: (item: T, index: number) => V,
): Record<K, (T | V)[]> {
  const result = Object.create(null) as Record<PropertyKey, (T | V)[]>;
  let index = 0;

  for (const item of items) {
    const key = keySelector(item, index);
    (result[key] ??= []).push(valueTransform ? valueTransform(item, index) : item);
    index++;
  }

  return result;
}

export function firstNotNullOfOrNull<T, R>(
  arr: T[],
  transform: (value: T, index: number, array: T[]) => R | null | undefined,
): NonNullable<R> | undefined {
  for (let i = 0; i < arr.length; i++) {
    const result = transform(arr[i], i, arr);

    if (result != null) {
      return result as NonNullable<R>;
    }
  }

  return undefined;
}

export function range(start: number, stop: number, step = 1): number[] {
  if (step === 0) {
    throw new RangeError("Step cannot be zero.");
  }

  if ((stop - start) * step < 0) {
    return [];
  }

  const result: number[] = [];

  for (
    let value = start;
    step > 0 ? value <= stop : value >= stop;
    value += step
  ) {
    result.push(value);
  }

  return result;
}

export function dropLastWhile(
  string: string,
  predicate: (character: string) => boolean,
): string {
  let end = string.length;

  while (end > 0 && predicate(string[end - 1])) {
    end--;
  }

  return string.slice(0, end);
}

export function isEmptySet<T>(value: Set<T>): boolean {
  return !((value?.size ?? 0) > 0);
}

export function isEmptyArray<T>(value: Array<T>): boolean {
  return !((value?.length ?? 0) > 0);
}

export function isBlank(value: string): boolean {
  return value.trim().length === 0;
}

export function setMap<T, R>(
  value: Set<T>,
  block: (value: T, index: number, set: Set<T>) => R,
): R[] {
  return iterableMap(value.values(), (item, index) => block(item, index, value));
}

function iterableMap<T, R>(
  iterable: Iterable<T>,
  block: (item: T, index: number) => R,
): R[] {
  const result: R[] = [];
  let index = 0;

  for (const item of iterable) {
    result.push(block(item, index++));
  }

  return result;
}

export function sumBy<T>(
  items: Iterable<T>,
  selector: (item: T) => number,
): number {
  let sum = 0;

  for (const item of items) {
    sum += selector(item);
  }

  return sum;
}
