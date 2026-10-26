<script lang="ts">
  import { page } from '$app/state';

  import ArtifactDetailSlot from './artifact-detail-slot.svelte';
</script>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-5">
  <!-- `{#key}`: ir de um artefato para outro (pelos Recentes, por exemplo) remonta o slot, e o
       hook busca a label nova em vez de continuar mostrando a anterior. -->
  {#key page.params.label}
    <ArtifactDetailSlot label={page.params.label ?? ''} />
  {/key}
</div>
