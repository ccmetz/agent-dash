import { render, screen } from '@testing-library/vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'

describe('Usage Overview shell', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows the connected Analytics Store from backend status', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        new Response(
          JSON.stringify({ ok: true, analyticsStorePath: 'data/agent-dash.sqlite' }),
          { headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )

    render(App)

    expect(await screen.findByText('Usage Overview')).toBeTruthy()
    expect(await screen.findByText('Connected')).toBeTruthy()
    expect(await screen.findByText('data/agent-dash.sqlite')).toBeTruthy()
  })
})
