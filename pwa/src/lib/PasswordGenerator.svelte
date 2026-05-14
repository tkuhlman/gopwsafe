<script>
    import { createEventDispatcher } from 'svelte';
    import { slide } from 'svelte/transition';
    const dispatch = createEventDispatcher();

    export let showOptions = false;

    const DEFAULT_EXCLUDED = "\"'<>\\`";

    const GROUPS = [
        { id: 'upper',   label: 'A–Z',     chars: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('') },
        { id: 'lower',   label: 'a–z',     chars: 'abcdefghijklmnopqrstuvwxyz'.split('') },
        { id: 'digits',  label: '0–9',     chars: '0123456789'.split('') },
        { id: 'symbols', label: 'Symbols', chars: '!"#$%&\'()*+,-./:;<=>?@[\\]^_`{|}~'.split('') },
    ];

    const VALID_STATES = new Set(['required', 'included', 'excluded']);

    function makeDefault() {
        return {
            length: 20,
            excludedChars: DEFAULT_EXCLUDED,
            groups: {
                upper:   { state: 'required' },
                lower:   { state: 'required' },
                digits:  { state: 'included' },
                symbols: { state: 'included' },
            }
        };
    }

    function load() {
        try {
            const raw = localStorage.getItem('pwgen');
            if (raw) {
                const p = JSON.parse(raw);
                if (typeof p.length === 'number' && typeof p.excludedChars === 'string') {
                    for (const g of GROUPS) {
                        if (!p.groups?.[g.id] || !VALID_STATES.has(p.groups[g.id].state)) {
                            return makeDefault();
                        }
                    }
                    return p;
                }
            }
        } catch (_) {}
        return makeDefault();
    }

    let s = load();
    let error = '';

    function persist() {
        localStorage.setItem('pwgen', JSON.stringify(s));
    }

    function reset() {
        s = makeDefault();
        error = '';
        persist();
    }

    function setGroupState(id, state) {
        s.groups[id].state = state;
        s = s;
        error = '';
        persist();
    }

    function updateLength(e) {
        const v = parseInt(e.target.value, 10);
        if (v >= 4 && v <= 64) {
            s.length = v;
            persist();
        } else {
            e.target.value = Math.min(64, Math.max(4, v || 4));
        }
    }

    function updateExcluded(e) {
        const seen = new Set();
        const filtered = [];
        for (const c of e.target.value) {
            const code = c.charCodeAt(0);
            if (code >= 0x21 && code <= 0x7E && !seen.has(c)) {
                seen.add(c);
                filtered.push(c);
            }
        }
        s.excludedChars = filtered.join('');
        e.target.value = s.excludedChars;
        s = s;
        error = '';
        persist();
    }

    function randInt(max) {
        const arr = new Uint32Array(1);
        const limit = Math.floor(0x100000000 / max) * max;
        let r;
        do { crypto.getRandomValues(arr); r = arr[0]; } while (r >= limit);
        return r % max;
    }

    export function generate() {
        error = '';
        const excluded = new Set(s.excludedChars);
        let pool = [];
        let guaranteed = [];

        for (const g of GROUPS) {
            const gs = s.groups[g.id];
            if (gs.state === 'excluded') continue;
            const active = g.chars.filter(c => !excluded.has(c));
            if (!active.length) continue;
            pool.push(...active);
            if (gs.state === 'required') {
                guaranteed.push(active[randInt(active.length)]);
            }
        }

        if (!pool.length) {
            error = 'No characters available — adjust settings';
            return;
        }

        const len = s.length;
        const result = guaranteed.slice(0, len);
        while (result.length < len) result.push(pool[randInt(pool.length)]);

        for (let i = result.length - 1; i > 0; i--) {
            const j = randInt(i + 1);
            [result[i], result[j]] = [result[j], result[i]];
        }
        dispatch('generate', result.join(''));
    }
</script>

{#if showOptions}
<div class="pwgen-panel" transition:slide={{ duration: 200 }}>
    <div class="panel-header">
        <span class="panel-title">Password options</span>
        <label class="length-label">
            Length
            <input
                type="number"
                min="4"
                max="64"
                value={s.length}
                on:change={updateLength}
                on:blur={updateLength}
                class="length-input"
            />
        </label>
        <button class="reset-btn" on:click={reset} title="Reset to defaults">↺ Reset</button>
    </div>

    {#each GROUPS as g}
        {@const gs = s.groups[g.id]}
        <div class="group">
            <span class="group-label">{g.label}</span>
            <div class="state-bar">
                {#each ['required', 'included', 'excluded'] as state}
                    <button
                        class="state-btn state-{state}"
                        class:active={gs.state === state}
                        on:click={() => setGroupState(g.id, state)}
                    >{state}</button>
                {/each}
            </div>
        </div>
    {/each}

    <div class="exclude-row">
        <label class="exclude-label" for="exclude-input">Excluded</label>
        <input
            id="exclude-input"
            class="exclude-input"
            type="text"
            value={s.excludedChars}
            on:input={updateExcluded}
            spellcheck="false"
            autocomplete="off"
        />
    </div>

    {#if error}
        <div class="error-msg">{error}</div>
    {/if}

    <slot />
</div>
{/if}

<style>
    .pwgen-panel {
        border: 1px solid #444;
        border-radius: 4px;
        padding: 8px 10px;
        background: #2a2a2a;
        margin-top: 6px;
        margin-bottom: 20px;
    }
    .panel-header {
        display: flex;
        align-items: center;
        gap: 10px;
        margin-bottom: 8px;
    }
    .panel-title {
        font-size: 0.75em;
        color: #888;
        font-weight: bold;
        flex: 1;
    }
    .length-label {
        display: flex;
        align-items: center;
        gap: 6px;
        color: #ccc;
        font-size: 0.85em;
    }
    .length-input {
        width: 52px;
        padding: 3px;
        background: #333;
        border: 1px solid #555;
        color: #fff;
        border-radius: 4px;
        text-align: center;
        font-size: 0.85em;
    }
    .reset-btn {
        background: none;
        border: 1px solid #555;
        color: #888;
        cursor: pointer;
        font-size: 0.72em;
        padding: 2px 7px;
        border-radius: 3px;
        line-height: 1.4;
    }
    .reset-btn:hover {
        color: #ccc;
        border-color: #888;
    }
    .group {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 5px;
    }
    .group-label {
        font-size: 0.8em;
        color: #aaa;
        font-weight: bold;
        min-width: 52px;
    }
    .state-bar {
        display: flex;
        border: 1px solid #555;
        border-radius: 3px;
    }
    .state-btn {
        background: none;
        border: none;
        border-right: 1px solid #555;
        border-radius: 0;
        color: #666;
        font-size: 0.72em;
        font-weight: 500;
        padding: 3px 9px;
        cursor: pointer;
        white-space: nowrap;
    }
    .state-btn:first-child {
        border-radius: 2px 0 0 2px;
    }
    .state-btn:last-child {
        border-right: none;
        border-radius: 0 2px 2px 0;
    }
    .state-btn:hover:not(.active) {
        background: #3a3a3a;
        color: #aaa;
    }
    .state-btn.state-required.active {
        background: #1a3a1a;
        color: #5cca60;
    }
    .state-btn.state-included.active {
        background: #2d2d2d;
        color: #d0d0d0;
    }
    .state-btn.state-excluded.active {
        background: #3a1a1a;
        color: #ef6060;
    }
    .exclude-row {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-top: 7px;
    }
    .exclude-label {
        font-size: 0.75em;
        color: #e05555;
        font-weight: bold;
        min-width: 52px;
    }
    .exclude-input {
        flex: 1;
        background: #2a1a1a;
        border: 1px solid #7a3333;
        color: #e05555;
        border-radius: 4px;
        padding: 3px 6px;
        font-family: monospace;
        font-size: 0.85em;
        letter-spacing: 0.08em;
    }
    .exclude-input:focus {
        outline: none;
        border-color: #e05555;
    }
    .error-msg {
        margin-top: 6px;
        font-size: 0.75em;
        color: #e05555;
        text-align: center;
    }
</style>
