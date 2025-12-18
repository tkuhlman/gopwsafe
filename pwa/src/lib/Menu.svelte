<script>
    import { createEventDispatcher } from "svelte";

    // We can accept menu items via props or use slots.
    // Let's use slots for maximum flexibility (custom actions).
    // Or we can simple use a variable state for open/close.

    let isOpen = false;

    function toggle() {
        isOpen = !isOpen;
    }

    function close() {
        isOpen = false;
    }

    // @ts-ignore
    const appVersion = __APP_VERSION__;
</script>

<div class="menu-container">
    <button class="hamburger" on:click={toggle} aria-label="Menu">
        <span></span>
        <span></span>
        <span></span>
    </button>

    {#if isOpen}
        <div class="backdrop" on:click={close}></div>
        <div class="menu-dropdown">
            <slot {close}></slot>
            <div class="menu-footer">v{appVersion}</div>
        </div>
    {/if}
</div>

<style>
    .menu-container {
        position: relative;
        display: inline-block;
    }
    .hamburger {
        background: none;
        border: none;
        cursor: pointer;
        padding: 5px;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        width: 30px;
        height: 24px;
        z-index: 1001; /* Above dropdown */
    }
    .hamburger span {
        display: block;
        height: 3px;
        width: 100%;
        background-color: #e0e0e0;
        border-radius: 2px;
    }

    .backdrop {
        position: fixed;
        top: 0;
        left: 0;
        width: 100vw;
        height: 100vh;
        z-index: 999;
        /* background: rgba(0,0,0,0.1); Optional dim? */
    }

    .menu-dropdown {
        position: absolute;
        top: 35px; /* below hasmburger */
        left: 0;
        min-width: 200px;
        background: #252526;
        border: 1px solid #333;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.5);
        border-radius: 4px;
        z-index: 1000;
        padding: 5px 0;
        display: flex;
        flex-direction: column;
    }

    /* We expect global or slotted styles for buttons, but let's provide some helpers if needed 
       Actually, `slot` content style is up to parent slightly, but usually we put buttons there.
       We can enforce structure by selecting `button` inside? */

    :global(.menu-dropdown button) {
        display: block;
        width: 100%;
        text-align: left;
        background: none;
        border: none;
        color: #e0e0e0;
        padding: 10px 15px;
        cursor: pointer;
        font-size: 14px;
    }
    :global(.menu-dropdown button:hover) {
        background: #37373d;
    }

    .menu-footer {
        padding: 5px 15px;
        color: #888;
        font-size: 0.8em;
        text-align: right;
        border-top: 1px solid #333;
        margin-top: 5px;
    }
</style>
