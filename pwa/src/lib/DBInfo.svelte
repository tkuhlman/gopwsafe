<script>
    import { createEventDispatcher } from "svelte";
    import { updateDBInfo } from "../wasm.js";

    export let info = {};
    export let filename = "";

    const dispatch = createEventDispatcher();

    let name = info.name || "";
    let description = info.description || "";

    function save() {
        try {
            updateDBInfo(name, description);
            dispatch("save");
        } catch (e) {
            console.error(e);
            alert("Failed to update DB info: " + e.message);
        }
    }
</script>

<div class="db-info">
    <div class="field">
        <label for="dbinfo-filename">Filename</label>
        <input
            id="dbinfo-filename"
            type="text"
            value={filename}
            readonly
            disabled
        />
    </div>

    <div class="field">
        <label for="dbinfo-name">Name</label>
        <input
            id="dbinfo-name"
            type="text"
            bind:value={name}
            placeholder="Database Name"
        />
    </div>

    <div class="field">
        <label for="dbinfo-desc">Description</label>
        <textarea
            id="dbinfo-desc"
            bind:value={description}
            rows="3"
            placeholder="Description"
        ></textarea>
    </div>

    <div class="meta-grid">
        <div class="meta-item">
            <label>Version</label>
            <span>{info.version}</span>
        </div>
        <div class="meta-item">
            <label>UUID</label>
            <span class="uuid">{info.uuid}</span>
        </div>
        <div class="meta-item">
            <label>Last Saved By</label>
            <span>{info.who} @ {info.what || "Unknown"}</span>
        </div>
        <div class="meta-item">
            <label>Last Saved Time</label>
            <span>{info.when}</span>
        </div>
    </div>

    <div class="actions">
        <button class="primary" on:click={save}>Save</button>
    </div>
</div>

<style>
    .db-info {
        display: flex;
        flex-direction: column;
        gap: 15px;
    }
    .field label,
    .meta-item label {
        display: block;
        color: #888;
        font-size: 0.9em;
        margin-bottom: 4px;
    }
    input[type="text"],
    textarea {
        width: 100%;
        padding: 8px;
        background: #333;
        border: 1px solid #444;
        color: #fff;
        border-radius: 4px;
        font-size: 1rem;
    }
    input:disabled {
        background: #2a2a2a;
        color: #aaa;
        cursor: not-allowed;
    }
    .meta-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 10px;
        background: #1e1e1e;
        padding: 10px;
        border-radius: 4px;
        margin-top: 10px;
    }
    .meta-item {
        overflow: hidden;
    }
    .meta-item span {
        display: block;
        color: #ddd;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    .uuid {
        font-family: monospace;
        font-size: 0.9em;
    }
    .actions {
        display: flex;
        justify-content: flex-end;
        margin-top: 10px;
    }
    button.primary {
        background: #007bff;
        color: white;
        border: none;
        padding: 8px 16px;
        border-radius: 4px;
        cursor: pointer;
    }
    button.primary:hover {
        background: #0056b3;
    }
</style>
