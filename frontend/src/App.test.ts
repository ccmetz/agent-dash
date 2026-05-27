import { cleanup, fireEvent, render, screen } from '@testing-library/vue';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App.vue';

describe('Usage Overview shell', () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it('shows Usage Sync diagnostics from backend status', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({
              ok: true,
              analyticsStorePath: 'data/agent-dash.sqlite',
              usageSource: {
                name: 'OpenCode',
                path: '/Users/test/.local/share/opencode/opencode.db',
                available: true,
                state: 'available',
              },
            }),
            {
              headers: { 'Content-Type': 'application/json' },
            },
          ),
      ),
    );

    render(App);

    expect(await screen.findByText('Usage Overview')).toBeTruthy();
    expect(await screen.findByText('Usage Sync')).toBeTruthy();
    expect(await screen.findByText('No sync runs yet')).toBeTruthy();
    expect(await screen.findByText('data/agent-dash.sqlite')).toBeTruthy();
    expect(await screen.findByText('/Users/test/.local/share/opencode/opencode.db')).toBeTruthy();
  });

  it('shows a disconnected state when backend status fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(null, { status: 500 })),
    );

    render(App);

    expect(await screen.findByText('Usage Overview')).toBeTruthy();
    expect(await screen.findByText('Disconnected')).toBeTruthy();
    expect(await screen.findAllByText('Waiting for backend status...')).toHaveLength(3);
  });

  it('shows setup guidance when the OpenCode Usage Source is missing', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({
              ok: true,
              analyticsStorePath: 'data/agent-dash.sqlite',
              usageSource: {
                name: 'OpenCode',
                path: '/Users/test/.local/share/opencode/opencode.db',
                available: false,
                state: 'missing',
              },
            }),
            { headers: { 'Content-Type': 'application/json' } },
          ),
      ),
    );

    render(App);

    expect(await screen.findByText('OpenCode Usage Source missing')).toBeTruthy();
    expect(
      await screen.findAllByText('/Users/test/.local/share/opencode/opencode.db'),
    ).toHaveLength(2);
    expect(
      await screen.findByText('Open OpenCode once, then refresh this dashboard.'),
    ).toBeTruthy();
  });

  it('shows successful Usage Sync diagnostics and can trigger manual refresh', async () => {
    const fetch = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async (url, init) => {
        if (url === '/api/usage-sync' && init?.method === 'POST') {
          return new Response(
            JSON.stringify({
              status: 'success',
              startedAt: '2026-01-01T00:00:00Z',
              finishedAt: '2026-01-01T00:00:01Z',
              inserted: 1,
              updated: 2,
              skipped: 3,
            }),
            { headers: { 'Content-Type': 'application/json' } },
          );
        }
        return new Response(
          JSON.stringify({
            ok: true,
            analyticsStorePath: 'data/agent-dash.sqlite',
            usageSource: {
              name: 'OpenCode',
              path: '/opencode.db',
              available: true,
              state: 'available',
            },
            usageSync: {
              status: 'success',
              pollSeconds: 60,
              nextPollAt: '2026-01-01T00:01:00Z',
              recentRuns: [],
              lastRun: {
                status: 'success',
                startedAt: '2026-01-01T00:00:00Z',
                finishedAt: '2026-01-01T00:00:01Z',
                inserted: 1,
                updated: 2,
                skipped: 3,
              },
            },
          }),
          { headers: { 'Content-Type': 'application/json' } },
        );
      },
    );
    vi.stubGlobal('fetch', fetch);

    render(App);

    expect(await screen.findByText('1 inserted, 2 updated, 3 skipped')).toBeTruthy();
    expect(await screen.findByText('2026-01-01T00:01:00Z')).toBeTruthy();
    await fireEvent.click(await screen.findByRole('button', { name: 'Refresh Usage' }));
    expect(fetch).toHaveBeenCalledWith('/api/usage-sync', { method: 'POST' });
  });

  it('starts a Usage Sync when the dashboard opens', async () => {
    let synced = false;
    const fetch = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async (url, init) => {
        if (url === '/api/usage-sync' && init?.method === 'POST') {
          synced = true;
          return new Response(
            JSON.stringify({
              status: 'success',
              startedAt: '2026-01-01T00:00:00Z',
              finishedAt: '2026-01-01T00:00:01Z',
              inserted: 1,
              updated: 0,
              skipped: 0,
            }),
            { headers: { 'Content-Type': 'application/json' } },
          );
        }
        return new Response(
          JSON.stringify({
            ok: true,
            analyticsStorePath: 'data/agent-dash.sqlite',
            usageSource: {
              name: 'OpenCode',
              path: '/opencode.db',
              available: true,
              state: 'available',
            },
            usageSync: {
              status: synced ? 'success' : 'never_synced',
              pollSeconds: 60,
              nextPollAt: '2026-01-01T00:01:00Z',
              recentRuns: [],
              lastRun: synced
                ? {
                    status: 'success',
                    startedAt: '2026-01-01T00:00:00Z',
                    finishedAt: '2026-01-01T00:00:01Z',
                    inserted: 1,
                    updated: 0,
                    skipped: 0,
                  }
                : undefined,
            },
          }),
          { headers: { 'Content-Type': 'application/json' } },
        );
      },
    );
    vi.stubGlobal('fetch', fetch);

    render(App);

    expect(await screen.findByText('1 inserted, 0 updated, 0 skipped')).toBeTruthy();
    expect(fetch).toHaveBeenCalledWith('/api/usage-sync', { method: 'POST' });
  });

  it('shows Usage Sync in-progress state during the opening sync', async () => {
    let finishSync: (response: Response) => void = () => {};
    vi.stubGlobal(
      'fetch',
      vi.fn<(url: string) => Promise<Response>>((url) => {
        if (url === '/api/usage-sync') {
          return new Promise<Response>((resolve) => {
            finishSync = resolve;
          });
        }
        return Promise.resolve(
          new Response(
            JSON.stringify({
              ok: true,
              analyticsStorePath: 'data/agent-dash.sqlite',
              usageSource: {
                name: 'OpenCode',
                path: '/opencode.db',
                available: true,
                state: 'available',
              },
              usageSync: {
                status: 'never_synced',
                pollSeconds: 60,
                nextPollAt: '2026-01-01T00:01:00Z',
                recentRuns: [],
              },
            }),
            { headers: { 'Content-Type': 'application/json' } },
          ),
        );
      }),
    );

    render(App);

    expect(await screen.findByText('Sync in progress...')).toBeTruthy();
    finishSync(
      new Response(
        JSON.stringify({
          status: 'success',
          finishedAt: 'done',
          inserted: 0,
          updated: 0,
          skipped: 0,
        }),
      ),
    );
  });

  it('keeps stale sync data visible when the latest Usage Sync fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({
              ok: true,
              analyticsStorePath: 'data/agent-dash.sqlite',
              usageSource: {
                name: 'OpenCode',
                path: '/opencode.db',
                available: true,
                state: 'available',
              },
              usageSync: {
                status: 'error',
                pollSeconds: 60,
                nextPollAt: '2026-01-01T00:01:00Z',
                recentRuns: [],
                lastRun: {
                  status: 'error',
                  finishedAt: '2026-01-01T00:00:01Z',
                  inserted: 4,
                  updated: 5,
                  skipped: 6,
                  errorMessage:
                    'Usage Sync failed. Check that the configured OpenCode database is available.',
                },
              },
            }),
            { headers: { 'Content-Type': 'application/json' } },
          ),
      ),
    );

    render(App);

    expect(await screen.findByText('4 inserted, 5 updated, 6 skipped')).toBeTruthy();
    expect(
      await screen.findByText(
        'Usage Sync failed. Check that the configured OpenCode database is available.',
      ),
    ).toBeTruthy();
  });
});
