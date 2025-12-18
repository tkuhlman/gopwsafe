<script>
    import { createEventDispatcher } from "svelte";
    import { dbItems, selectedFile } from "../store.js";

    import {
        getRecordData,
        getDatabaseInfo,
        saveDatabase,
        addRecord,
        updateRecord,
        deleteRecord,
        getDatabaseData,
    } from "../wasm.js";
    import Menu from "./Menu.svelte";
    import Modal from "./Modal.svelte";

    import DBInfo from "./DBInfo.svelte";

    const dispatch = createEventDispatcher();

    let items = [];
    let filteredItems = [];
    let searchTerm = "";
    let selectedRecord = null;
    let oldTitle = ""; // Track for renames
    let showPassword = false;
    let groupedItems = {};
    let searchInput; // Reference for autofocus
    let copyUserSuccess = false;
    let copyPassSuccess = false;
    let isNewRecord = false;

    let isDirty = false;

    let showModal = false;
    let modalConfig = {
        title: "",
        message: "",
        type: "confirm",
        confirmLabel: "OK",
        cancelLabel: "Cancel",
        onConfirm: () => {},
    };

    function triggerModal(config) {
        modalConfig = {
            confirmLabel: "OK",
            cancelLabel: "Cancel",
            type: "confirm",
            ...config,
        };
        showModal = true;
    }

    function handleKeydown(event) {
        if (!selectedRecord) return;

        if ((event.ctrlKey || event.metaKey) && event.key === "u") {
            event.preventDefault();
            copyToClipboard(selectedRecord.Username, "user");
        } else if ((event.ctrlKey || event.metaKey) && event.key === "p") {
            event.preventDefault();
            copyToClipboard(selectedRecord.Password, "pass");
        }
    }

    async function copyToClipboard(text, type) {
        try {
            await navigator.clipboard.writeText(text);
            if (type === "user") {
                copyUserSuccess = true;
                setTimeout(() => (copyUserSuccess = false), 2000);
            } else if (type === "pass") {
                copyPassSuccess = true;
                setTimeout(() => (copyPassSuccess = false), 2000);
            }
        } catch (err) {
            console.error("Failed to copy!", err);
        }
    }

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
            const rec = getRecordData(item.title);
            selectedRecord = rec;
            oldTitle = rec.Title; // Store original title
            showPassword = false;
            isNewRecord = false;
        } catch (e) {
            console.error(e);
            alert("Failed to load record details");
        }
    }

    function createNewRecord() {
        // Template for new record
        selectedRecord = {
            Title: "New Record",
            Group: "",
            Username: "",
            Password: "",
            URL: "",
            Notes: "",
            // Add other fields as necessary with defaults
            UUID: Array(16).fill(0),
            CreateTime: new Date().toISOString(),
            ModTime: new Date().toISOString(),
        };
        oldTitle = "";
        showPassword = true;
        isNewRecord = true;
    }

    // Bind this to the new record event from the menu
    $: {
        // This is a bit of a hack to listen to events from Menu if passed via props,
        // but here Menu is a component in the markup.
        // We'll handle the event in the markup.
    }

    function formatDate(str) {
        if (!str) return "";
        try {
            return new Date(str).toLocaleString();
        } catch (e) {
            return str;
        }
    }

    async function save() {
        try {
            const data = saveDatabase(); // Uint8Array
            let handle = $selectedFile ? $selectedFile.handle : null;

            if (!handle) {
                // Save As
                handle = await window.showSaveFilePicker({
                    suggestedName: $selectedFile
                        ? $selectedFile.name
                        : "pwsafe.psafe3",
                    types: [
                        {
                            description: "Password Safe DB",
                            accept: {
                                "application/octet-stream": [".psafe3", ".dat"],
                            },
                        },
                    ],
                });
            }

            // Write to file
            const writable = await handle.createWritable();
            await writable.write(data);
            await writable.close();

            triggerModal({
                title: "Success",
                message: "Database saved successfully!",
                type: "alert",
            });
            isDirty = false;

            // update store if it was a new file
            if (!$selectedFile || $selectedFile.handle !== handle) {
                selectedFile.update((s) => ({
                    ...s,
                    handle: handle,
                    name: handle.name,
                }));
            }
        } catch (e) {
            console.error("Save failed", e);
            if (e.name !== "AbortError") {
                alert("Failed to save: " + e.message);
            }
        }
    }

    function saveRecord() {
        try {
            if (!selectedRecord.Title) {
                alert("Title is required");
                return;
            }

            // Update mod time
            selectedRecord.ModTime = new Date().toISOString();

            if (isNewRecord) {
                addRecord(selectedRecord);
            } else {
                updateRecord(oldTitle, selectedRecord);
            }

            // Refresh list
            const items = getDatabaseData();
            dbItems.set(items);

            // Re-select to refresh state (or update oldTitle)
            oldTitle = selectedRecord.Title;
            isNewRecord = false;
            isDirty = true;
        } catch (e) {
            console.error(e);
            alert("Failed to save record: " + e.message);
        }
    }

    function deleteCurrentRecord() {
        triggerModal({
            title: "Delete Record",
            message: `Are you sure you want to delete "${selectedRecord.Title}"?`,
            type: "danger",
            confirmLabel: "Delete",
            onConfirm: () => {
                performDelete();
            },
        });
    }

    function performDelete() {
        try {
            deleteRecord(selectedRecord.Title);
            selectedRecord = null;
            isNewRecord = false;
            isDirty = true;

            // Refresh list
            const items = getDatabaseData();
            dbItems.set(items);
        } catch (e) {
            console.error(e);
            alert("Failed to delete record: " + e.message);
        }
    }

    function showDBInfo() {
        try {
            const info = getDatabaseInfo();
            triggerModal({
                title: "Database Info",
                component: DBInfo,
                props: {
                    info: info,
                    filename: $selectedFile ? $selectedFile.name : "",
                },
                confirmLabel: "Close", // Or hide the confirm button? DBInfo has its own Save button.
                // If we want to hide Modal footer buttons, we might need more Modal config.
                // For now, "Close" acts as cancel/close.
                type: "info", // "info" isn't a standard type in Modal yet, but it falls back to primary/alert logic maybe?
                // Actually Modal type controls button styles.
                // Let's use 'alert' type so we only have one button effectively?
                // Waait, DBInfo has a "Save" button inside it.
                // If the user clicks Save in DBInfo, it dispatches 'save'.
                // We should probably just have a "Close" button in the modal footer.
                type: "alert",
                confirmLabel: "Close",
            });
        } catch (e) {
            console.error(e);
            alert("Failed to get DB info: " + e.message);
        }
    }

    function closeDb() {
        if (isDirty) {
            triggerModal({
                title: "Unsaved Changes",
                message:
                    "You have unsaved changes. Are you sure you want to close without saving?",
                confirmLabel: "Close without saving",
                type: "confirm",
                onConfirm: () => {
                    dispatch("close");
                    isDirty = false;
                },
            });
            return;
        }
        dispatch("close");
        isDirty = false;
    }

    // Warn on tab close
    window.addEventListener("beforeunload", (e) => {
        if (isDirty) {
            e.preventDefault();
            e.returnValue = "";
        }
    });
</script>

<svelte:window on:keydown={handleKeydown} />

{#if showModal}
    <Modal
        title={modalConfig.title}
        message={modalConfig.message}
        type={modalConfig.type}
        confirmLabel={modalConfig.confirmLabel}
        cancelLabel={modalConfig.cancelLabel}
        on:confirm={() => {
            if (modalConfig.onConfirm) modalConfig.onConfirm();
            showModal = false;
        }}
        on:cancel={() => (showModal = false)}
    >
        {#if modalConfig.component}
            <svelte:component
                this={modalConfig.component}
                {...modalConfig.props}
                on:save={() => {
                    showModal = false;
                    isDirty = true; // Mark DB as dirty after info update (though main.go modifies in-memory DB directly too)
                    // Actually, main.go modifies the struct. saveDB() marshals that struct.
                    // So we should mark as dirty.
                    triggerModal({
                        title: "Success",
                        message:
                            "Detail updated. Don't forget to save the database file.",
                        type: "alert",
                    });
                }}
            />
        {:else}
            <p>{modalConfig.message}</p>
        {/if}
    </Modal>
{/if}

<div class="dashboard">
    <div class="sidebar">
        <div class="toolbar">
            <Menu let:close>
                <button
                    on:click={() => {
                        close();
                        createNewRecord();
                    }}>New Record</button
                >
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
                        closeDb();
                    }}>Close DB</button
                >
            </Menu>
            <!-- Visual Indicator for Dirty State (e.g. dot on menu or title?) 
                 Since we don't have a title bar here (it's in toolbar), maybe add a dot next to Menu?
                 Or just next to Save DB button inside?
            -->
            {#if isDirty}
                <span class="dirty-indicator" title="Unsaved Changes">●</span>
            {/if}

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
                    <h2>{isNewRecord ? "New Record" : selectedRecord.Title}</h2>
                </div>

                <div class="field">
                    <label>Title</label>
                    <input
                        type="text"
                        bind:value={selectedRecord.Title}
                        placeholder="Title"
                    />
                </div>

                <div class="field">
                    <label>Group</label>
                    <input
                        type="text"
                        bind:value={selectedRecord.Group}
                        placeholder="Group"
                    />
                </div>

                <div class="field">
                    <label>Username</label>
                    <div class="field-row">
                        <input
                            type="text"
                            bind:value={selectedRecord.Username}
                            placeholder="Username"
                        />
                        <button
                            class="icon-btn"
                            on:click={() =>
                                copyToClipboard(
                                    selectedRecord.Username,
                                    "user",
                                )}
                            title="Copy Username"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                ><rect
                                    x="9"
                                    y="9"
                                    width="13"
                                    height="13"
                                    rx="2"
                                    ry="2"
                                ></rect><path
                                    d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"
                                ></path></svg
                            >
                        </button>
                        {#if copyUserSuccess}
                            <span class="copy-feedback">Copied!</span>
                        {/if}
                    </div>
                </div>
                <div class="field">
                    <label>Password</label>
                    <div class="password-row">
                        <input
                            type={showPassword ? "text" : "password"}
                            bind:value={selectedRecord.Password}
                            placeholder="Password"
                        />
                        <button on:click={() => (showPassword = !showPassword)}>
                            {showPassword ? "Hide" : "Show"}
                        </button>
                        <button
                            class="icon-btn"
                            on:click={() =>
                                copyToClipboard(
                                    selectedRecord.Password,
                                    "pass",
                                )}
                            title="Copy Password"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                ><rect
                                    x="9"
                                    y="9"
                                    width="13"
                                    height="13"
                                    rx="2"
                                    ry="2"
                                ></rect><path
                                    d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"
                                ></path></svg
                            >
                        </button>
                        {#if copyPassSuccess}
                            <span class="copy-feedback">Copied!</span>
                        {/if}
                    </div>
                </div>
                <div class="field">
                    <label>URL</label>
                    <div class="field-row">
                        <input
                            type="text"
                            bind:value={selectedRecord.URL}
                            placeholder="URL"
                        />
                        {#if selectedRecord.URL}
                            <a
                                href={selectedRecord.URL}
                                target="_blank"
                                class="icon-btn"
                                title="Open URL"
                            >
                                ↗
                            </a>
                        {/if}
                    </div>
                </div>
                <div class="field">
                    <label>Notes</label>
                    <textarea
                        bind:value={selectedRecord.Notes}
                        rows="5"
                        placeholder="Notes"
                    ></textarea>
                </div>

                <div class="actions-row">
                    <button class="primary" on:click={saveRecord}
                        >Save Record</button
                    >
                    {#if !isNewRecord}
                        <button class="danger" on:click={deleteCurrentRecord}
                            >Delete Record</button
                        >
                    {/if}
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
        margin-bottom: 20px;
    }
    .field label {
        display: block;
        color: #888;
        font-size: 0.9em;
        margin-bottom: 6px;
    }
    .field input[type="text"],
    .field input[type="password"],
    .field textarea {
        width: 100%;
        padding: 8px;
        background: #333;
        border: 1px solid #444;
        color: #fff;
        border-radius: 4px;
        font-size: 1rem;
    }
    .field input:focus,
    .field textarea:focus {
        border-color: #007bff;
        outline: none;
    }
    .password-row {
        display: flex;
        gap: 10px;
    }
    .password-row input {
        flex: 1;
        width: auto;
        min-width: 0;
    }
    pre,
    textarea {
        background: #2d2d2d;
        padding: 10px;
        border-radius: 4px;
        white-space: pre-wrap;
        font-family: inherit;
        resize: vertical;
    }
    .empty-state {
        display: flex;
        height: 100%;
        align-items: center;
        justify-content: center;
        color: #666;
    }
    .field-row {
        display: flex;
        align-items: center;
        gap: 10px;
    }
    .field-row input {
        flex: 1;
        width: auto; /* Override the 100% from general input selector */
        min-width: 0;
    }
    .icon-btn {
        background: none;
        border: none;
        color: #ccc;
        cursor: pointer;
        padding: 4px;
        display: flex;
        align-items: center;
        border-radius: 4px;
    }
    .icon-btn:hover {
        background: #333;
        color: #fff;
    }
    .copy-feedback {
        color: #4caf50;
        font-size: 0.9em;
        animation: fadeOut 2s forwards;
    }
    .actions-row {
        margin-top: 30px;
        display: flex;
        gap: 10px;
        justify-content: flex-end;
    }
    button.primary {
        background: #007bff;
        color: white;
        border: none;
        padding: 10px 20px;
        border-radius: 4px;
        cursor: pointer;
        font-size: 1rem;
    }
    button.primary:hover {
        background: #0056b3;
    }
    button.danger {
        background: #dc3545;
        color: white;
        border: none;
        padding: 10px 20px;
        border-radius: 4px;
        cursor: pointer;
        font-size: 1rem;
    }
    button.danger:hover {
        background: #a71d2a;
    }

    @keyframes fadeOut {
        0% {
            opacity: 1;
        }
        70% {
            opacity: 1;
        }
        100% {
            opacity: 0;
        }
    }
</style>
