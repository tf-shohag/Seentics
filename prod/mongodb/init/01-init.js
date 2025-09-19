// Runs only on first initialization (empty /data/db)
// Official entrypoint will create the root user using MONGO_INITDB_ROOT_USERNAME/PASSWORD.
// This script ensures application databases exist.
(function () {
  try {
    const usersDb = db.getSiblingDB('seentics');
    usersDb.createCollection('init_marker');

    const workflowsDb = db.getSiblingDB('seentics_workflows');
    workflowsDb.createCollection('init_marker');

    print('Mongo init: ensured seentics and seentics_workflows DBs.');
  } catch (e) {
    print('Mongo init error: ' + e.message);
  }
})();
