# Features Service gRPC API Reference

This document describes the gRPC API endpoints exposed by the Features Service. These endpoints are accessible via the gRPC Gateway (Kong) which converts them to REST endpoints.

## Service: FeatureService

### ListFeatures

Lists features within a bounding box.

**Request:**
```protobuf
message ListFeaturesRequest {
  repeated string points = 1;  // Bounding box: ["minX,minY", "maxX,minY", "maxX,maxY", "minX,maxY"]
  bool load_buildings = 2;
  bool user_features_location = 3;
}
```

**Response:**
```protobuf
message FeaturesResponse {
  repeated Feature features = 1;
}
```

**REST Equivalent:** `GET /api/features?points=...`

### GetFeature

Retrieves a single feature by ID.

**Request:**
```protobuf
message GetFeatureRequest {
  uint64 feature_id = 1;
}
```

**Response:**
```protobuf
message Feature {
  uint64 id = 1;
  // ... feature details
}
```

**REST Equivalent:** `GET /api/features/{feature}`

## Service: FeatureMarketplaceService

### BuyFeature

Purchases a feature directly (limited, RGB, or user-to-user).

**Request:**
```protobuf
message BuyFeatureRequest {
  uint64 feature_id = 1;
  uint64 buyer_id = 2;
}
```

**Response:**
```protobuf
message Feature {
  // Updated feature after purchase
}
```

**REST Equivalent:** `POST /api/features/buy/{feature}`

### SendBuyRequest

Creates a buy request for a feature.

**Request:**
```protobuf
message SendBuyRequestRequest {
  uint64 buyer_id = 1;
  uint64 feature_id = 2;
  string price_psc = 3;
  string price_irr = 4;
  string note = 5;
}
```

**Response:**
```protobuf
message BuyRequestResponse {
  BuyRequest buy_request = 1;
}
```

**REST Equivalent:** `POST /api/buy-requests/store/{feature}`

### AcceptBuyRequest

Accepts a buy request (seller action).

**Request:**
```protobuf
message AcceptBuyRequestRequest {
  uint64 request_id = 1;
  uint64 seller_id = 2;
}
```

**Response:**
```protobuf
message BuyRequestResponse {
  BuyRequest buy_request = 1;
}
```

**REST Equivalent:** `POST /api/buy-requests/accept/{buyFeatureRequest}`

### RejectBuyRequest

Rejects a buy request (seller action).

**Request:**
```protobuf
message RejectBuyRequestRequest {
  uint64 request_id = 1;
  uint64 seller_id = 2;
}
```

**Response:**
```protobuf
google.protobuf.Empty
```

**REST Equivalent:** `POST /api/buy-requests/reject/{buyFeatureRequest}`

### DeleteBuyRequest

Deletes a buy request (buyer cancels their own offer).

**Request:**
```protobuf
message DeleteBuyRequestRequest {
  uint64 request_id = 1;
  uint64 buyer_id = 2;
}
```

**Response:**
```protobuf
google.protobuf.Empty
```

**REST Equivalent:** `DELETE /api/buy-requests/delete/{buyFeatureRequest}`

### ListBuyRequests

Lists all buy requests for a buyer.

**Request:**
```protobuf
message ListBuyRequestsRequest {
  uint64 buyer_id = 1;
}
```

**Response:**
```protobuf
message BuyRequestsResponse {
  repeated BuyRequest buy_requests = 1;
}
```

**REST Equivalent:** `GET /api/buy-requests`

### ListReceivedBuyRequests

Lists all buy requests received by a seller.

**Request:**
```protobuf
message ListReceivedBuyRequestsRequest {
  uint64 seller_id = 1;
}
```

**Response:**
```protobuf
message BuyRequestsResponse {
  repeated BuyRequest buy_requests = 1;
}
```

**REST Equivalent:** `GET /api/buy-requests/recieved`

### UpdateGracePeriod

Updates the grace period for a buy request.

**Request:**
```protobuf
message UpdateGracePeriodRequest {
  uint64 request_id = 1;
  uint64 seller_id = 2;
  int32 grace_period_days = 3;
}
```

**Response:**
```protobuf
google.protobuf.Empty
```

**REST Equivalent:** `POST /api/buy-requests/add-grace-period/{buyFeatureRequest}`

### CreateSellRequest

Creates a sell request for a feature.

**Request:**
```protobuf
message CreateSellRequestRequest {
  uint64 seller_id = 1;
  uint64 feature_id = 2;
  string price_psc = 3;  // Optional
  string price_irr = 4;  // Optional
  int32 minimum_price_percentage = 5;  // Optional (mutually exclusive with price_psc/price_irr)
}
```

**Response:**
```protobuf
message SellRequestResponse {
  SellRequest sell_request = 1;
}
```

**REST Equivalent:** `POST /api/sell-requests/store/{feature}`

### ListSellRequests

Lists all sell requests for a seller.

**Request:**
```protobuf
message ListSellRequestsRequest {
  uint64 seller_id = 1;
}
```

**Response:**
```protobuf
message SellRequestsResponse {
  repeated SellRequest sell_requests = 1;
}
```

**REST Equivalent:** `GET /api/sell-requests`

### DeleteSellRequest

Deletes a sell request.

**Request:**
```protobuf
message DeleteSellRequestRequest {
  uint64 sell_request_id = 1;
  uint64 seller_id = 2;
}
```

**Response:**
```protobuf
google.protobuf.Empty
```

**REST Equivalent:** `DELETE /api/sell-requests/{sellRequest}`

## Service: FeatureProfitService

### GetHourlyProfits

Retrieves paginated hourly profits for a user.

**Request:**
```protobuf
message GetHourlyProfitsRequest {
  uint64 user_id = 1;
  int32 page = 2;
  int32 page_size = 3;
}
```

**Response:**
```protobuf
message HourlyProfitsResponse {
  repeated HourlyProfit profits = 1;
  string total_maskoni = 2;
  string total_tejari = 3;
  string total_amozeshi = 4;
}
```

**REST Equivalent:** `GET /api/hourly-profits`

### GetSingleProfit

Withdraws a single hourly profit.

**Request:**
```protobuf
message GetSingleProfitRequest {
  uint64 profit_id = 1;
  uint64 user_id = 2;
}
```

**Response:**
```protobuf
message HourlyProfit {
  uint64 id = 1;
  // ... profit details
}
```

**REST Equivalent:** `POST /api/hourly-profits/{featureHourlyProfit}`

### GetProfitsByApplication

Withdraws all profits by karbari (application type).

**Request:**
```protobuf
message GetProfitsByApplicationRequest {
  uint64 user_id = 1;
  string karbari = 2;  // "m", "t", or "a"
}
```

**Response:**
```protobuf
message ProfitsByApplicationResponse {
  double total_amount = 1;
}
```

**REST Equivalent:** `POST /api/hourly-profits`

## Service: MapsService

### ListMaps

Lists all maps with feature rollups.

**Request:**
```protobuf
message ListMapsRequest {
  // Empty
}
```

**Response:**
```protobuf
message ListMapsResponse {
  repeated Map maps = 1;
}
```

**REST Equivalent:** `GET /api/v2/maps`

### GetMap

Retrieves a single map by ID.

**Request:**
```protobuf
message GetMapRequest {
  uint64 map_id = 1;
}
```

**Response:**
```protobuf
message Map {
  uint64 id = 1;
  // ... map details
}
```

**REST Equivalent:** `GET /api/v2/maps/{map}`

### GetMapBorder

Retrieves border coordinates for a map.

**Request:**
```protobuf
message GetMapRequest {
  uint64 map_id = 1;
}
```

**Response:**
```protobuf
message GetMapBorderResponse {
  repeated Coordinate coordinates = 1;
}
```

**REST Equivalent:** `GET /api/v2/maps/{map}/border`

## Service: BuildingService

### GetBuildPackage

Retrieves build package information.

**Request:**
```protobuf
message GetBuildPackageRequest {
  uint64 feature_id = 1;
}
```

**Response:**
```protobuf
message BuildPackage {
  // Build package details
}
```

**REST Equivalent:** `GET /api/build-package/{feature}`

### BuildFeature

Initiates building construction for a feature.

**Request:**
```protobuf
message BuildFeatureRequest {
  uint64 feature_id = 1;
  uint64 user_id = 2;
  // ... building parameters
}
```

**Response:**
```protobuf
message Building {
  // Building details
}
```

**REST Equivalent:** `POST /api/build-feature/{feature}`

## Error Codes

All endpoints may return the following gRPC status codes:

- `OK`: Success
- `InvalidArgument`: Validation error (e.g., invalid price, missing required field)
- `NotFound`: Resource not found (e.g., feature, request)
- `Unauthenticated`: Missing or invalid authentication token
- `PermissionDenied`: User not authorized (e.g., not the owner)
- `FailedPrecondition`: Business rule violation (e.g., underpriced restriction, insufficient balance)
- `Internal`: Unexpected server error

## Authentication

Most endpoints require authentication via `auth:sanctum` middleware. The user ID is extracted from the JWT token and passed in the request context.

Public endpoints (no authentication):
- Health checks
- Some read-only endpoints (if configured)

## Rate Limiting

Rate limiting is handled by Kong API Gateway. Default limits:
- 100 requests per minute per user
- 1000 requests per minute per IP

## Versioning

Current API version: v1

Proto definitions: `shared/proto/features/features.proto`

