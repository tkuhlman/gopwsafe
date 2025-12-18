export async function loadWasm() {
    const go = new Go();

    // Fetch the WASM file
    // Note: On GitHub Pages, .gz files are served as application/gzip without Content-Encoding: gzip
    // On local Vite dev server, it serves with Content-Encoding: gzip
    const response = await fetch("/gopwsafe/gopwsafe.wasm.gz");

    if (!response.ok) {
        console.error("Failed to fetch WASM:", response.status, response.statusText, response.url);
        throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    let instance;
    try {
        // Check if the browser already handled the decompression (local dev)
        // or if we need to do it manually (GitHub Pages)
        let source = response;
        const contentEncoding = response.headers.get("Content-Encoding");

        // If it's gzipped and NOT automatically decompressed (no content-encoding header acting as transport encoding),
        // or if it explicitly says it is gzipped content that hasn't been undone.
        // GitHub pages just serves the file. Let's look at the magic number or filename if we really want to be sure,
        // but simpler is to try/catch or just checking if we can stream it.

        // Strategy:
        // 1. If we are on a platform that serves correct headers (Vite), instantiateStreaming works.
        // 2. If we are on GH Pages, it serves as a binary blob.

        // Let's try to detect if we need to decompress.
        // A simple heuristic: if the URL ends in .gz and Content-Encoding is NOT gzip, we probably need to decompress.
        if (response.url.endsWith(".gz") && contentEncoding !== "gzip") {
            const ds = new DecompressionStream("gzip");
            const decompressedStream = response.body.pipeThrough(ds);
            // Create a new Response with the decompressed stream and correct content type
            source = new Response(decompressedStream, { headers: { "Content-Type": "application/wasm" } });
        }

        const result = await WebAssembly.instantiateStreaming(source, go.importObject);
        instance = result.instance;
    } catch (e) {
        console.warn("instantiateStreaming failed, trying fallback", e);
        // Fallback for environments where streaming fails or manual decompression setup above failed
        // We will simple fetch, arrayBuffer, (decompress if needed), and instantiate.
        // Since we already consumed the body in the try block if we piped it, we can't easily retry with the same response object.
        // So we might need to re-fetch or just handle the error.

        // Retrying with a fresh fetch for the fallback is safest.
        // This is a robust fallback for "everything else".
        const response2 = await fetch("/gopwsafe/gopwsafe.wasm.gz");
        let buffer = await response2.arrayBuffer();

        // Manual magic bytes check for GZIP (1f 8b)
        const view = new Uint8Array(buffer);
        if (view[0] === 0x1f && view[1] === 0x8b) {
            console.log("Manual decompression required for fallback");
            const ds = new DecompressionStream("gzip");
            const stream = new Blob([buffer]).stream().pipeThrough(ds);
            buffer = await new Response(stream).arrayBuffer();
        }

        const result = await WebAssembly.instantiate(buffer, go.importObject);
        instance = result.instance;
    }

    go.run(instance);
    console.log("WASM loaded");
}

export function openDatabase(fileData, password) {
    // fileData should be Uint8Array
    const err = window.openDB(fileData, password);
    if (err) {
        throw new Error(err);
    }
}

export function getDatabaseData() {
    const res = window.getDBData();
    // primitive error handling based on string return
    if (typeof res === 'string' && (res.startsWith("database not open") || res.startsWith("json marshal error"))) {
        throw new Error(res);
    }
    const parsed = JSON.parse(res);
    return parsed || [];
}

export function getRecordData(title) {
    const res = window.getRecord(title);
    if (typeof res === 'string' && (res === "record not found" || res === "database not open")) {
        throw new Error(res);
    }
    return JSON.parse(res);
}

export function createDatabase(password) {
    const err = window.createDatabase(password);
    if (err) {
        throw new Error(err);
    }
}

export function getDatabaseInfo() {
    const res = window.getDBInfo();
    if (typeof res === 'string' && res.startsWith("database not open")) {
        throw new Error(res);
    }
    // Try to parse JSON.
    if (typeof res === 'string' && res.startsWith("json marshal error")) {
        throw new Error(res);
    }
    return JSON.parse(res);
}

export function saveDatabase() {
    const res = window.saveDB();
    if (typeof res === 'string') {
        throw new Error(res);
    }
    return res; // Uint8Array
}

export function addRecord(record) {
    // record is object, convert to JSON string
    const json = JSON.stringify(record);
    const err = window.addRecord(json);
    if (err) {
        throw new Error(err);
    }
}

export function updateRecord(oldTitle, record) {
    const json = JSON.stringify(record);
    const err = window.updateRecord(oldTitle, json);
    if (err) {
        throw new Error(err);
    }
}


export function deleteRecord(title) {
    const err = window.deleteRecord(title);
    if (err) {
        throw new Error(err);
    }
}

export function updateDBInfo(name, description) {
    const err = window.updateDBInfo(name, description);
    if (err) {
        throw new Error(err);
    }
}
