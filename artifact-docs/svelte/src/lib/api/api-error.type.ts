/** O formato único de erro da API, devolvido por todo handler do Go. */
export type ApiErrorBody = {
  statusCode: number;
  message: string;
  details?: { field: string; message: string }[];
  path: string;
  timestamp: string;
};
