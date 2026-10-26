<script lang="ts">
  import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
  import type { Snippet } from 'svelte';

  import AppErrorBoundary from '$lib/components/acerola-app-error-boundary/acerola-app-error-boundary.svelte';
  import '$lib/theme/tokens.css';

  /**
   * O que TODA tela precisa: o cliente de dados e a captura de erro.
   *
   * A casca (a barra do topo) NÃO mora aqui: ela é de `routes/(app)/+layout.svelte`. Este
   * layout raiz vale também para a tela de erro, que aparece sem a barra.
   */
  let { children }: { children: Snippet } = $props();

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        /* Tentar de novo sozinho esconde a falha por segundos e faz a tela mentir enquanto
           isso. Quando falha, a pessoa precisa ver — em vermelho, com o motivo. */
        retry: false,
        refetchOnWindowFocus: false,
        staleTime: 30_000,
      },
      mutations: {
        /* Numa gravação isso pesa mais: repetir um "criar" que falhou por tempo esgotado
           criaria o registro duas vezes. */
        retry: false,
      },
    },
  });
</script>

<AppErrorBoundary>
  <QueryClientProvider client={queryClient}>
    {@render children()}
  </QueryClientProvider>
</AppErrorBoundary>
