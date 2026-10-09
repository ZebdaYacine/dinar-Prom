import {describe, expect, it} from 'vitest';
import {isSupportedLocale} from './locale';

describe('isSupportedLocale', () => {
  it('accepts the two supported UI languages', () => {
    expect(isSupportedLocale('ar')).toBe(true);
    expect(isSupportedLocale('fr')).toBe(true);
  });

  it('rejects unsupported languages', () => {
    expect(isSupportedLocale('en')).toBe(false);
    expect(isSupportedLocale('')).toBe(false);
  });
});
