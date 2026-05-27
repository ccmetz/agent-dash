import { cleanup, render, screen } from '@testing-library/vue';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App.vue';

describe('Usage Overview shell', () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it('shows the connected Analytics Store from backend status', async () => {
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
    expect(await screen.findAllByText('Connected')).toHaveLength(2);
    expect(await screen.findByText('data/agent-dash.sqlite')).toBeTruthy();
    expect(await screen.findByText('OpenCode Usage Source')).toBeTruthy();
    expect(await screen.findByText('/Users/test/.local/share/opencode/opencode.db')).toBeTruthy();
  });

  it('shows a disconnected state when backend status fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(null, { status: 500 })),
    );

    render(App);

    expect(await screen.findByText('Usage Overview')).toBeTruthy();
    expect(await screen.findAllByText('Disconnected')).toHaveLength(2);
    expect(await screen.findAllByText('Waiting for backend status...')).toHaveLength(2);
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
});
