export async function loadWasm() {
    const go = new Go();

    // Fetch and decompress
    const response = await fetch("/gopwsafe.wasm.gz");
    if (!response.ok) {
        console.error("Failed to fetch WASM:", response.status, response.statusText, response.url);
        throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    let instance;
    try {
        // Try streaming instantiation. This works if the server sends the correct Content-Type (application/wasm)
        // and handles compression (Content-Encoding: gzip) transparently, which Vite does.
        const result = await WebAssembly.instantiateStreaming(response, go.importObject);
        instance = result.instance;
    } catch (e) {
        console.warn("instantiateStreaming failed, falling back to arrayBuffer", e);
        // Fallback: This might be needed if Content-Type is wrong or other streaming issues.
        // We clone headers? No, response body is used. If streaming failed mid-way, response might be disturbed.
        // But usually instantiateStreaming checks mime type first.
        // If response is disturbed we can't retry. Ideally we should have cloned it if we wanted to retry.
        // But simpler: just throw/log for now or try a fresh fetch if we really wanted to be robust.
        // For now, let's assume if streaming fails, we might need to handle the "manual decompression" case 
        // ONLY if we know it failed due to "magic header" (meaning it was compressed but browser didn't decompress).

        // HOWEVER, since we know Vite serves it correctly, let's stick to the standard path.
        // If we really need to support raw .gz serving without headers, we would check the error or try-catch block differently.
        throw e;
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
