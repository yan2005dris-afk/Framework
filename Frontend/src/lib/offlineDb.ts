import { openDB, type DBSchema, type IDBPDatabase } from "idb";
import type { Producto } from "@/data/productos";

const DB_NAME = "multicatalogo_offline_db";
const DB_VERSION = 1;
const STORE_PRODUCTOS = "productos";

interface MultiCatalogoDB extends DBSchema {
  productos: {
    key: number;
    value: Producto;
    indexes: { "by-categoria": string };
  };
}

let dbPromise: Promise<IDBPDatabase<MultiCatalogoDB>> | null = null;

const getDb = () => {
  if (!dbPromise) {
    dbPromise = openDB<MultiCatalogoDB>(DB_NAME, DB_VERSION, {
      upgrade(db) {
        if (!db.objectStoreNames.contains(STORE_PRODUCTOS)) {
          const store = db.createObjectStore(STORE_PRODUCTOS, { keyPath: "id" });
          store.createIndex("by-categoria", "categoria");
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
