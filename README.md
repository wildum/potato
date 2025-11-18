# Potato Service API

A comprehensive RESTful API service for managing potato inventory, recipes, and analytics. Built with Go and designed for demonstration purposes with a professional structure.

## Features

- 🥔 **Potato Management**: Full CRUD operations for potato inventory
- 📊 **Analytics**: Real-time inventory analytics and statistics
- 📖 **Recipe Database**: Store and retrieve potato recipes
- 🎯 **Recipe Recommendations**: Smart recipe suggestions based on variety and difficulty
- ✅ **Freshness Tracking**: Calculate potato freshness based on harvest date
- 📦 **Inventory Summary**: Comprehensive inventory reporting by variety
- 🔄 **Background Processing**: Automatic inventory updates and quality degradation
  - New potatoes added every 3 seconds
  - New recipes generated every 8 seconds
  - Potato quality degrades over time (every 20 seconds)

## Quick Start

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
# Clone the repository
git clone https://github.com/williamdumont/potato-demo.git
cd potato-demo

# Download dependencies
go mod download

# Run the service
go run main.go
```

The service will start on `http://localhost:8081`

## API Endpoints

### Health Check

```
GET /api/v1/health
```

Returns the health status of the service.

**Response:**
```json
{
  "status": "healthy",
  "service": "potato-service"
}
```

### Potatoes

#### Get All Potatoes

```
GET /api/v1/potatoes
GET /api/v1/potatoes?variety=Russet
```

Retrieve all potatoes or filter by variety.

**Response:**
```json
[
  {
    "id": "p001",
    "variety": "Russet",
    "origin": "Idaho",
    "weight": 0.45,
    "quality": "Premium",
    "harvest_date": "2024-11-13T10:00:00Z",
    "price": 2.99
  }
]
```

#### Get Potato by ID

```
GET /api/v1/potatoes/{id}
```

Retrieve a specific potato by ID.

#### Create Potato

```
POST /api/v1/potatoes
```

Create a new potato entry.

**Request Body:**
```json
{
  "id": "p009",
  "variety": "Russet",
  "origin": "Idaho",
  "weight": 0.45,
  "quality": "Premium",
  "harvest_date": "2024-11-13T10:00:00Z",
  "price": 2.99
}
```

**Supported Varieties:**
- Russet
- Yukon Gold
- Red Potato
- Fingerling
- Sweet Potato
- Purple Potato

**Quality Levels:**
- Premium
- Standard
- Economy

#### Update Potato

```
PUT /api/v1/potatoes/{id}
```

Update an existing potato.

**Request Body:**
```json
{
  "variety": "Russet",
  "origin": "Idaho",
  "weight": 0.50,
  "quality": "Premium",
  "harvest_date": "2024-11-13T10:00:00Z",
  "price": 3.49
}
```

#### Delete Potato

```
DELETE /api/v1/potatoes/{id}
```

Delete a potato from inventory.

#### Check Freshness

```
GET /api/v1/potatoes/{id}/freshness
```

Calculate the freshness status of a potato based on harvest date.

**Response:**
```json
{
  "id": "p001",
  "variety": "Russet",
  "freshness": "Fresh"
}
```

**Freshness Levels:**
- **Fresh**: 0-7 days since harvest
- **Good**: 8-30 days since harvest
- **Fair**: 31-90 days since harvest
- **Old**: 90+ days since harvest

### Inventory

#### Get Inventory Summary

```
GET /api/v1/inventory
```

Get a comprehensive inventory summary with totals and breakdown by variety.

**Response:**
```json
{
  "total_potatoes": 8,
  "total_weight": 3.11,
  "total_value": 27.51,
  "by_variety": [
    {
      "variety": "Russet",
      "total_quantity": 2,
      "total_weight": 0.97,
      "average_price": 2.89
    }
  ]
}
```

### Analytics

#### Get Analytics

```
GET /api/v1/analytics
```

Get analytics data including most popular variety, average weight, and quality distribution.

**Response:**
```json
{
  "most_popular_variety": "Russet",
  "average_weight": 0.38875,
  "premium_percentage": 62.5,
  "total_value": 27.51
}
```

### Recipes

#### Get All Recipes

```
GET /api/v1/recipes
GET /api/v1/recipes?variety=Russet
```

Retrieve all recipes or filter by potato variety.

**Response:**
```json
[
  {
    "id": "r001",
    "name": "Classic Baked Potato",
    "variety": "Russet",
    "cooking_time": 60,
    "difficulty": "Easy",
    "ingredients": [
      "1 large Russet potato",
      "2 tbsp butter",
      "Salt and pepper"
    ],
    "instructions": [
      "Preheat oven to 400°F",
      "Wash and dry potato thoroughly"
    ],
    "servings": 1
  }
]
```

#### Get Recipe by ID

```
GET /api/v1/recipes/{id}
```

Retrieve a specific recipe by ID.

#### Create Recipe

```
POST /api/v1/recipes
```

Add a new recipe to the database.

**Request Body:**
```json
{
  "id": "r007",
  "name": "Hasselback Potatoes",
  "variety": "Yukon Gold",
  "cooking_time": 50,
  "difficulty": "Medium",
  "ingredients": [
    "4 Yukon Gold potatoes",
    "4 tbsp butter",
    "Herbs"
  ],
  "instructions": [
    "Slice potatoes thinly without cutting through",
    "Brush with butter",
    "Bake at 425°F for 50 minutes"
  ],
  "servings": 4
}
```

#### Recommend Recipe

```
GET /api/v1/recipes/recommend?variety=Russet&difficulty=Easy
```

Get a recipe recommendation based on potato variety and optional difficulty level.

**Query Parameters:**
- `variety` (required): Potato variety
- `difficulty` (optional): Recipe difficulty (Easy, Medium, Hard)

## Project Structure

```
potato-demo/
├── main.go              # Application entry point and router setup
├── go.mod               # Go module dependencies
├── models/              # Data models
│   ├── potato.go
│   ├── recipe.go
│   └── inventory.go
├── storage/             # Data storage layer
│   └── storage.go
├── service/             # Business logic layer
│   ├── potato_service.go
│   └── recipe_service.go
├── handlers/            # HTTP handlers
│   ├── potato_handler.go
│   ├── recipe_handler.go
│   └── helpers.go
├── background/          # Background workers
│   └── worker.go
└── seed/                # Sample data
    └── seed.go
```

## Architecture

The application follows a layered architecture pattern:

1. **Models**: Define data structures and constants
2. **Storage**: In-memory storage with thread-safe operations
3. **Service**: Business logic and validation
4. **Handlers**: HTTP request/response handling
5. **Background**: Goroutines for periodic data updates
6. **Main**: Application initialization and routing

### Background Workers

The service includes three background goroutines that continuously update the system:

- **Potato Generator** (3s interval): Automatically adds new potatoes to inventory with random varieties, origins, and qualities
- **Recipe Generator** (8s interval): Creates new recipes for different potato varieties with varying difficulties
- **Quality Degradation** (20s interval): Simulates aging by downgrading potato quality over time:
  - Premium → Standard (after 30 days)
  - Standard → Economy (after 60 days)

These workers demonstrate Go's concurrency capabilities and make the demo more dynamic, even without incoming HTTP requests.

## Sample Data

The application comes preloaded with 8 sample potatoes and 6 recipes covering various varieties and cooking methods.

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- `200 OK`: Success
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Development

### Building

```bash
go build -o potato-service
```

### Running

```bash
./potato-service
```

The service runs on port 8080 by default.

## License

MIT License - feel free to use this for learning and demonstration purposes.

