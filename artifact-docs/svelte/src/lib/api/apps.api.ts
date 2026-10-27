import { type AppHub, type AppLink, type AppLinkInput } from '$lib/contracts/app-link.schema';

import { apiRequest } from './http-client';

/** Uma função por endpoint de `/api/apps`. Tipos e apps chegam juntos: a tela sempre agrupa. */
export const appsApi = {
  hub: () => apiRequest<AppHub>('/apps'),
  create: (input: AppLinkInput) => apiRequest<AppLink>('/apps', { method: 'POST', body: input }),
  update: (id: string, input: AppLinkInput) =>
    apiRequest<AppLink>(`/apps/${encodeURIComponent(id)}`, { method: 'PATCH', body: input }),
  remove: (id: string) => apiRequest<void>(`/apps/${encodeURIComponent(id)}`, { method: 'DELETE' }),
};
