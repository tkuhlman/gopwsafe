<script>
    import { createEventDispatcher } from "svelte";
    import { dbItems, selectedFile } from "../store.js";

    import { getRecordData, getDatabaseInfo } from "../wasm.js";
    import Menu from "./Menu.svelte";

    const dispatch = createEventDispatcher();

    let items = [];
    let filteredItems = [];
    let searchTerm = "";
    let selectedRecord = null;
    let showPassword = false;
    let groupedItems = {};
    let searchInput; // Reference for autofocus

    dbItems.subscribe((val) => {
        items = val || [];
        filterItems();
        // Autofocus search when items are loaded (DB opened)
        setTimeout(() => {
            if (searchInput) searchInput.focus();
        }, 100);
    });

    function filterItems() {
        if (!searchTerm) {
            filteredItems = items;
        } else {
            const lower = searchTerm.toLowerCase();
            filteredItems = items.filter(
                (i) =>
                    i.title.toLowerCase().includes(lower) ||
                    i.group.toLowerCase().includes(lower),
            );
        }
        groupItems(filteredItems);
    }

    function groupItems(itemList) {
        const groups = {};
        itemList.forEach((item) => {
            const g = item.group || "Ungrouped";
            if (!groups[g]) groups[g] = [];
            groups[g].push(item);
        });
        // Sort groups and items in groups
        const sortedKeys = Object.keys(groups).sort();
        const grouped = {};
        sortedKeys.forEach((k) => {
            grouped[k] = groups[k].sort((a, b) =>
                a.title.localeCompare(b.title),
            );
        });
        groupedItems = grouped;
    }

    function selectItem(item) {
        try {
            // fetch full details
            // item.title is the key
            const rec = getRecordData(item.title);
            selectedRecord = rec;
            showPassword = false;
        } catch (e) {
            console.error(e);
            alert("Failed to load record details");
        }
    }

    function formatDate(str) {
        if (!str) return "";
        return new Date(str).toLocaleString();
    }

    // Save isn't implemented in WASM/Go yet properly for updates, just read-only per initial simplified prompt?
    // "The main view will also need a menu for showing DB metadata and adding new entries as well as saving."
    // "If this is too large of an initial task the first version could be read only."
    // I have NOT implemented saveDB fully in Go side yet (just referenced it).
    // I will check Go implementation.
    // Actually I implemented `openDB`, `getDBData`, `getRecord` in Go. `saveDB` is commented out.
    // So for now, it IS read-only.

    function save() {
        alert("Save functionality not yet implemented in V1");
    }

    function showDBInfo() {
        try {
            const info = getDatabaseInfo();
            const msg = `
DB Info:
Description: ${info.description}
Version: ${info.version}
UUID: ${info.uuid}
Last Save: ${info.when} by ${info.who} using ${info.what}
            `;
            alert(msg);
        } catch (e) {
            console.error(e);
            alert("Failed to get DB info");
        }
    }
</script>

<div class="dashboard">
    <div class="sidebar">
        <div class="toolbar">
            <Menu let:close>
                <button
                    on:click={() => {
                        close();
                        save();
                    }}>Save DB</button
                >
                <button
                    on:click={() => {
                        close();
                        showDBInfo();
                    }}>DB Info</button
                >
                <hr
                    style="border: 0; border-top: 1px solid #444; margin: 5px 0;"
                />
                <button
                    on:click={() => {
                        close();
                        dispatch("close");
                    }}>Close DB</button
                >
            </Menu>
            <input
                bind:this={searchInput}
                type="text"
                placeholder="Search..."
                bind:value={searchTerm}
                on:input={filterItems}
            />
        </div>
        <div class="tree">
            {#each Object.keys(groupedItems) as group}
                <details open>
                    <summary>{group}</summary>
                    <ul>
                        {#each groupedItems[group] as item}
                            <li
                                class:selected={selectedRecord &&
                                    selectedRecord.Title === item.title}
                                on:click={() => selectItem(item)}
                            >
                                {item.title}
                            </li>
                        {/each}
                    </ul>
                </details>
            {/each}
        </div>
    </div>

    <div class="main-content" class:mobile-open={!!selectedRecord}>
        {#if selectedRecord}
            <div class="record-details">
                <div class="details-header">
                    <button
                        class="close-details"
                        on:click={() => (selectedRecord = null)}>✕</button
                    >
                    <h2>{selectedRecord.Title}</h2>
                </div>
                <div class="field">
                    <label>Group</label>
                    <div>{selectedRecord.Group}</div>
                </div>
                <div class="field">
                    <label>Username</label>
                    <div>{selectedRecord.Username}</div>
                </div>
                <div class="field">
                    <label>Password</label>
                    <div class="password-row">
                        <input
                            type={showPassword ? "text" : "password"}
                            value={selectedRecord.Password}
                            readonly
                        />
                        <button on:click={() => (showPassword = !showPassword)}>
                            {showPassword ? "Hide" : "Show"}
                        </button>
                    </div>
                </div>
                <div class="field">
                    <label>URL</label>
                    <div>
                        <a href={selectedRecord.URL} target="_blank"
                            >{selectedRecord.URL}</a
                        >
                    </div>
                </div>
                <div class="field">
                    <label>Notes</label>
                    <pre>{selectedRecord.Notes}</pre>
                </div>

                <hr />
                <div class="meta">
                    <small>Modified: {formatDate(selectedRecord.ModTime)}</small
                    >
                </div>
            </div>
        {:else}
            <div class="empty-state">Select a record to view details</div>
        {/if}
    </div>
</div>

<style>
    .dashboard {
        display: flex;
        height: 100vh;
        width: 100%;
        text-align: left;
    }
    .sidebar {
        width: 300px;
        background: #252526;
        border-right: 1px solid #333;
        display: flex;
        flex-direction: column;
    }
    .toolbar {
        padding: 10px;
        border-bottom: 1px solid #333;
        display: flex;
        gap: 10px;
        align-items: center;
    }
    .toolbar input {
        width: 100%;
        padding: 5px;
        background: #3c3c3c;
        border: 1px solid #555;
        color: #fff;
    }
    .tree {
        flex: 1;
        overflow-y: auto;
        padding: 10px;
    }
    .tree ul {
        list-style: none;
        padding-left: 20px;
        margin: 5px 0;
    }
    .tree li {
        padding: 4px 8px;
        cursor: pointer;
        border-radius: 3px;
    }
    .tree li:hover {
        background: #37373d;
    }
    .tree li.selected {
        background: #094771;
        color: white;
    }
    details summary {
        cursor: pointer;
        font-weight: bold;
        color: #ccc;
    }
    .footer {
        padding: 10px;
        border-top: 1px solid #333;
    }
    .main-content {
        flex: 1;
        padding: 20px;
        overflow-y: auto;
        background: #1e1e1e;
    }
    .details-header {
        display: flex;
        align-items: center;
        gap: 10px;
    }
    .close-details {
        background: none;
        border: none;
        color: #ccc;
        font-size: 1.5rem;
        cursor: pointer;
        display: none; /* Hidden on desktop */
    }
    @media (max-width: 768px) {
        .sidebar {
            width: 100%;
            height: 100vh;
        }
        .main-content {
            position: fixed; /* Overlay */
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            transform: translateX(100%); /* Hidden by default */
            transition: transform 0.3s ease-in-out;
            z-index: 2000;
        }
        .main-content.mobile-open {
            transform: translateX(0);
        }
        .close-details {
            display: block;
        }
    }
    .record-details {
        max-width: 800px;
        margin: 0 auto;
    }
    .field {
        margin-bottom: 15px;
    }
    .field label {
        display: block;
        color: #888;
        font-size: 0.9em;
        margin-bottom: 4px;
    }
    .password-row {
        display: flex;
        gap: 10px;
    }
    .password-row input {
        flex: 1;
        background: #333;
        border: 1px solid #444;
        color: #fff;
        padding: 5px;
    }
    pre {
        background: #2d2d2d;
        padding: 10px;
        border-radius: 4px;
        white-space: pre-wrap;
    }
    .empty-state {
        display: flex;
        height: 100%;
        align-items: center;
        justify-content: center;
        color: #666;
    }
</style>
