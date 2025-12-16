<script>
  import { onMount } from "svelte";
  import { loadWasm } from "./wasm.js";
  import StartPage from "./lib/StartPage.svelte";
  import Dashboard from "./lib/Dashboard.svelte";
  import { selectedFile } from "./store.js";

  let wasmReady = false;
  let view = "start"; // start, dashboard

  onMount(async () => {
    try {
      await loadWasm();
      wasmReady = true;
    } catch (e) {
      console.error("WASM Load Error", e);
      alert("Failed to initialize engine.");
    }
  });

  function onOpened(event) {
    console.log("App received opened event, switching to dashboard");
    // event.detail has { items, fileName }
    // Store accessible via store or context?
    // Since specific Record fetching is needed later, StartPage just unlocks DB.
    view = "dashboard";
  }
</script>

<main class="container">
  {#if !wasmReady}
    <div class="loading">Loading Password Safe Core...</div>
  {:else if view === "start"}
    <StartPage on:opened={onOpened} />
  {:else}
    <Dashboard on:close={() => (view = "start")} />
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
    background: #1e1e1e;
    color: #e0e0e0;
  }
  .container {
    height: 100vh;
    display: flex;
    flex-direction: column;
  }
  .loading {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
  }
</style>
