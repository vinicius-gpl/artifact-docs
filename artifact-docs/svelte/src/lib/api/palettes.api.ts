import { type PaletteList } from '$lib/contracts/palette.schema';

import { apiRequest } from './http-client';

/** Uma função por endpoint de `/api/palettes`. */
export const palettesApi = {
  list: () => apiRequest<PaletteList>('/palettes'),
};
