# UUID Migration Instructions

## Overview

This guide explains how to migrate your ACC Server Manager database from integer primary keys to UUIDs. This migration is required to update your system with the new role management features and improved database architecture.

## ⚠️ Important: Backup First

**ALWAYS backup your database before running any migration!**

```bash
# Create a backup of your database
copy acc.db acc.db.backup
# or on Linux/Mac
cp acc.db acc.db.backup
```

## Migration Methods

### Option 1: Standalone Migration Tool (Recommended)

Use the dedicated migration tool to safely migrate your database:

```bash
# Navigate to the project directory
cd acc-server-manager

# Build and run the migration tool
go run cmd/migrate/main.go

# Or specify a custom database path
go run cmd/migrate/main.go path/to/your/acc.db
```

**What this does:**
- Checks if migration is needed
- Uses the existing SQL migration script (`scripts/migrations/002_migrate_servers_to_uuid.sql`)
- Safely migrates all tables from integer IDs to UUIDs
- Preserves all existing data and relationships
- Creates migration tracking records

### Option 2: Using the Migration Script

You can also run the standalone migration script:

```bash
cd acc-server-manager
go run scripts/run_migrations.go
```

**Note:** Both migration tools use the same SQL migration file (`scripts/migrations/002_migrate_servers_to_uuid.sql`) to ensure consistency.

## What Gets Migrated

The migration will update these tables to use UUID primary keys:

1. **servers** - Server configurations
2. **configs** - Configuration change history  
3. **state_histories** - Server state tracking
4. **steam_credentials** - Steam login credentials
5. **system_configs** - System configuration settings

## Verification

After migration, verify it worked correctly:

1. **Check Migration Status:**
   ```bash
   go run cmd/migrate/main.go
   # Should show: "Migration not needed - database already uses UUID primary keys"
   ```

2. **Check Database Schema:**
   ```bash
   sqlite3 acc.db ".schema servers"
   # Should show: CREATE TABLE `servers` (`id` text,...)
   ```

3. **Start the Application:**
   ```bash
   go run cmd/api/main.go
   # Should start without UUID-related errors
   ```

## Troubleshooting

### Error: "NOT NULL constraint failed"

If you see this error, it means there's a conflict between the old schema and new models. Run the migration tool first:

```bash
# Stop the application if running
# Run migration
go run cmd/migrate/main.go
# Then restart the application
go run cmd/api/main.go
```

### Error: "Database is locked"

Make sure the ACC Server Manager application is not running during migration:

```bash
# Stop any running instances
# Then run migration
go run cmd/migrate/main.go
```

### Error: "Migration failed"

If migration fails:

1. Restore from backup:
   ```bash
   copy acc.db.backup acc.db
   ```

2. Check database integrity:
   ```bash
   sqlite3 acc.db "PRAGMA integrity_check;"
   ```

3. Try migration again or contact support

## After Migration

Once migration is complete:

1. **New Role System Available:**
   - Super Admin (cannot be deleted)
   - Admin (full permissions)
   - Manager (limited permissions)

2. **Improved Frontend:**
   - Role dropdown in user creation
   - Better user management interface

3. **Enhanced Security:**
   - Super Admin deletion protection
   - Permission-based access control

## Migration Safety Features

- **Transaction-based:** All changes are wrapped in database transactions
- **Rollback support:** Failed migrations are automatically rolled back  
- **Data preservation:** All existing data is maintained
- **Idempotent:** Can be safely run multiple times
- **Backup creation:** Temporary backup tables during migration

## Manual Rollback (If Needed)

If you need to rollback manually:

```bash
# Restore from backup
copy acc.db.backup acc.db

# Or if you have the old integer schema SQL:
sqlite3 acc.db < old_schema.sql
```

## Testing Migration

To test the migration on a copy of your database:

```bash
# Create test copy
copy acc.db test.db

# Run migration on test copy
go run cmd/migrate/main.go test.db

# Verify test database works
# If successful, run on real database
```

## Support

If you encounter issues:

1. Check this troubleshooting guide
2. Verify you have a backup
3. Check the application logs for detailed errors
4. Try running the test migration first

## Migration Checklist

- [ ] Application is stopped
- [ ] Database is backed up
- [ ] Migration tool is run successfully
- [ ] Migration verification completed
- [ ] Application starts without errors
- [ ] User management features work correctly
- [ ] Role dropdown functions properly

## Technical Details

The migration:
- Uses the SQL script: `scripts/migrations/002_migrate_servers_to_uuid.sql`
- Converts integer primary keys to UUID (stored as TEXT)
- Updates all foreign key references
- Preserves data integrity and relationships
- Uses SQLite-compatible UUID generation
- Creates proper indexes for performance
- Maintains GORM model compatibility
- Both Go and standalone tools use the same SQL for consistency

This migration is a one-time process that prepares your database for the enhanced role management system and future scalability improvements.