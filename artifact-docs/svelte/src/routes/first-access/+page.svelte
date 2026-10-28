<script lang="ts">
  import FirstAccessScreen from './components/acerola-first-access-screen/acerola-first-access-screen.svelte';
  import { useFirstAccessModel } from '$lib/hooks/use-first-access/use-first-access.svelte';

  /* Fora do grupo `(app)`, como o login: quem está aqui ainda não tem sessão, e a guarda
     daquele grupo devolveria a pessoa para `/login` num laço. */
  const model = useFirstAccessModel();
</script>

<svelte:head>
  <title>Primeiro acesso</title>
</svelte:head>

<FirstAccessScreen data={model.data} state={model.state} actions={model.actions} />
