/**
 * Lightweight client-side router for Boardgame Luna SPA.
 * Uses Svelte 5 runes — no external libraries needed.
 */

export type RouteParams =
  | { name: 'find' }
  | { name: 'create' }
  | { name: 'profile' }
  | { name: 'filter' }
  | { name: 'map'; mode: 'discover' | 'select' }
  | { name: 'chat'; meetup: any }
  | { name: 'manage'; meetup: any };

/** Main tabs (visible in bottom nav / desktop nav) */
export const MAIN_TABS: RouteParams['name'][] = ['find', 'create', 'profile'];

/** Routes that are "child" fullscreen pages — hide tab bar */
export const CHILD_ROUTES: RouteParams['name'][] = ['filter', 'map', 'chat', 'manage'];

// ── Svelte 5 rune-based global state ──────────────────────────────────────────

let _route = $state<RouteParams>({ name: 'find' });
let _history = $state<RouteParams[]>([]);

export function currentRoute(): RouteParams {
  return _route;
}

/**
 * Navigate to a new route.
 * Pushes current route to history stack before navigating.
 */
export function navigate(to: RouteParams) {
  _history.push(_route);
  _route = to;
}

/**
 * Go back to the previous route.
 * If history is empty, falls back to 'find'.
 */
export function goBack() {
  if (_history.length > 0) {
    _route = _history.pop()!;
  } else {
    _route = { name: 'find' };
  }
}

/**
 * Navigate to a main tab, clearing history.
 */
export function navigateToTab(tab: 'find' | 'create' | 'profile') {
  _history = [];
  _route = { name: tab };
}

/** True when current route is a child page (tab bar should be hidden) */
export function isChildRoute(): boolean {
  return CHILD_ROUTES.includes(_route.name);
}
