/**
 * Lightweight Hash-based client-side router for Boardgame Luna SPA.
 * Uses Svelte 5 runes — fully synchronized with browser address bar.
 */

export type RouteParams =
  | { name: 'find' }
  | { name: 'my-meetups' }
  | { name: 'chats' }
  | { name: 'create' }
  | { name: 'profile' }
  | { name: 'filter' }
  | { name: 'map'; mode: 'discover' | 'select'; meetupId?: string }
  | { name: 'chat'; meetupId: string }
  | { name: 'manage'; meetupId: string };

/** Routes that are "child" fullscreen pages — hide tab bar */
export const CHILD_ROUTES: RouteParams['name'][] = ['filter', 'map', 'chat', 'manage'];

// ── Svelte 5 rune-based global state ──────────────────────────────────────────

let _route = $state<RouteParams>({ name: 'find' });
let _history = $state<RouteParams[]>([]);

export function currentRoute(): RouteParams {
  return _route;
}

/**
 * Parse hash string (e.g. "#/manage/123") into RouteParams object
 */
export function parseHash(hash: string): RouteParams {
  const cleanHash = hash.startsWith('#') ? hash.slice(1) : hash;
  if (!cleanHash || cleanHash === '/' || cleanHash === '/find') {
    return { name: 'find' };
  }
  if (cleanHash === '/my-meetups') {
    return { name: 'my-meetups' };
  }
  if (cleanHash === '/chats') {
    return { name: 'chats' };
  }
  if (cleanHash === '/create') {
    return { name: 'create' };
  }
  if (cleanHash === '/profile') {
    return { name: 'profile' };
  }
  if (cleanHash === '/filter') {
    return { name: 'filter' };
  }
  if (cleanHash === '/map/discover') {
    return { name: 'map', mode: 'discover' };
  }
  if (cleanHash === '/map/select') {
    return { name: 'map', mode: 'select' };
  }

  // Parse dynamic paths (e.g. ["", "map", "discover", "123"] or ["", "chat", "123"])
  const parts = cleanHash.split('/');
  if (parts.length >= 3) {
    const routeName = parts[1];
    const subParam = parts[2];
    if (routeName === 'map') {
      const mode = (subParam === 'select' ? 'select' : 'discover') as 'discover' | 'select';
      const meetupId = parts[3];
      return { name: 'map', mode, meetupId };
    }
    if (routeName === 'chat') {
      return { name: 'chat', meetupId: subParam };
    }
    if (routeName === 'manage') {
      return { name: 'manage', meetupId: subParam };
    }
  }

  return { name: 'find' };
}

/**
 * Build hash string from RouteParams object
 */
export function buildHash(route: RouteParams): string {
  if (route.name === 'find') return '#/find';
  if (route.name === 'my-meetups') return '#/my-meetups';
  if (route.name === 'chats') return '#/chats';
  if (route.name === 'create') return '#/create';
  if (route.name === 'profile') return '#/profile';
  if (route.name === 'filter') return '#/filter';
  if (route.name === 'map') {
    return route.meetupId ? `#/map/${route.mode}/${route.meetupId}` : `#/map/${route.mode}`;
  }
  if (route.name === 'chat') return `#/chat/${route.meetupId}`;
  if (route.name === 'manage') return `#/manage/${route.meetupId}`;
  return '#/find';
}

/**
 * Navigate to a new route by setting window hash.
 */
export function navigate(to: RouteParams) {
  const newHash = buildHash(to);
  _history.push(_route);
  window.location.hash = newHash;
}

/**
 * Go back to the previous route.
 */
export function goBack() {
  if (_history.length > 0) {
    const prev = _history.pop()!;
    window.location.hash = buildHash(prev);
  } else {
    window.location.hash = '#/find';
  }
}

export function navigateToTab(tab: 'find' | 'my-meetups' | 'chats' | 'create' | 'profile') {
  _history = [];
  window.location.hash = `#/${tab}`;
}

/** True when current route is a child page (tab bar should be hidden) */
export function isChildRoute(): boolean {
  return CHILD_ROUTES.includes(_route.name);
}

// ── Event Listener for Hash Change ───────────────────────────────────────────

export function initRouter() {
  const syncRoute = () => {
    _route = parseHash(window.location.hash);
  };

  // Sync initially
  syncRoute();

  // Listen to browser back/forward and programmatic hash edits
  window.addEventListener('hashchange', syncRoute);

  return () => {
    window.removeEventListener('hashchange', syncRoute);
  };
}
