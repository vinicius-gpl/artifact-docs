import { type AppHub } from '$lib/contracts/app-link.schema';

import { apiRequest } from './http-client';

/** Uma função por endpoint de `/api/apps`. Tipos e apps chegam juntos: a tela sempre agrupa. */
export const appsApi = {
  hub: () => apiRequest<AppHub>('/apps'),
};
