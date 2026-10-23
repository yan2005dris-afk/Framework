import { openDB, type DBSchema, type IDBPDatabase } from "idb";
import type { Producto } from "@/data/productos";
import type { Referido } from "@/data/red";

const DB_NAME = "multicatalogo_offline_db";
const DB_VERSION = 2;
const STORE_PRODUCTOS = "productos";
const STORE_RED = "red";

interface MultiCatalogoDB extends DBSchema {
  productos: {
    key: number;
    value: Producto;
    indexes: { "by-categoria": string };
  };
  red: {
    key: string;
    value: { id: string; data: Referido };
  };
}

let dbPromise: Promise<IDBPDatabase<MultiCatalogoDB>> | null = null;

const getDb = () => {
  if (!dbPromise) {
    dbPromise = openDB<MultiCatalogoDB>(DB_NAME, DB_VERSION, {
      upgrade(db, oldVersion) {
        if (oldVersion < 1 || !db.objectStoreNames.contains(STORE_PRODUCTOS)) {
          const store = db.createObjectStore(STORE_PRODUCTOS, { keyPath: "id" });
          store.createIndex("by-categoria", "categoria");
        }
        if (oldVersion < 2 || !db.objectStoreNames.contains(STORE_RED)) {
          db.createObjectStore(STORE_RED, { keyPath: "id" });
        }
      },
    });
  }
  return dbPromise;
};

/**
 * Guarda o actualiza una lista de productos en IndexedDB.
 */
export const guardarProductosOffline = async (productos: Producto[]): Promise<void> => {
  try {
    const db = await getDb();
    const tx = db.transaction(STORE_PRODUCTOS, "readwrite");
    for (const producto of productos) {
      await tx.store.put(producto);
    }
    await tx.done;
  } catch (error) {
    console.warn("[offlineDb] Error guardando productos en IndexedDB:", error);
  }
};

/**
 * Recupera todos los productos almacenados en IndexedDB.
 */
export const obtenerProductosOffline = async (): Promise<Producto[]> => {
  try {
    const db = await getDb();
    return await db.getAll(STORE_PRODUCTOS);
  } catch (error) {
    console.warn("[offlineDb] Error leyendo productos desde IndexedDB:", error);
    return [];
  }
};

/**
 * Recupera un producto individual por id desde IndexedDB.
 */
export const obtenerProductoPorIdOffline = async (id: number): Promise<Producto | undefined> => {
  try {
    const db = await getDb();
    return await db.get(STORE_PRODUCTOS, id);
  } catch (error) {
    console.warn(`[offlineDb] Error leyendo producto ${id} desde IndexedDB:`, error);
    return undefined;
  }
};

/**
 * Guarda la estructura de la red multinivel en IndexedDB.
 */
export const guardarRedOffline = async (red: Referido): Promise<void> => {
  try {
    const db = await getDb();
    await db.put(STORE_RED, { id: "mi_red", data: red });
  } catch (error) {
    console.warn("[offlineDb] Error guardando red en IndexedDB:", error);
  }
};

/**
 * Recupera la estructura de la red multinivel desde IndexedDB.
 */
export const obtenerRedOffline = async (): Promise<Referido | undefined> => {
  try {
    const db = await getDb();
    const registro = await db.get(STORE_RED, "mi_red");
    return registro?.data;
  } catch (error) {
    console.warn("[offlineDb] Error leyendo red desde IndexedDB:", error);
    return undefined;
  }
};
