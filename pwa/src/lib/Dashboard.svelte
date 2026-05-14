<script>
    import { createEventDispatcher } from "svelte";
    import { slide } from "svelte/transition";
    import { dbItems, selectedFile } from "../store.js";

    import {
        getRecordData,
        getDatabaseInfo,
        saveDatabase,
        addRecord,
        updateRecord,
        deleteRecord,
        getDatabaseData,
        searchRecords,
        getAutocompleteSuggestion,
    } from "../wasm.js";
    import Menu from "./Menu.svelte";
    import Modal from "./Modal.svelte";

    import DBInfo from "./DBInfo.svelte";
    import PasswordGenerator from "./PasswordGenerator.svelte";

    const dispatch = createEventDispatcher();

    function autoGrow(node) {
        function resize() {
            node.style.height = 'auto';
            node.style.height = node.scrollHeight + 'px';
        }
        node.addEventListener('input', resize);
        resize();
        return {
            update() { setTimeout(resize, 0); },
            destroy() { node.removeEventListener('input', resize); },
        };
    }

    // Password history field format: fmmnnTLPTLP...
    //   f=enabled(1/0), mm=maxEntries(hex), nn=count(hex)
    //   each entry: T=timestamp(8hex) L=pwLen(4hex) P=password
    function parsePasswordHistory(raw) {
        if (!raw || raw.length < 5) return null;
        const rawBytes = new TextEncoder().encode(raw);
        const decoder = new TextDecoder();
        const readStr = (start, length) => decoder.decode(rawBytes.slice(start, start + length));

        const enabled = readStr(0, 1) === '1';
        const max = parseInt(readStr(1, 2), 16) || 10;
        const count = parseInt(readStr(3, 2), 16);
        const entries = [];
        let pos = 5;
        for (let i = 0; i < count; i++) {
            if (pos + 12 > rawBytes.length) break;
            const timestamp = parseInt(readStr(pos, 8), 16);
            pos += 8;
            const len = parseInt(readStr(pos, 4), 16);
            pos += 4;
            if (pos + len > rawBytes.length) break;
            entries.push({ timestamp, password: readStr(pos, len) });
            pos += len;
        }
        return { enabled, max, entries };
    }

    function serializePasswordHistory(h) {
        const encoder = new TextEncoder();
        const body = h.entries.map(e => {
            const t = Math.floor(e.timestamp).toString(16).padStart(8, '0');
            const l = encoder.encode(e.password).length.toString(16).padStart(4, '0');
            return t + l + e.password;
        }).join('');
        return (h.enabled ? '1' : '0')
            + h.max.toString(16).padStart(2, '0')
            + h.entries.length.toString(16).padStart(2, '0')
            + body;
    }

    function pushPasswordHistory(raw, oldPassword) {
        let h = parsePasswordHistory(raw) ?? { enabled: true, max: 10, entries: [] };
        if (!h.enabled) return raw;
        h.entries.push({ timestamp: Date.now() / 1000, password: oldPassword });
        while (h.entries.length > h.max) h.entries.shift();
        return serializePasswordHistory(h);
    }

    let items = [];
    let filteredItems = [];
    let searchTerm = "";
    let searchNamesOnly = localStorage.getItem('searchNamesOnly') !== 'false';
    let selectedRecord = null;
    let oldTitle = ""; // Track for renames
    let showPassword = false;
    let groupedItems = {};
    let searchInput; // Reference for autofocus
    let copyUserSuccess = false;
    let copyPassSuccess = false;
    let copyUrlSuccess = false;
    let isNewRecord = false;
    let showHistory = false;
    let historyRevealedSet = new Set();

    let isDirty = false;
    let isSaving = false;

    let groupSuggestion = "";
    let groupGhostSuffix = "";
    let usernameSuggestion = "";
    let usernameGhostSuffix = "";

    function clearGhosts() {
        groupSuggestion = "";
        groupGhostSuffix = "";
        usernameSuggestion = "";
        usernameGhostSuffix = "";
    }

    function applySuggestion(field, value) {
        const s = getAutocompleteSuggestion(field, value);
        if (s && s.toLowerCase() !== value.toLowerCase()) {
            return { suggestion: s, ghost: s.slice(value.length) };
        }
        return { suggestion: "", ghost: "" };
    }

    function onGroupInput() {
        const v = selectedRecord.Group;
        if (!v) { groupSuggestion = ""; groupGhostSuffix = ""; return; }
        const r = applySuggestion("group", v);
        groupSuggestion = r.suggestion;
        groupGhostSuffix = r.ghost;
    }

    function onGroupKeydown(e) {
        if (e.key === "Tab" && groupSuggestion) {
            e.preventDefault();
            selectedRecord = { ...selectedRecord, Group: groupSuggestion };
            groupSuggestion = "";
            groupGhostSuffix = "";
        } else if (e.key === "Escape") {
            groupSuggestion = "";
            groupGhostSuffix = "";
        }
    }

    function onUsernameInput() {
        const v = selectedRecord.Username;
        if (!v) { usernameSuggestion = ""; usernameGhostSuffix = ""; return; }
        const r = applySuggestion("username", v);
        usernameSuggestion = r.suggestion;
        usernameGhostSuffix = r.ghost;
    }

    function onUsernameKeydown(e) {
        if (e.key === "Tab" && usernameSuggestion) {
            e.preventDefault();
            selectedRecord = { ...selectedRecord, Username: usernameSuggestion };
            usernameSuggestion = "";
            usernameGhostSuffix = "";
        } else if (e.key === "Escape") {
            usernameSuggestion = "";
            usernameGhostSuffix = "";
        }
    }

    let generator;
    let showGenOptions = false;

    let contextMenu = null; // { x, y, rec }
    function openContextMenu(e, item) {
        e.preventDefault();
        try {
            const rec = getRecordData(item.title);
            contextMenu = { x: e.clientX, y: e.clientY, rec };
        } catch (err) {
            console.error("Context menu: failed to load record", err);
        }
    }

    async function contextCopy(text) {
        contextMenu = null;
        if (!text) return;
        try {
            await navigator.clipboard.writeText(text);
        } catch (err) {
            console.error("Failed to copy", err);
        }
    }

    let collapseAtStartup = localStorage.getItem('collapseAtStartup') === 'true';
    function toggleCollapseAtStartup() {
        collapseAtStartup = !collapseAtStartup;
        localStorage.setItem('collapseAtStartup', String(collapseAtStartup));
    }

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
        if (event.key === "Escape" && contextMenu) {
            contextMenu = null;
            return;
        }
        // Global shortcuts
        if (
            event.key === "/" &&
            !event.ctrlKey &&
            !event.metaKey &&
            !event.altKey
        ) {
            const tag = document.activeElement.tagName.toLowerCase();
            // Ignore if typing in an input or textarea
            if (tag !== "input" && tag !== "textarea") {
                event.preventDefault();
                searchInput.focus();
                searchInput.select();
                return;
            }
        }

        if (!selectedRecord) return;

        if ((event.ctrlKey || event.metaKey) && event.key === "u") {
            event.preventDefault();
            copyToClipboard(selectedRecord.Username, "user");
        } else if ((event.ctrlKey || event.metaKey) && event.key === "p") {
            event.preventDefault();
            copyToClipboard(selectedRecord.Password, "pass");
        } else if ((event.ctrlKey || event.metaKey) && event.key === "o") {
            event.preventDefault();
            if (selectedRecord.URL) {
                window.open(selectedRecord.URL, "_blank");
            }
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
            } else if (type === "url") {
                copyUrlSuccess = true;
                setTimeout(() => (copyUrlSuccess = false), 2000);
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
        if (!searchTerm.trim()) {
            filteredItems = items;
        } else {
            const matchedTitles = new Set(searchRecords(searchTerm, searchNamesOnly));
            filteredItems = items.filter(i => matchedTitles.has(i.title));
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
        console.log("Grouped items keys:", Object.keys(grouped));
    }

    function selectItem(item) {
        console.log("selectItem called for:", item.title);
        try {
            const rec = getRecordData(item.title);
            selectedRecord = rec;
            console.log("Record loaded:", rec ? rec.Title : "null");
            oldTitle = rec.Title; // Store original title
            showPassword = false;
            isNewRecord = false;
            showGenOptions = false;
            clearGhosts();
            showHistory = false;
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
        showGenOptions = false;
        clearGhosts();
        showHistory = false;
        historyRevealedSet = new Set();
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

    async function save(silent = false) {
        if (isSaving) {
            alert("Database is already saving. Please wait.");
            return;
        }
        isSaving = true;
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

            if (!silent) {
                triggerModal({
                    title: "Success",
                    message: "Database saved successfully!",
                    type: "alert",
                });
            }
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
        } finally {
            isSaving = false;
        }
    }

    async function saveRecord() {
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
                const oldRec = getRecordData(oldTitle);
                if (oldRec && oldRec.Password !== selectedRecord.Password && oldRec.Password !== "") {
                    selectedRecord.PasswordHistory = pushPasswordHistory(
                        selectedRecord.PasswordHistory,
                        oldRec.Password,
                    );
                }
                updateRecord(oldTitle, selectedRecord);
            }

            // Refresh list
            const items = getDatabaseData();
            dbItems.set(items);

            // Re-select to refresh state (or update oldTitle)
            oldTitle = selectedRecord.Title;
            isNewRecord = false;
            isDirty = true;

            // Clear search if we updated the title so it doesn't get filtered out if it no longer matches
            if (searchTerm) {
                searchTerm = "";
                filterItems();
            }
        } catch (e) {
            console.error("saveRecord failed:", e);
            alert("Failed to save record: " + e.message);
            return;
        }
        await save(true);
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

    async function performDelete() {
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
            return;
        }
        await save(true);
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
    function handleTreeNavigation(e) {
        if (e.key === "ArrowDown" || e.key === "ArrowUp") {
            e.preventDefault();
            // Find all visible focusable items
            // We need to look at the entire tree from the container perspective
            const tree = document.querySelector(".tree");
            if (!tree) return;

            const focusable = Array.from(
                tree.querySelectorAll('summary, li[tabindex="0"]'),
            );

            const visibleFocusable = focusable.filter((el) => {
                let parent = el.parentElement;
                while (parent && parent !== tree) {
                    // Check both property and attribute for robustness
                    if (
                        parent.tagName === "DETAILS" &&
                        !parent.open &&
                        !parent.hasAttribute("open")
                    )
                        return false;
                    parent = parent.parentElement;
                }
                return true;
            });

            const idx = visibleFocusable.indexOf(e.target);

            if (idx === -1) {
                // Try finding by activeElement if target mismatch
                const idx2 = visibleFocusable.indexOf(document.activeElement);
                if (idx2 !== -1) {
                    // Use idx2
                    if (e.key === "ArrowDown") {
                        const next = visibleFocusable[idx2 + 1];
                        if (next) {
                            next.focus();
                        }
                    } else if (e.key === "ArrowUp") {
                        const prev = visibleFocusable[idx2 - 1];
                        if (prev) {
                            prev.focus();
                        } else if (idx2 === 0) searchInput.focus();
                    }
                    return;
                }
                return;
            }

            if (e.key === "ArrowDown") {
                const next = visibleFocusable[idx + 1];
                if (next) {
                    next.focus();
                } else {
                    // No next item
                }
            } else if (e.key === "ArrowUp") {
                const prev = visibleFocusable[idx - 1];
                if (prev) prev.focus();
                else if (idx === 0) {
                    searchInput.focus();
                }
            }
        }
    }
</script>

<svelte:window on:keydown={handleKeydown} on:click={() => { if (contextMenu) contextMenu = null; }} />

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
                on:save={async () => {
                    showModal = false;
                    isDirty = true; // Mark DB as dirty after info update (though main.go modifies in-memory DB directly too)
                    // Actually, main.go modifies the struct. saveDB() marshals that struct.
                    // So we should mark as dirty.
                    await save(true);
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
                <hr
                    style="border: 0; border-top: 1px solid #444; margin: 5px 0;"
                />
                <button
                    on:click={() => {
                        close();
                        toggleCollapseAtStartup();
                    }}>{collapseAtStartup ? '✓' : '\u00a0\u00a0'} Collapse at startup</button
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
                placeholder={searchNamesOnly ? "Search names…" : "Search details…"}
                bind:value={searchTerm}
                on:input={filterItems}
                on:keydown={(e) => {
                    if (e.key === "Enter") {
                        if (filteredItems.length === 1) {
                            selectItem(filteredItems[0]);
                            // Move focus to details view for accessibility and to satisfy tests
                            // Wait for DOM update
                            setTimeout(() => {
                                const closeBtn =
                                    document.querySelector(".close-details");
                                if (closeBtn) {
                                    closeBtn.focus();
                                } else {
                                    // If for some reason close button isn't there, blur search
                                    e.target.blur();
                                }
                            }, 50);
                        }
                    } else if (e.key === "ArrowDown") {
                        e.preventDefault();
                        const tree = document.querySelector(".tree");
                        if (!tree) {
                            return;
                        }
                        const firstFocusable = tree.querySelector(
                            'summary, li[tabindex="0"]',
                        );
                        if (firstFocusable) {
                            firstFocusable.focus();
                        }
                    } else if (e.key === "ArrowUp") {
                        e.preventDefault();
                    }
                }}
            />
            <label class="scope-label">
                <input
                    type="checkbox"
                    bind:checked={searchNamesOnly}
                    on:change={() => {
                        localStorage.setItem('searchNamesOnly', String(searchNamesOnly));
                        filterItems();
                    }}
                />
                Names only
            </label>
        </div>

        <div class="tree">
            {#each Object.keys(groupedItems) as group}
                <details open={!collapseAtStartup || !!searchTerm}>
                    <summary tabindex="0" on:keydown={handleTreeNavigation}
                        >{group}</summary
                    >
                    <ul role="listbox">
                        {#each groupedItems[group] as item}
                            <li
                                role="option"
                                aria-selected={!!(selectedRecord &&
                                    selectedRecord.Title === item.title)}
                                tabindex="0"
                                class:selected={selectedRecord &&
                                    selectedRecord.Title === item.title}
                                on:click={() => selectItem(item)}
                                on:dblclick={async () => {
                                    try {
                                        const rec = getRecordData(item.title);
                                        if (rec && rec.Password) {
                                            await copyToClipboard(rec.Password, 'pass');
                                        }
                                    } catch (err) {
                                        console.error("Double-click copy failed", err);
                                    }
                                }}
                                on:contextmenu={(e) => openContextMenu(e, item)}
                                on:keydown={(e) => {
                                    if (e.key === "Enter") {
                                        selectItem(item);
                                    } else {
                                        handleTreeNavigation(e);
                                    }
                                }}
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
                    <label for="record-title">Title</label>
                    <input
                        id="record-title"
                        type="text"
                        bind:value={selectedRecord.Title}
                        placeholder="Title"
                    />
                </div>

                <div class="field">
                    <label for="record-group">Group</label>
                    <div class="autocomplete-wrap">
                        {#if groupGhostSuffix}
                            <div class="ghost-overlay" aria-hidden="true">
                                <span class="ghost-typed">{selectedRecord.Group}</span><span class="ghost-suffix">{groupGhostSuffix}</span>
                            </div>
                        {/if}
                        <input
                            id="record-group"
                            type="text"
                            bind:value={selectedRecord.Group}
                            placeholder="Group"
                            on:input={onGroupInput}
                            on:keydown={onGroupKeydown}
                            on:blur={() => { groupSuggestion = ""; groupGhostSuffix = ""; }}
                        />
                    </div>
                </div>

                <div class="field">
                    <button type="button" class="field-label-btn" title="Click to copy" on:click={() => copyToClipboard(selectedRecord.Username, 'user')} on:contextmenu|preventDefault={() => copyToClipboard(selectedRecord.Username, 'user')}>Username</button>
                    <div class="field-row">
                        <div class="autocomplete-wrap">
                            {#if usernameGhostSuffix}
                                <div class="ghost-overlay" aria-hidden="true">
                                    <span class="ghost-typed">{selectedRecord.Username}</span><span class="ghost-suffix">{usernameGhostSuffix}</span>
                                </div>
                            {/if}
                            <input
                                id="record-username"
                                aria-label="Username"
                                type="text"
                                bind:value={selectedRecord.Username}
                                placeholder="Username"
                                on:input={onUsernameInput}
                                on:keydown={onUsernameKeydown}
                                on:blur={() => { usernameSuggestion = ""; usernameGhostSuffix = ""; }}
                            />
                        </div>
                        <button
                            class="icon-btn"
                            on:click={() =>
                                copyToClipboard(
                                    selectedRecord.Username,
                                    "user",
                                )}
                            title="Copy Username (Ctrl+U)"
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
                    <button type="button" class="field-label-btn" title="Click to copy" on:click={() => copyToClipboard(selectedRecord.Password, 'pass')} on:contextmenu|preventDefault={() => copyToClipboard(selectedRecord.Password, 'pass')}>Password</button>
                    <div class="password-row">
                        <div class="password-input-row">
                            <input
                                id="record-password"
                                aria-label="Password"
                                type={showPassword ? "text" : "password"}
                                bind:value={selectedRecord.Password}
                                placeholder="Password"
                            />
                            <button
                                class="icon-btn"
                                on:click={() => copyToClipboard(selectedRecord.Password, "pass")}
                                title="Copy Password (Ctrl+P)"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                            </button>
                            {#if copyPassSuccess}
                                <span class="copy-feedback">Copied!</span>
                            {/if}
                        </div>
                        <div class="password-actions">
                            <button on:click={() => (showPassword = !showPassword)}>
                                {showPassword ? "Hide" : "Show"}
                            </button>
                            <button class="generate-btn" on:click={() => generator.generate()}>
                                Generate
                            </button>
                            <button class="icon-btn" on:click={() => (showGenOptions = !showGenOptions)} title="Password options">
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
                <PasswordGenerator
                    bind:this={generator}
                    bind:showOptions={showGenOptions}
                    on:generate={(e) => {
                        selectedRecord.Password = e.detail;
                        showPassword = true;
                    }}
                >
                    {#if !isNewRecord && selectedRecord.PasswordHistory}
                        {@const hist = parsePasswordHistory(selectedRecord.PasswordHistory)}
                        {#if hist && hist.entries.length > 0}
                            <hr class="panel-divider" />
                            <button
                                type="button"
                                class="history-toggle-btn"
                                on:click={() => (showHistory = !showHistory)}
                            >
                                Show {hist.entries.length} previous password{hist.entries.length === 1 ? '' : 's'}
                                <span>{showHistory ? '▲' : '▶'}</span>
                            </button>
                            {#if showHistory}
                                <div class="history-list" transition:slide={{ duration: 150 }}>
                                    {#each [...hist.entries].reverse() as entry}
                                        <div class="history-entry">
                                            <span class="history-date">{formatDate(new Date(entry.timestamp * 1000).toISOString())}</span>
                                            <span class="history-pw">{entry.password}</span>
                                            <button
                                                class="icon-btn"
                                                title="Copy"
                                                on:click={() => copyToClipboard(entry.password, 'hist')}
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                                            </button>
                                        </div>
                                    {/each}
                                </div>
                            {/if}
                        {/if}
                    {/if}
                </PasswordGenerator>
                <div class="field">
                    <button type="button" class="field-label-btn" title="Click to copy" on:click={() => copyToClipboard(selectedRecord.URL, 'url')} on:contextmenu|preventDefault={() => copyToClipboard(selectedRecord.URL, 'url')}>URL</button>
                    <div class="field-row">
                        <input
                            id="record-url"
                            aria-label="URL"
                            type="text"
                            bind:value={selectedRecord.URL}
                            placeholder="URL"
                        />
                        {#if selectedRecord.URL}
                            <a
                                href={selectedRecord.URL}
                                target="_blank"
                                class="icon-btn"
                                title="Open URL (Ctrl+O)"
                            >
                                ↗
                            </a>
                            <button
                                class="icon-btn"
                                on:click={() => copyToClipboard(selectedRecord.URL, 'url')}
                                title="Copy URL"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                            </button>
                            {#if copyUrlSuccess}
                                <span class="copy-feedback">Copied!</span>
                            {/if}
                        {/if}
                    </div>
                </div>
                <div class="field">
                    <label for="record-notes">Notes</label>
                    <textarea
                        id="record-notes"
                        bind:value={selectedRecord.Notes}
                        placeholder="Notes"
                        use:autoGrow={selectedRecord}
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

{#if contextMenu}
    <div
        class="context-menu"
        role="menu"
        tabindex="-1"
        style="left:{contextMenu.x}px;top:{contextMenu.y}px"
        on:click|stopPropagation
        on:keydown|stopPropagation
    >
        <button on:click={() => contextCopy(contextMenu.rec.Username)}>
            Copy Username
        </button>
        <button on:click={() => contextCopy(contextMenu.rec.Password)}>
            Copy Password
        </button>
        {#if contextMenu.rec.URL}
            <button on:click={() => contextCopy(contextMenu.rec.URL)}>
                Copy URL
            </button>
            <button on:click={() => { const url = contextMenu.rec.URL; contextMenu = null; window.open(url, '_blank'); }}>
                Open URL
            </button>
        {/if}
    </div>
{/if}

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
    .field-label-btn {
        display: block;
        background: none;
        border: none;
        color: #888;
        font-size: 0.9em;
        margin-bottom: 6px;
        padding: 0;
        cursor: pointer;
        font-family: inherit;
        text-align: left;
    }
    .field-label-btn:hover {
        color: #bbb;
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
        flex-direction: column;
        gap: 6px;
    }
    .password-input-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }
    .password-input-row input {
        flex: 1;
        min-width: 0;
        width: auto;
    }
    .password-actions {
        display: flex;
        align-items: center;
        gap: 8px;
    }
    textarea {
        background: #2d2d2d;
        padding: 10px;
        border-radius: 4px;
        white-space: pre-wrap;
        font-family: inherit;
        resize: none;
        line-height: 1.5;
        min-height: calc(5 * 1.5em + 20px);
        max-height: calc(20 * 1.5em + 20px);
        overflow-y: auto;
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
    .panel-divider {
        border: none;
        border-top: 1px solid #3a3a3a;
        margin: 8px 0 6px;
    }
    .history-toggle-btn {
        display: flex;
        align-items: center;
        justify-content: space-between;
        width: 100%;
        background: none;
        border: none;
        color: #888;
        font-size: 0.8em;
        padding: 2px 0;
        cursor: pointer;
        font-family: inherit;
        text-align: left;
    }
    .history-toggle-btn:hover {
        color: #bbb;
    }
    .history-list {
        margin-top: 6px;
        border: 1px solid #333;
        border-radius: 4px;
        overflow: hidden;
    }
    .history-entry {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 6px 10px;
        border-bottom: 1px solid #2a2a2a;
        font-size: 0.85em;
    }
    .history-entry:last-child {
        border-bottom: none;
    }
    .history-entry:nth-child(odd) {
        background: #252525;
    }
    .history-date {
        color: #666;
        white-space: nowrap;
        flex-shrink: 0;
    }
    .history-pw {
        flex: 1;
        font-family: monospace;
        color: #ccc;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
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
    .context-menu {
        position: fixed;
        z-index: 2000;
        background: #252526;
        border: 1px solid #444;
        border-radius: 4px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
        padding: 4px 0;
        min-width: 160px;
    }
    .context-menu button {
        display: block;
        width: 100%;
        text-align: left;
        background: none;
        border: none;
        color: #e0e0e0;
        padding: 8px 14px;
        cursor: pointer;
        font-size: 0.9em;
    }
    .context-menu button:hover {
        background: #37373d;
    }
    .autocomplete-wrap {
        position: relative;
        display: block;
    }
    .autocomplete-wrap input {
        width: 100%;
    }
    .field-row .autocomplete-wrap {
        flex: 1;
        min-width: 0;
    }
    .field-row .autocomplete-wrap input {
        width: 100%;
    }
    .ghost-overlay {
        position: absolute;
        inset: 0;
        padding: 8px;
        pointer-events: none;
        font-size: 1rem;
        font-family: inherit;
        line-height: 1.5;
        white-space: pre;
        overflow: hidden;
        border: 1px solid transparent;
        border-radius: 4px;
        z-index: 1;
        display: flex;
        align-items: center;
    }
    .ghost-typed {
        color: transparent;
    }
    .ghost-suffix {
        color: #666;
    }
</style>
