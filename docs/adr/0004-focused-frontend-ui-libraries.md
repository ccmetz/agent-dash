# Focused Frontend UI Libraries

The Vue frontend will not adopt a full opinionated component framework at this stage. It will keep Tailwind as the styling foundation, extract thin app-owned primitive UI components immediately, and add focused libraries lazily when concrete needs appear: Chart.js for charts, TanStack Table for interactive table state, and Reka UI for accessible primitives.

This keeps the Usage Overview visually distinct and domain-shaped while avoiding custom implementations of charting, table state, and accessibility-heavy primitives. Broader frameworks such as Vuetify, PrimeVue, Element Plus, or Quasar remain valid future options if the app grows many complex forms, needs primitives faster than app-owned styling can support, requires heavy data-grid behavior such as virtualization or pinned/resizable columns, or custom primitive maintenance becomes a measurable drag.

Third-party library APIs should stay contained inside app-owned primitives, chart wrappers, table components, or feature-local components rather than spreading through page-level views. Primitive UI components should remain thin style and accessibility wrappers; domain behavior such as Usage Sync loading, Billing Caveat presentation, or Usage Overview composition belongs in feature components.
