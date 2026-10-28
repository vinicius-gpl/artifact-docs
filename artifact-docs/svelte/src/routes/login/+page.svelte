<script lang="ts">
  import { page } from '$app/state';

  import LoginScreen from './components/acerola-login-screen/acerola-login-screen.svelte';
  import { useLoginModel } from '$lib/hooks/use-login/use-login.svelte';

  /* Fora do grupo `(app)` de propósito: sem a barra e sem a guarda de sessão, que mandaria de
     volta para cá num laço. */
  const returnTo = page.data.returnTo ?? '/';
  const signIn = useLoginModel({ returnTo, error: page.data.error ?? null });
</script>

<svelte:head>
  <title>Entrar</title>
</svelte:head>

<LoginScreen
  data={{
    isMockMode: signIn.data.isMockMode,
    signIn: signIn.data,
  }}
  state={{ signIn: signIn.state }}
  actions={{ signIn: signIn.actions }}
/>
