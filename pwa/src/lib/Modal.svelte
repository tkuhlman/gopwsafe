<script>
    import { createEventDispatcher } from "svelte";
    import { fly, fade } from "svelte/transition";

    export let title = "Confirm";
    export let message = "Are you sure?";
    export let confirmLabel = "OK";
    export let cancelLabel = "Cancel";
    export let type = "confirm"; // 'confirm', 'alert', 'danger'

    const dispatch = createEventDispatcher();

    function onConfirm() {
        dispatch("confirm");
    }

    function onCancel() {
        dispatch("cancel");
    }
</script>

<div
    class="backdrop"
    role="button"
    tabindex="0"
    on:click={onCancel}
    on:keydown={(e) => e.key === "Escape" && onCancel()}
    transition:fade
>
    <div class="modal" transition:fly={{ y: -50, duration: 300 }}>
        <div class="header">
            <h3>{title}</h3>
        </div>
        <div class="body">
            <p>{message}</p>
        </div>
        <div class="footer">
            {#if type !== "alert"}
                <button class="secondary" on:click={onCancel}
                    >{cancelLabel}</button
                >
            {/if}
            <button
                class:danger={type === "danger"}
                class:primary={type !== "danger"}
                on:click={onConfirm}>{confirmLabel}</button
            >
        </div>
    </div>
</div>

<style>
    .backdrop {
        position: fixed;
        top: 0;
        left: 0;
        width: 100vw;
        height: 100vh;
        background: rgba(0, 0, 0, 0.5);
        z-index: 3000;
        display: flex;
        justify-content: center;
        align-items: center;
    }
    .modal {
        background: #252526;
        border: 1px solid #333;
        border-radius: 8px;
        min-width: 300px;
        max-width: 500px;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
        display: flex;
        flex-direction: column;
    }
    .header {
        padding: 15px 20px;
        border-bottom: 1px solid #333;
    }
    .header h3 {
        margin: 0;
        font-size: 1.2rem;
        color: #e0e0e0;
    }
    .body {
        padding: 20px;
        color: #ccc;
        font-size: 1rem;
    }
    .body p {
        margin: 0;
    }
    .footer {
        padding: 15px 20px;
        border-top: 1px solid #333;
        display: flex;
        justify-content: flex-end;
        gap: 10px;
    }
    button {
        padding: 8px 16px;
        border-radius: 4px;
        cursor: pointer;
        border: none;
        font-size: 0.9rem;
    }
    button.secondary {
        background: transparent;
        color: #aaa;
        border: 1px solid #444;
    }
    button.secondary:hover {
        border-color: #666;
        color: #ccc;
    }
    button.primary {
        background: #007bff;
        color: white;
    }
    button.primary:hover {
        background: #0056b3;
    }
    button.danger {
        background: #dc3545;
        color: white;
    }
    button.danger:hover {
        background: #a71d2a;
    }
</style>
