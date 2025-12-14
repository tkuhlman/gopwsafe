import { writable } from 'svelte/store';

export const selectedFile = writable(null); // { handle, name }
export const dbItems = writable([]); // Array of { uuid, title, group }
