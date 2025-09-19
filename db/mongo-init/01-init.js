// Runs only on first initialization of the MongoDB data directory.
// Root user is created by the official entrypoint using MONGO_INITDB_ROOT_USERNAME/PASSWORD.
// Here we just ensure application databases exist and create basic collections/indexes if desired.

(function() {
  try {
    // Ensure databases exist by touching a collection
    const usersDb = db.getSiblingDB('seentics');
    usersDb.createCollection('init_marker');

    const workflowsDb = db.getSiblingDB('seentics_workflows');
    workflowsDb.createCollection('init_marker');

    // Example indexes (commented out; uncomment if you know the schema)
    // usersDb.users.createIndex({ email: 1 }, { unique: true });

    print('Mongo init script executed: databases ensured.');
  } catch (e) {
    print('Mongo init script error: ' + e.message);
  }
})();
