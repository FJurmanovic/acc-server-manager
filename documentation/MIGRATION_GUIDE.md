# ACC Server Manager - Migration and Role System Enhancement Guide

## Overview

This guide documents the comprehensive updates made to the ACC Server Manager, including UUID migrations, enhanced role system, and frontend improvements.

## Changes Made

### 1. Database Migration to UUID Primary Keys

**Problem**: The original database used integer primary keys which could cause issues with scaling and distributed systems.

**Solution**: Migrated all primary keys to UUIDs while preserving data integrity and foreign key relationships.

#### Tables Migrated:
- `servers` - Server configurations
- `configs` - Configuration history
- `state_histories` - Server state tracking
- `steam_credentials` - Steam login credentials
- `system_configs` - System configuration settings

#### Migration Scripts:
- `scripts/migrations/002_migrate_servers_to_uuid.sql` - SQL migration script
- `local/migrations/002_migrate_to_uuid.go` - Go migration handler
- `scripts/run_migrations.go` - Standalone migration runner
- `scripts/test_migrations.go` - Test suite for migrations

### 2. Enhanced Role System

**Problem**: The original system only had "Super Admin" role with limited role management.

**Solution**: Implemented a comprehensive role system with three predefined roles and permission-based access control.

#### New Roles:

1. **Super Admin**
   - All permissions
   - Cannot be deleted (protected)
   - System administrator level access

2. **Admin**
   - All permissions (same as Super Admin)
   - Can be deleted
   - Regular administrative access

3. **Manager**
   - Limited permissions
   - Cannot create/delete: servers, users, roles, memberships
   - Can view and manage existing resources

#### Permission Structure:
```
Server Permissions:
- server.view, server.create, server.update, server.delete
- server.start, server.stop

Configuration Permissions:
- config.view, config.update

User Management Permissions:
- user.view, user.create, user.update, user.delete

Role Management Permissions:
- role.view, role.create, role.update, role.delete

Membership Permissions:
- membership.view, membership.create, membership.edit
```

### 3. Frontend Improvements

**Problem**: Role assignment used a text input field, making it error-prone and inconsistent.

**Solution**: Replaced text input with a dropdown populated from the backend API.

#### Changes:
- Added `/membership/roles` API endpoint
- Updated membership service to fetch available roles
- Modified create user modal to use dropdown selection
- Improved user experience with consistent role selection

### 4. Super Admin Protection

**Problem**: No protection against accidentally deleting the Super Admin user.

**Solution**: Added validation to prevent deletion of users with "Super Admin" role.

#### Implementation:
- Service-level validation in `DeleteUser` method
- Returns error: "cannot delete Super Admin user"
- Maintains system integrity by ensuring at least one Super Admin exists

## Installation and Usage

### Running Migrations

#### Option 1: Automatic Migration (Recommended)
Migrations run automatically when the application starts:

```bash
cd acc-server-manager
go run cmd/api/main.go
```

#### Option 2: Manual Migration
Run migrations manually using the migration script:

```bash
cd acc-server-manager
go run scripts/run_migrations.go [database_path]
```

#### Option 3: Test Migrations
Test the migration process with the test suite:

```bash
cd acc-server-manager
go run scripts/test_migrations.go
```

### Backend API Changes

#### New Endpoints:

1. **Get All Roles**
   ```
   GET /membership/roles
   Authorization: Bearer <token>
   Required Permission: role.view
   
   Response:
   [
     {
       "id": "uuid",
       "name": "Super Admin"
     },
     {
       "id": "uuid", 
       "name": "Admin"
     },
     {
       "id": "uuid",
       "name": "Manager"
     }
   ]
   ```

2. **Enhanced User Creation**
   ```
   POST /membership
   Authorization: Bearer <token>
   Required Permission: membership.create
   
   Body:
   {
     "username": "string",
     "password": "string", 
     "role": "Super Admin|Admin|Manager"
   }
   ```

### Frontend Changes

#### Role Selection Dropdown
The user creation form now includes a dropdown for role selection:

```html
<select name="role" required>
  <option value="">Select a role...</option>
  <option value="Super Admin">Super Admin</option>
  <option value="Admin">Admin</option>
  <option value="Manager">Manager</option>
</select>
```

#### Updated API Service
The membership service includes the new `getRoles()` method:

```typescript
async getRoles(event: RequestEvent): Promise<Role[]> {
  return await fetchAPIEvent(event, '/membership/roles');
}
```

## Migration Safety

### Backup Strategy
1. **Automatic Backup**: The migration script creates temporary backup tables
2. **Transaction Safety**: All migrations run within database transactions
3. **Rollback Support**: Failed migrations are automatically rolled back

### Data Integrity
- Foreign key relationships are maintained during migration
- Existing data is preserved with new UUID identifiers
- Lookup tables (tracks, car models, etc.) remain unchanged

### Validation
- UUID format validation for all migrated IDs
- Referential integrity checks after migration
- Comprehensive test suite verifies migration success

## Troubleshooting

### Common Issues

1. **Migration Already Applied**
   - Error: "UUID migration already applied, skipping"
   - Solution: This is normal, migrations are idempotent

2. **Database Lock Error**
   - Error: "database is locked"
   - Solution: Ensure no other processes are using the database

3. **Permission Denied**
   - Error: "failed to execute UUID migration"
   - Solution: Check file permissions and disk space

4. **Foreign Key Constraint Error**
   - Error: "FOREIGN KEY constraint failed"
   - Solution: Verify database integrity before running migration

### Debugging

Enable debug logging to see detailed migration progress:

```bash
# Set environment variable
export DEBUG=true

# Or modify the Go code
logging.Init(true) // Enable debug logging
```

### Recovery

If migration fails:

1. **Restore from backup**: Use the backup files created during migration
2. **Re-run migration**: The migration is idempotent and can be safely re-run
3. **Manual cleanup**: Remove temporary tables and retry

## Testing

### Automated Tests
Run the comprehensive test suite:

```bash
cd acc-server-manager
go run scripts/test_migrations.go
```

### Manual Testing
1. Create test users with different roles
2. Verify permission restrictions work correctly
3. Test Super Admin deletion prevention
4. Confirm frontend dropdown functionality

### Test Database
The test suite creates a temporary database (`test_migrations.db`) that is automatically cleaned up after testing.

## Performance Considerations

### Database Performance
- UUIDs are stored as TEXT in SQLite for compatibility
- Indexes are created on frequently queried UUID columns
- Foreign key constraints ensure referential integrity

### Memory Usage
- Migration process uses temporary tables to minimize memory footprint
- Batch processing for large datasets
- Transaction-based approach reduces memory leaks

## Security Enhancements

### Role-Based Access Control
- Granular permissions for different operations
- Service-level permission validation
- Middleware-based authentication and authorization

### Super Admin Protection
- Prevents accidental deletion of critical users
- Maintains system accessibility
- Audit trail for all user management operations

## Future Enhancements

### Planned Features
1. **Custom Roles**: Allow creation of custom roles with specific permissions
2. **Role Inheritance**: Implement role hierarchy with permission inheritance
3. **Audit Logging**: Track all role and permission changes
4. **Bulk Operations**: Support for bulk user management operations

### Migration Extensions
1. **Data Archival**: Migrate old data to archive tables
2. **Performance Optimization**: Add database-specific optimizations
3. **Incremental Migrations**: Support for partial migrations

## Support

For issues or questions regarding the migration and role system:

1. Check the logs for detailed error messages
2. Review this guide for common solutions
3. Run the test suite to verify system integrity
4. Consult the API documentation for endpoint details

## Changelog

### Version 2.0.0
- ✅ Migrated all primary keys to UUID
- ✅ Added Super Admin, Admin, and Manager roles
- ✅ Implemented permission-based access control
- ✅ Added Super Admin deletion protection
- ✅ Created role selection dropdown in frontend
- ✅ Added comprehensive test suite
- ✅ Improved database migration system

### Version 1.0.0
- Basic user management with Super Admin role
- Integer primary keys
- Text-based role assignment