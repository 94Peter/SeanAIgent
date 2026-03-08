/**
 * Migration: Remove Deprecated Boolean Fields from Appointment Collection
 * 
 * This script removes 'is_checked_in', 'is_on_leave', and 'is_check_in' fields
 * from all documents in the 'appointment' collection. 
 * 
 * The business logic has been migrated to use the 'status' field (V2).
 * 
 * Usage: 
 * mongosh <connection_string> remove_deprecated_appt_fields.js
 */

const dbName = "seanAIgent"; // Adjust if your database name is different
const collectionName = "appointment";

const db = db.getSiblingDB(dbName);

print(`Starting migration: Removing deprecated fields from ${collectionName}...`);

const result = db.getCollection(collectionName).updateMany(
    {}, 
    {
        $unset: {
            is_checked_in: "",
            is_on_leave: "",
            is_check_in: ""
        }
    }
);

print(`Migration completed!`);
print(`Matched documents: ${result.matchedCount}`);
print(`Modified documents: ${result.modifiedCount}`);
