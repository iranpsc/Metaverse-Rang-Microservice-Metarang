# Build Feature API (v2)

## Overview
- Manage the lifecycle of a building tied to a feature (parcel) within MetaRGB.
- Covers requesting available build packages, starting construction, retrieving construction state, updating active builds, and removing constructed models.
- All routes live under `Route::prefix('features')` in `routes/api_v2.php` and require the v2 stack (`/api/v2` prefix when served by Laravel).

## Authentication & Permissions
- **Middleware:** all routes run behind `auth:sanctum` and `verified`. Anonymous access is not possible.
- **Route Model Binding:** both `feature` and `buildingModel` parameters use Laravel route model binding. The `buildingModel` parameter resolves by `model_id` column.
- **Policy Hooks:**
  - `FeaturePolicy@build` (invoked through `$this->authorize('build', [$feature, $buildingModel])`) is used for build and update operations.
  - `FeaturePolicy@destroyBuilding` (invoked through `$this->authorize('destroyBuilding', [$feature, $buildingModel])`) is used specifically for building removal.
- **Request-Level Authorization:**
  - `StartBuildingFeatureRequest::authorize()` ensures the authenticated user owns the feature AND no building is currently attached (`buildingModels->count() == 0`).
  - `UpdateBuildingFeatureRequest::authorize()` ensures only the feature owner can modify existing builds.
  - `getBuildPackage()` performs direct ownership validation via `throw_unless($feature->owner->id === auth()->id(), AuthorizationException::class)`.

## Endpoint Summary

| Method | Route | Controller Method | Purpose |
| --- | --- | --- | --- |
| GET | `/api/v2/features/{feature}/build/package` | `getBuildPackage` | Fetch pre-calculated building package and feature coordinates. |
| POST | `/api/v2/features/{feature}/build/{buildingModel:model_id}` | `buildFeature` | Start construction of a building model on the feature. |
| GET | `/api/v2/features/{feature}/build/buildings` | `getBuildings` | List building model(s) attached to the feature. |
| PUT | `/api/v2/features/{feature}/build/buildings/{buildingModel:model_id}` | `updateBuilding` | Update construction details for an attached building. |
| DELETE | `/api/v2/features/{feature}/build/buildings/{buildingModel:model_id}` | `destroyBuilding` | Detach a building from the feature and reactivate hourly profits. |

## Endpoint Details

### GET `/api/v2/features/{feature}/build/package`
- **Purpose:** Request a build package from the external 3D Meta service for the selected feature.
- **Query Parameters:** 
  - `page` (optional, defaults to `1`)
- **Behavior:**
  - Loads feature with `properties:id,feature_id,area,density,karbari`, `owner:id`, and `coordinates` relationships.
  - Throws `AuthorizationException` (403) if the authenticated user is not the feature owner.
  - Builds query string containing:
    - `feature_id`: The feature's ID
    - `area`: From feature properties
    - `density`: From feature properties
    - `karbari`: From feature properties (land use type)
    - `page`: Pagination parameter
  - Calls the external API at `config('app.three_d_meta_url') . '/api/v1/build-package'` via HTTP GET.
  - Enriches the remote response with:
    - `required_satisfaction` per model, calculated as: `area × feature->getKarbariCoefficient() × density × 0.1 ÷ 100`, formatted to 4 decimal places.
    - `feature.coordinates`: feature polygon points extracted via `$coordinate->implodeXY()` on each coordinate model.
  - Persists or updates remote building models locally via `BuildingModel::upsert` in a DB transaction:
    - Upserts on `model_id` (unique key)
    - Updates: `name`, `sku`, `images` (JSON), `attributes` (JSON), `file` (JSON), `required_satisfaction`
- **Response Shape:** Mirrors the remote API payload structure with added `feature.coordinates` array and augmented `data[].required_satisfaction`. 
- **Error Handling:** On failure to reach 3D Meta service, returns JSON response with `message` and `error` fields (note: error response code may be inconsistent due to exception handling).

### POST `/api/v2/features/{feature}/build/{buildingModel:model_id}`
- **Purpose:** Start constructing a building model on the given feature.
- **Prerequisites:**
  - User must own the feature.
  - Feature must not already have a building (enforced by `StartBuildingFeatureRequest::authorize()` which checks `buildingModels->count() == 0`).
  - `buildingModel` must exist locally in the database.
  - User must have sufficient satisfaction in wallet (validated by request rules).
- **Validation:** See [Validation Rules](#validation-rules).
- **Behavior (in order):**
  1. Authorization via `FeaturePolicy@build`.
  2. Calculates construction duration in hours: `buildingModel.required_satisfaction × 288000 ÷ launched_satisfaction`.
  3. **Immediately decrements** user wallet satisfaction by `launched_satisfaction` amount.
  4. Calculates construction end timestamp:
     - Converts hours to total seconds
     - Breaks down into days, hours, minutes, seconds
     - Adds to current timestamp via `now()->addDays()->addHours()->addMinutes()->addSeconds()`
  5. **If `activity_line` is provided:** Creates or retrieves `IsicCode` record using `firstOrCreate` with trimmed activity line name. Prepares `information` JSON containing: `activity_line`, `name`, `address`, `postal_code`, `website`, `description`.
  6. Attaches the building model via pivot table (`building`) with:
     - `construction_start_date`: Current timestamp
     - `construction_end_date`: Calculated timestamp
     - `launched_satisfaction`: Amount invested
     - `information`: Business metadata (only if `activity_line` was provided)
     - `rotation`: Rotation degrees
     - `position`: X,Y coordinates string
  7. Deactivates all `FeatureHourlyProfit` records for this feature (`is_active = false`).
  8. Calculates `bubble_diameter`:
     - Extracts `width`, `length`, `density` from building model attributes (by slug)
     - Calculates perimeter: `2 × (width + length)`
     - Calculates density coefficient: starts at 1, adds 0.3 for each density level above 1 (e.g., density=3 → coefficient=1.6)
     - Final diameter: `perimeter × coefficient`
  9. Updates the pivot record with calculated `bubble_diameter`.
  10. Eager loads `buildingModels` relationship with selected fields and pivot data.
- **Response:** Returns HTTP 200 with JSON containing:
  ```json
  {
    "data": {
      // Feature object with buildingModels relationship loaded
      // Including: id, model_id, file
      // Pivot data: construction_start_date, construction_end_date, rotation, position, bubble_diameter
    }
  }
  ```

### GET `/api/v2/features/{feature}/build/buildings`
- **Purpose:** Return all building models linked to the feature with full construction details.
- **Response:** JSON array of `BuildingModelResource` items.

**Resource Structure** (`BuildingModelResource`):

```json
{
  "data": [
    {
      "id": 123,
      "model_id": "model_abc_001",
      "name": "Modern Office Building",
      "sku": "SKU-12345",
      "images": ["url1.jpg", "url2.jpg"],
      "attributes": [
        {"slug": "width", "value": 50},
        {"slug": "length", "value": 30},
        {"slug": "density", "value": 3},
        {"slug": "area", "value": 1500}
      ],
      "file": {"gltf": "model.gltf", "size": 15000},
      "required_satisfaction": "12.5000",
      "building": {
        "model_id": "model_abc_001",
        "feature_id": 456,
        "construction_start_date": "1402/10/15 14:30:25",
        "construction_end_date": "1403/02/20 18:45:30",
        "launched_satisfaction": "25.0000",
        "information": {
          "activity_line": "Software Development",
          "name": "Tech Solutions Inc",
          "address": "123 Main St, Tehran",
          "postal_code": "1234567890",
          "website": "https://example.com",
          "description": "Leading software company"
        },
        "rotation": 45.5,
        "position": "100.5, -50.25",
        "bubble_diameter": 256.5
      }
    }
  ]
}
```

**Field Notes:**
- `images`, `attributes`, `file`: JSON-decoded from stored JSON columns
- `required_satisfaction`, `launched_satisfaction`: Formatted to 4 decimal places as strings
- `construction_start_date`, `construction_end_date`: Jalali (Persian) calendar format via `jdate()` helper
- `information`: Will be `null` if no `activity_line` was provided during build
- `rotation`: Numeric degrees
- `position`: String in "X, Y" format (may include decimals and negatives)

### PUT `/api/v2/features/{feature}/build/buildings/{buildingModel:model_id}`
- **Purpose:** Update an existing building attachment (e.g., adjust satisfaction or metadata).
- **Prerequisites:** User owns the feature (enforced by `UpdateBuildingFeatureRequest::authorize()`).
- **Validation:** Same rules as build creation (see below).
- **Behavior (in order):**
  1. Re-authorizes through `FeaturePolicy@build`.
  2. Recalculates construction duration in hours: `buildingModel.required_satisfaction × 288000 ÷ launched_satisfaction`.
  3. Recalculates construction end timestamp (same logic as build).
  4. **If `activity_line` is provided:** Creates or retrieves `IsicCode` record. Prepares `information` JSON with business metadata.
  5. Updates existing pivot record with:
     - `construction_start_date`: Reset to current timestamp
     - `construction_end_date`: Newly calculated timestamp
     - `launched_satisfaction`: Updated amount
     - `information`: Updated business metadata (only if `activity_line` provided)
     - `rotation`: Updated rotation
     - `position`: Updated position
- **Key Differences from buildFeature:**
  - **Does NOT** decrement user wallet (no satisfaction charge on update)
  - **Does NOT** deactivate `FeatureHourlyProfit` records
  - **Does NOT** recalculate or update `bubble_diameter`
  - **Does NOT** detach/reattach building; only updates pivot
- **Response:** Empty JSON with HTTP 200.

### DELETE `/api/v2/features/{feature}/build/buildings/{buildingModel:model_id}`
- **Purpose:** Remove the building association from the feature and refund invested satisfaction.
- **Behavior (in order):**
  1. Authorizes via `FeaturePolicy@destroyBuilding` (distinct policy method from build/update).
  2. Detaches the building model from the feature (removes pivot record).
  3. **Refunds satisfaction:** Increments user wallet by `$buildingModel->building->launched_satisfaction` (the amount originally invested).
  4. Reactivates all `FeatureHourlyProfit` records for the feature by setting `is_active = true`.
- **Note:** The refund mechanism accesses pivot data via `$buildingModel->building->launched_satisfaction`, where `building` is a relationship on the BuildingModel that retrieves the pivot record.
- **Response:** Empty JSON with HTTP 200.

## Validation Rules

### Shared Field Constraints (Both Build and Update)

| Field | Type | Rules | Notes |
| --- | --- | --- | --- |
| `activity_line` | string | nullable, max 255 | **Conditional:** Triggers `IsicCode::firstOrCreate`. If provided, all business information fields are saved; if omitted, no `information` is stored. |
| `name` | string | nullable, max 255 | Optional business name. Only saved if `activity_line` is provided. |
| `address` | string | nullable, max 255 | Mailing or site address. Only saved if `activity_line` is provided. |
| `postal_code` | string | nullable, `ir_postal_code` | Iran postal code validation (10-digit format). Only saved if `activity_line` is provided. |
| `website` | string | nullable, `active_url`, max 255 | Must resolve via DNS check. Only saved if `activity_line` is provided. |
| `description` | string | nullable, max 5000 | Free-form business description. Only saved if `activity_line` is provided. |
| `launched_satisfaction` | numeric | required, min = `buildingModel.required_satisfaction`, max = authenticated user's current wallet satisfaction | Governs construction duration via formula: `required_satisfaction × 288000 ÷ launched_satisfaction`. Higher values = faster build. |
| `rotation` | numeric | required | Rotation angle for building model placement. |
| `position` | string | required, regex `^(-?\d+(\.\d+)?),\s*(-?\d+(\.\d+)?)$` | Comma-separated X,Y coordinates with optional decimals and negatives (e.g., `-12.5, 34.75`). |

### Request-Specific Authorization

```27:44:app/Http/Requests/StartBuildingFeatureRequest.php
return [
    'launched_satisfaction' => [
        'required',
        'numeric',
        'min:' . $this->route('buildingModel')->required_satisfaction,
        'max:' . $this->user()->wallet->satisfaction,
    ],
    'rotation' => 'required|numeric',
    'position' => [
        'required',
        'regex:/^(-?\d+(\.\d+)?),\s*(-?\d+(\.\d+)?)$/'
    ],
];
```

- `authorize()` additionally ensures only the feature owner can initiate a build and prevents duplicate active builds.

```27:44:app/Http/Requests/UpdateBuildingFeatureRequest.php
return [
    'launched_satisfaction' => [
        'required',
        'numeric',
        'min:' . $this->route('buildingModel')->required_satisfaction,
        'max:' . $this->user()->wallet->satisfaction,
    ],
    'rotation' => 'required|numeric',
    'position' => [
        'required',
        'regex:/^(-?\d+(\.\d+)?),\s*(-?\d+(\.\d+)?)$/'
    ],
];
```

- `authorize()` ensures only the owner may update existing builds.

## Data & Side Effects

### Database Operations
- **`BuildingModel` Caching:** `getBuildPackage()` upserts remote models locally via `BuildingModel::upsert($models, ['model_id'], ['name', 'sku', 'images', 'attributes', 'file', 'required_satisfaction'])` within a DB transaction. This reduces subsequent external API calls.
- **Pivot Table (`building`):** Stores comprehensive metadata:
  - Timestamps: `construction_start_date`, `construction_end_date`
  - Economics: `launched_satisfaction` (amount invested)
  - Business info: `information` JSON (conditional on `activity_line` presence)
  - Spatial: `rotation`, `position`, `bubble_diameter`

### Wallet Transactions
- **Build:** Decrements user wallet satisfaction by `launched_satisfaction` **before** attaching building.
- **Update:** No wallet transaction occurs (free operation).
- **Destroy:** Increments user wallet satisfaction by the original `launched_satisfaction` amount (full refund).

### Profit System Integration
- **On Build:** Sets `FeatureHourlyProfit.is_active = false` for all records matching the feature (disables passive income).
- **On Destroy:** Sets `FeatureHourlyProfit.is_active = true` for all records (re-enables passive income).
- **On Update:** No change to profit status.

### Calculated Fields

#### Required Satisfaction (per model)
Formula: `area × feature->getKarbariCoefficient() × density × 0.1 ÷ 100`
- Formatted to 4 decimal places
- Calculated during `getBuildPackage()` and stored in `BuildingModel.required_satisfaction`

#### Construction Duration
Formula: `required_satisfaction × 288000 ÷ launched_satisfaction` (result in hours)
- **Constant:** 288000 is the base time multiplier (equivalent to ~32.88 years at 1:1 ratio)
- **Inverse Relationship:** Higher `launched_satisfaction` = shorter construction time
- **Minimum:** Must invest at least `required_satisfaction` (enforced by validation)
- **Example Calculations:**
  - Required: 10, Launched: 10 → `10 × 288000 ÷ 10 = 288000 hours` (~32.88 years)
  - Required: 10, Launched: 20 → `10 × 288000 ÷ 20 = 144000 hours` (~16.44 years)
  - Required: 10, Launched: 100 → `10 × 288000 ÷ 100 = 28800 hours` (~3.29 years)
- **Conversion:** Result is broken down via `calculateEndTime()`:
  1. Convert hours → seconds (× 3600)
  2. Extract days (÷ 86400)
  3. Extract hours from remainder (÷ 3600)
  4. Extract minutes from remainder (÷ 60)
  5. Round remaining seconds
  6. Apply to current timestamp via Carbon date manipulation

#### Bubble Diameter
Formula: `perimeter × density_coefficient`
- Where: `perimeter = 2 × (width + length)`
- Where: `density_coefficient = 1 + (0.3 × (density - 1))`
- Example: density=1 → coefficient=1; density=3 → coefficient=1.6
- Only calculated during `buildFeature()`, not on update
- Requires `width`, `length`, `density` attribute slugs in `BuildingModel.attributes`

### ISIC Code Management
- Creates or retrieves `IsicCode` record when `activity_line` is provided
- Uses `firstOrCreate(['name' => trim($activity_line)])` to prevent duplicates
- ISIC codes are stored trimmed of whitespace

## Private Helper Methods

### `calculateRequiredSatisfaction(Feature $feature, array $data): array`
- **Purpose:** Augment each building model in the response with calculated satisfaction requirement.
- **Algorithm:**
  1. Iterates through `$data['data']` array by reference
  2. Extracts `area` from attributes (slug: 'area')
  3. Extracts `density` from attributes (slug: 'density')
  4. Calculates: `area × $feature->getKarbariCoefficient() × density × 0.1 ÷ 100`
  5. Formats result to 4 decimal places via `number_format()`
  6. Adds as `required_satisfaction` key to each item
- **Returns:** Modified data array with augmented items

### `updateOrCreateModels(array $data): void`
- **Purpose:** Persist building models from external API to local database.
- **Algorithm:**
  1. Maps each data item to model array with fields: `model_id`, `name`, `sku`, `images` (JSON), `attributes` (JSON), `file` (JSON), `required_satisfaction`
  2. Wraps upsert in `DB::transaction()`
  3. Calls `BuildingModel::upsert($models, ['model_id'], [...updateFields])`
  4. On conflict by `model_id`, updates all specified fields
- **Side Effects:** Database write within transaction

### `sendRequest(string $url, $query = null): mixed`
- **Purpose:** Make HTTP GET request to external service with error handling.
- **Algorithm:**
  1. Attempts `Http::get($url, $query)`
  2. On exception, returns `response()->json(['message' => ..., 'error' => ...], $e->getCode())`
  3. On success, returns `$response->json()` (array)
- **⚠️ Issue:** Error path returns Laravel Response object, success path returns array—inconsistent return type

### `mergeCoordinates(Feature $feature, array $response): array`
- **Purpose:** Inject feature coordinates into API response.
- **Algorithm:**
  1. Maps feature coordinates collection via `$coordinate->implodeXY()`
  2. Assigns to `$response['feature']['coordinates']`
- **Returns:** Modified response array

### `getConstructionEndDate($constructionLengthHours): Carbon`
- **Purpose:** Calculate future timestamp based on construction duration.
- **Algorithm:**
  1. Calls `calculateEndTime($constructionLengthHours)` to get time components
  2. Extracts `days`, `hours`, `minutes`, `seconds`
  3. Returns `now()->addDays($days)->addHours($hours)->addMinutes($minutes)->addSeconds($seconds)`
- **Returns:** Carbon timestamp instance

### `calculateEndTime($hours): array`
- **Purpose:** Break down hours into time component array.
- **Algorithm:**
  1. Converts hours to seconds: `$seconds = $hours × 3600`
  2. Calculates days: `floor($seconds ÷ 86400)`, subtracts from total
  3. Calculates hours: `floor($seconds ÷ 3600)`, subtracts from remaining
  4. Calculates minutes: `floor($seconds ÷ 60)`, subtracts from remaining
  5. Rounds remaining seconds
- **Returns:** `['days' => int, 'hours' => int, 'minutes' => int, 'seconds' => int]`

### `calculateBubbleDiameter(BuildingModel $buildingModel): float`
- **Purpose:** Compute bubble collision diameter for building placement.
- **Algorithm:**
  1. Collects attributes from building model
  2. Extracts `width` (slug: 'width')
  3. Extracts `length` (slug: 'length')
  4. Calculates perimeter: `2 × (width + length)`
  5. Extracts `density` (slug: 'density')
  6. Calculates coefficient: `1 + (0.3 × (density - 1))` via loop
  7. Returns: `perimeter × coefficient`
- **Formula Details:**
  - Density 1: coefficient = 1.0
  - Density 2: coefficient = 1.3
  - Density 3: coefficient = 1.6
  - Each additional density level adds 0.3 to coefficient
- **⚠️ Dependency:** Requires `width`, `length`, `density` slugs in attributes; fails silently with null if missing

## External Dependencies
- **3D Meta Service:** `config('app.three_d_meta_url')/api/v1/build-package` supplies the available building models and their metadata.
- **Jalali Date Formatting:** Global `jdate()` helper converts Carbon timestamps to Persian calendar format in `BuildingModelResource`.
- **Postal Code Validation:** Custom `ir_postal_code` validation rule enforces Iranian postal code format (10 digits).
- **Feature Model Method:** `$feature->getKarbariCoefficient()` returns a multiplier based on land use type (karbari).

## Implementation Details & Edge Cases

### Error Handling
- **External API Failure:** The `sendRequest()` method catches exceptions when calling 3D Meta API. However, it returns a Laravel response object (`response()->json()`) instead of an array, which may cause issues in `getBuildPackage()` expecting array structure.
- **Missing Attributes:** Both `calculateRequiredSatisfaction()` and `calculateBubbleDiameter()` use `collect()->firstWhere('slug', ...)` on attributes. Missing required slugs will return null and cause arithmetic errors.
- **Authorization Exceptions:** Ownership checks throw `AuthorizationException` with HTTP 403 status.

### Relationships & Data Access
- **BuildingModel Relationships:**
  - Has `attributes` accessor (casts to collection/array)
  - Has `building` relationship to access pivot data in destroy operation
  - Accessed via `$buildingModel->building->launched_satisfaction` in `destroyBuilding()`
- **Feature Relationships:**
  - `owner` (User)
  - `properties` (with select: id, feature_id, area, density, karbari)
  - `coordinates` (collection with `implodeXY()` method)
  - `buildingModels` (many-to-many with pivot data)
- **Route Model Binding:** `buildingModel` resolves by `model_id` column (not default `id`), configured in route definition.

### Business Logic Constraints
- **Single Building Rule:** Features can only have one building at a time (enforced by `StartBuildingFeatureRequest::authorize()` checking `buildingModels->count() == 0`).
- **Update vs Build:**
  - Update allows modifying existing construction without financial penalty
  - Build requires payment and locks out other builds
  - Only Build calculates bubble diameter; Update preserves existing diameter
- **Information Conditionality:** All business information fields (`name`, `address`, etc.) are **only persisted if `activity_line` is provided**. If `activity_line` is null/empty, the `$information` variable is never set, and `information` pivot column receives `null`.

### Performance Considerations
- **DB Transactions:** `updateOrCreateModels()` wraps upsert in transaction for atomicity.
- **Eager Loading:** 
  - `getBuildPackage()` loads minimal properties with field selection
  - `buildFeature()` eager loads building models with specific fields after attachment
  - `getBuildings()` loads all pivot fields without optimization
- **No Pagination:** `getBuildings()` returns all attached buildings without pagination (currently safe due to single-building constraint).

## Quick Reference: Operation Comparison

| Aspect | Build (POST) | Update (PUT) | Destroy (DELETE) |
|--------|-------------|--------------|------------------|
| **Policy Method** | `build` | `build` | `destroyBuilding` |
| **Wallet Action** | Decrement by `launched_satisfaction` | None | Increment by `launched_satisfaction` |
| **Hourly Profit** | Deactivate all for feature | No change | Reactivate all for feature |
| **Bubble Diameter** | Calculate and store | No change (preserve existing) | Removed with pivot |
| **Construction Time** | Calculate from satisfaction | Recalculate and update | N/A |
| **Information Saved** | If `activity_line` provided | If `activity_line` provided | Deleted with pivot |
| **Prerequisite** | No existing building | Building must exist | Building must exist |
| **Response** | Feature with buildings | Empty JSON | Empty JSON |
| **DB Operations** | Insert pivot + wallet update + profit update | Update pivot only | Delete pivot + wallet update + profit update |

## Common Workflow Example

1. **Discovery Phase:**
   ```
   GET /api/v2/features/123/build/package?page=1
   → Returns available building models with required_satisfaction
   ```

2. **Construction Phase:**
   ```
   POST /api/v2/features/123/build/model_abc_001
   Body: {
     "launched_satisfaction": 25.0,
     "rotation": 45,
     "position": "100.5, -50.25",
     "activity_line": "Software Development",
     "name": "Tech Solutions Inc"
   }
   → Deducts 25.0 satisfaction from wallet
   → Returns feature with attached building
   → Disables hourly profit
   ```

3. **Query Phase:**
   ```
   GET /api/v2/features/123/build/buildings
   → Returns current building(s) with construction progress
   ```

4. **Modification Phase (Optional):**
   ```
   PUT /api/v2/features/123/build/buildings/model_abc_001
   Body: {
     "launched_satisfaction": 50.0,
     "rotation": 90,
     "position": "120, -60"
   }
   → No wallet charge
   → Updates construction timeline
   → Bubble diameter unchanged
   ```

5. **Demolition Phase:**
   ```
   DELETE /api/v2/features/123/build/buildings/model_abc_001
   → Refunds 25.0 satisfaction to wallet (original amount)
   → Re-enables hourly profit
   → Removes building attachment
   ```

## Versioning & Future Considerations
- Routes are scoped under API v2; breaking changes should either extend this controller or create v3 counterparts.
- `FeaturePolicy@build` currently returns `true`; if stricter logic is required, update the policy while ensuring controller checks remain consistent.
- Consider caching or paginating `getBuildings` responses if the pivot expands to many records in future iterations.
- **Breaking Change Warning:** The single-building constraint (`buildingModels->count() == 0`) would need refactoring if multi-building support is added.
- **Error Handling Improvement Needed:** `sendRequest()` should return consistent array structure instead of response object on error.
- **Potential Enhancement:** Consider webhooks or events for construction completion notifications.
- **Scaling Consideration:** External 3D Meta API calls in `getBuildPackage()` are synchronous; consider caching or async job processing for high traffic.


