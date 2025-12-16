<script>
    import { onMount, createEventDispatcher } from "svelte";
    import { get, set } from "idb-keyval";
    import { openDatabase, getDatabaseData } from "../wasm.js";
    import { selectedFile, dbItems } from "../store.js";

    const dispatch = createEventDispatcher();

    let password = "";
    let error = "";
    let isLoading = false;
    let recentFiles = [];
    let currentHandle = null;

    onMount(async () => {
        const recents = await get("recentFiles");
        if (recents) {
            recentFiles = recents;
        }
    });

    async function openFile() {
        try {
            [currentHandle] = await window.showOpenFilePicker({
                types: [
                    {
                        description: "Password Safe Strings",
                        accept: {
                            "application/octet-stream": [".psafe3", ".dat"],
                        },
                    },
                ],
            });
            console.log("File selected:", currentHandle.name);
        } catch (e) {
            // User cancelled or not supported
            console.error("File selection failed", e);
            return;
        }
    }

    async function unlock() {
        if (!currentHandle) return;
        isLoading = true;
        error = "";

        try {
            const file = await currentHandle.getFile();
            const arrayBuffer = await file.arrayBuffer();
            const uint8Array = new Uint8Array(arrayBuffer);

            openDatabase(uint8Array, password);

            const items = getDatabaseData();
            dbItems.set(items);
            selectedFile.set({
                handle: currentHandle,
                name: currentHandle.name,
            });

            // Update recents
            // Note: Storing handles in IndexedDB is supported in some browsers but can be tricky permission-wise.
            // Usually you store the handle and request permission again on next load.
            // For simplicity, we just store name/date here for list, but we'd need the handle to re-open easily.
            // Key storage of handles:
            const newRecent = {
                name: currentHandle.name,
                date: new Date().toISOString(),
            };
            const otherRecents = recentFiles.filter(
                (r) => r.name !== currentHandle.name,
            );
            await set("recentFiles", [newRecent, ...otherRecents.slice(0, 4)]);
            await set("lastHandle", currentHandle);

            dispatch("opened");
        } catch (e) {
            console.error(e);
            error = "Failed to unlock: " + e.message;
        } finally {
            isLoading = false;
        }
    }

    async function loadRecent(fileInfo) {
        // Re-opening from recent list is complex with File System Access API as you need to store the handle.
        // We will skip "click to re-open" for now and just ask user to pick file again,
        // or implemented "lastHandle" restoration if we had time.
        // Since user wants simple: we stick to "Open File" button but remembering location is what they asked.
        // The browser picker defaults to last location often.
        // "Web storage should be used to remember the location of previously selected files."
        // Since we can't easily instruct the picker to start at a path in web, we rely on browser default or just showing the name.
        // Actually we CAN store the handle in IndexedDB and reuse it (perm request needed).

        // Attempt to retrieve last handle?
        try {
            const handle = await get("lastHandle");
            if (handle && handle.name === fileInfo.name) {
                // verify permission
                if ((await handle.queryPermission()) === "granted") {
                    currentHandle = handle;
                    return;
                }
                if ((await handle.requestPermission()) === "granted") {
                    currentHandle = handle;
                    return;
                }
            }
        } catch (e) {
            console.log("Could not restore handle", e);
        }
        alert("Please select the file again.");
        openFile();
    }
</script>

<div class="start-page">
    <h1>Password Safe</h1>

    {#if !currentHandle}
        <div class="actions">
            <button on:click={openFile}>Open Database File</button>
        </div>

        {#if recentFiles.length > 0}
            <div class="recents">
                <h3>Recent Files</h3>
                <ul>
                    {#each recentFiles as file}
                        <li>
                            <button
                                class="link-button"
                                on:click={() => loadRecent(file)}
                                >{file.name}</button
                            >
                            <span class="date"
                                >{new Date(
                                    file.date,
                                ).toLocaleDateString()}</span
                            >
                        </li>
                    {/each}
                </ul>
            </div>
        {/if}
    {:else}
        <div class="login-box">
            <h2>Unlock {currentHandle.name}</h2>
            <div class="input-group">
                <input
                    type="password"
                    bind:value={password}
                    placeholder="Password"
                    autofocus
                    on:keydown={(e) => e.key === "Enter" && unlock()}
                />
                <button on:click={unlock} disabled={isLoading}>
                    {isLoading ? "Unlocking..." : "Unlock"}
                </button>
            </div>
            {#if error}
                <div class="error">{error}</div>
            {/if}
            <button
                class="secondary"
                on:click={() => {
                    currentHandle = null;
                    password = "";
                    error = "";
                }}>Back</button
            >
        </div>
    {/if}
</div>

<style>
    .start-page {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        flex-grow: 1;
        padding: 2rem;
    }
    .actions button {
        font-size: 1.2rem;
        padding: 1rem 2rem;
        background: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        cursor: pointer;
    }
    .recents {
        margin-top: 2rem;
        width: 100%;
        max-width: 400px;
    }
    .recents ul {
        list-style: none;
        padding: 0;
    }
    .recents li {
        display: flex;
        justify-content: space-between;
        padding: 0.5rem;
        border-bottom: 1px solid #333;
    }
    .link-button {
        background: none;
        border: none;
        color: #64b5f6;
        cursor: pointer;
        text-decoration: underline;
        padding: 0;
        font-size: 1rem;
    }
    .login-box {
        background: #2d2d2d;
        padding: 2rem;
        border-radius: 8px;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
        width: 100%;
        max-width: 400px;
    }
    .input-group {
        display: flex;
        gap: 0.5rem;
        margin-bottom: 1rem;
    }
    input {
        flex-grow: 1;
        padding: 0.5rem;
        border-radius: 4px;
        border: 1px solid #444;
        background: #333;
        color: white;
    }
    .error {
        color: #ff6b6b;
        margin-bottom: 1rem;
    }
    .secondary {
        background: transparent;
        border: 1px solid #666;
        color: #999;
        padding: 0.4rem 1rem;
        cursor: pointer;
        border-radius: 4px;
    }
</style>
