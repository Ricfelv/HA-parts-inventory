# Home Assistant Parts Inventory

## 1. Purpose

A small personal web application for recording and browsing an electronics parts inventory.

The application is intended to run as a Home Assistant App/Add-on.

The primary goal is simple:

> Quickly find out what electronic parts are available, how many are currently in stock, and identify the exact part when purchasing replacements.

This is a personal inventory/reference tool, not a warehouse or purchasing management system.

## 2. Target Environment

The application is intended to run locally on Home Assistant.

Deployment target:

- Home Assistant OS
- Home Assistant App/Add-on
- Containerized application
- Persistent application data stored in the application's `/data` directory

The application should be usable immediately after installation without requiring manual database setup.

On first startup, if the SQLite database does not exist, the application must create and initialize it automatically.

## 3. Technology

### Required

- Go
- SQLite
- Go `html/template` for server-side HTML rendering

### Frontend

The frontend should remain deliberately simple.

- Plain HTML
- Minimal CSS
- Minimal JavaScript only when it provides a clear usability benefit
- No frontend framework
- No CSS framework
- No Node.js/npm build system unless there is a compelling technical reason

The application should preferably work using normal HTTP requests and server-rendered HTML.

### General principle

Keep the implementation small and understandable.

Do not introduce frameworks, services, libraries, or infrastructure unless they provide a clear and necessary benefit for this application.

## 4. Part Data Model

Each inventory item represents one type of part.

Fields:

| Field | Required | Description |
|---|---|---|
| `id` | System | Unique identifier |
| `type` | Yes | Part category |
| `value` | Yes | Main value/name of the part |
| `package` | No | Package or physical format |
| `desc` | No | Free-form description |
| `mpn` | No | Manufacturer Part Number |
| `quantity` | Yes | Current quantity in stock |
| `updated_at` | System | Last modification timestamp |

### Examples

A resistor:

- Type: `Resistor`
- Value: `10kΩ`
- Package: `0603`
- Desc: `1%`
- MPN: `RC0603FR-0710KL`
- Quantity: `87`

A capacitor:

- Type: `Capacitor`
- Value: `10µF`
- Package: `0603`
- Desc: `16V X7R`
- MPN: optional
- Quantity: `32`

An IC:

- Type: `IC`
- Value: `TPS62745`
- Package: optional
- Desc: `1.8V buck converter`
- MPN: optional
- Quantity: `2`

## 5. Part Types

`type` should be selected from a predefined English-language dropdown.

Initial list:

- Resistor
- Capacitor
- Inductor
- Diode
- LED
- Transistor
- MOSFET
- IC
- Connector
- Module
- Switch
- Fuse
- Other

Users should not need to create or manage custom types through the UI in the initial version.

The list can be expanded in future versions if necessary.

## 6. Main Page

The main page is the primary interface of the application.

It should display the inventory as a table.

There is no need for pagination in the initial version. The expected inventory size is small, so displaying all matching records is acceptable.

### Header area

The top of the page should contain:

- Search/filter field
- "In stock only" filter
- "Add Part" button

Example:

```text
[ Search...                         ] [x] In stock only    [ + Add Part ]
```

### Default filter

`In stock only` should be enabled by default.

When enabled:

```text
quantity > 0
```

Items with quantity `0` remain in the database but are hidden from the normal inventory view.

The user can disable the filter to see all records, including out-of-stock items.

## 7. Search / Filter

Search and filter are considered the same feature in the initial version.

There should be one simple text search field.

Search should be case-insensitive and should perform a basic substring search.

The search should cover the useful textual fields of a part, including:

- `type`
- `value`
- `package`
- `desc`
- `mpn`

Examples:

Searching for:

```text
10k
```

should find parts whose value or other searchable field contains `10k`.

Searching for:

```text
0603
```

should find parts using the 0603 package.

Searching for:

```text
RC0603
```

should find a matching MPN.

No advanced query language is required.

## 8. Sorting

The inventory table should support sorting by clicking column headers.

Clicking a sortable column should toggle between:

- Ascending
- Descending

Clicking the same column repeatedly should alternate between the two directions.

Multi-column sorting is not required.

The initial/default sort should be:

```text
id ascending
```

The `id` does not need to be displayed as a visible table column.

Sortable columns should include the useful user-facing fields, such as:

- Type
- Value
- Package
- Description
- MPN
- Quantity

## 9. Inventory Table

The table should display the useful inventory information.

Suggested columns:

```text
Type
Value
Package
Description
MPN
Quantity
Edit
```

Each row should have an `Edit` action.

The table should use subtle alternating row backgrounds (zebra striping) to improve readability.

No elaborate visual design is required.

The UI should prioritize:

1. Readability
2. Fast searching
3. Easy quantity checking
4. Simple interaction

## 10. Add Part

An `Add Part` button should be displayed above the inventory table.

The user should be able to create a new part with:

- Type
- Value
- Package
- Description
- MPN
- Quantity

Required fields:

- Type
- Value
- Quantity

Optional fields:

- Package
- Description
- MPN

After successfully creating a part, the user should be returned to the inventory view.

## 11. Edit Part

Each inventory row should have an `Edit` button.

The edit page should allow the user to modify all user-editable part fields:

- Type
- Value
- Package
- Description
- MPN
- Quantity

There should be no Delete button in the UI.

A part with quantity `0` should remain in the database.

If the user wants to permanently remove a record, database-level/manual removal is acceptable for this personal application.

## 12. Quantity Adjustment

The edit page should provide two ways to change quantity.

### Direct quantity editing

Display the current quantity in an editable text box.

Example:

```text
Current quantity
[ 87 ]
```

The user can directly replace the value when they perform a physical stock count.

### Quantity adjustment

Provide a separate adjustment amount and explicit Add/Remove actions.

Example:

```text
Current quantity
[ 87 ]

Adjustment
[ 3 ]

[ - Remove 3 ]    [ + Add 3 ]
```

This should make the direction of the operation completely explicit.

For example:

```text
87
Remove 3
84
```

or:

```text
87
Add 3
90
```

The quantity must never become negative.

The adjustment control is intended for quick stock changes, while direct editing is intended for correcting the actual stock count.

## 13. No Delete UI

The application should not provide a Delete button.

Inventory records with zero quantity are intentionally retained.

This allows the application to function as both:

- Current inventory
- A small personal parts catalog

## 14. Database

SQLite should be used as the only database.

The database should be stored in the persistent application data directory:

```text
/data/parts.db
```

The database file must not be stored inside the application source tree.

The database file must not be committed to Git.

Example `.gitignore` entries should prevent accidental commits of:

```text
*.db
*.db-shm
*.db-wal
```

### Database initialization

On application startup:

1. Open the SQLite database.
2. If the database file does not exist, create it.
3. Initialize the required schema.
4. Run any required database migrations.
5. Start serving the web application.

The application must be able to start from an empty `/data` directory.

## 15. Database Migrations

The project should use a simple migration mechanism so that future schema changes can be applied without requiring the user to delete their database.

For example:

```text
Initial schema
      ↓
Future schema changes
      ↓
Migration
```

The exact migration implementation is left to the developer, but it should remain lightweight and appropriate for a small Go + SQLite application.

## 16. Persistence

Application data must survive application/container restarts and upgrades.

The `/data` directory must therefore be configured as persistent storage when packaged as a Home Assistant App/Add-on.

The application must never depend on the container filesystem for persistent inventory data.

## 17. Home Assistant App/Add-on

The application should be packaged so that it can be installed through Home Assistant as an App/Add-on.

The intended user workflow is:

```text
Add repository
      ↓
Install Parts Inventory
      ↓
Start
      ↓
Open Web UI
```

The application should not require the user to manually SSH into Home Assistant, install Go, install SQLite, or manually configure the database.

The exact Home Assistant packaging implementation can be decided during development.

## 18. Security / Authentication

No application-level authentication is required in the initial version.

The application is intended for personal use inside the user's Home Assistant environment.

Do not add a login system unless it becomes necessary later.

## 19. Explicit Non-Goals

The initial version must NOT attempt to implement:

- Warehouse/location management
- Multiple storage locations
- Supplier management
- Purchasing management
- Purchase orders
- Price tracking
- BOM management
- Project management
- Multi-user accounts
- User permissions
- Cloud synchronization
- External inventory services
- Automatic part identification
- Automatic distributor lookup
- Advanced analytics
- Stock forecasting
- Minimum-stock alerts
- Inventory history/audit system
- Delete UI
- Complex dashboard functionality

The application should remain a small personal parts inventory and reference tool.

## 20. Development Principles

The project should prioritize simplicity over abstraction.

Prefer:

```text
Simple Go code
Simple SQL
Simple HTML
Simple HTTP
```

over introducing additional frameworks or infrastructure.

Avoid premature optimization.

Avoid building functionality that is not required by this specification.

The application should be easy for a technically capable user who is not a professional software developer to understand, build, deploy, and maintain.

When there are multiple reasonable implementation choices, prefer the smaller and more maintainable solution.

## 21. Initial Success Criteria

The first usable version is successful when the user can:

1. Install the application on Home Assistant.
2. Start it without manually creating a database.
3. Open the web interface.
4. Add an electronic part.
5. See it in the inventory table.
6. Search for it.
7. Sort the inventory.
8. Edit its information.
9. Increase or decrease its quantity.
10. Set its quantity to zero.
11. Hide zero-stock items using the default filter.
12. Restart/reinstall the application without losing the inventory data, provided `/data` is preserved.
