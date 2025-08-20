# Chat Functionality API Documentation

## Overview
This API provides chat functionality for task-based service marketplace where service providers can make offers on tasks and communicate with customers.

## Workflow
1. **Task Creation**: Customer creates a task
2. **Offer Submission**: Service providers submit offers with prices
3. **Chat Creation**: Each offer automatically creates a chat between customer and provider
4. **Communication**: Both parties can exchange messages
5. **Offer Acceptance**: Customer accepts one offer, disabling other chats
6. **Task Completion**: Work proceeds through the accepted chat

## API Endpoints

### 1. Task Management

#### Create Task
```http
POST /tasks
Content-Type: multipart/form-data
```

**Parameters:**
- `category` (string, required): Task category
- `title` (string, required): Task title
- `description` (string, required): Task description
- `budget` (string, required): Expected budget
- `location` (string, required): Task location
- `date` (string, required): Task date
- `created_by` (string, required): Customer profile ID
- `image` (file, optional): Task image

**Response:**
```json
{
  "id": 1,
  "category": "Home Repair",
  "title": "Fix Kitchen Sink",
  "description": "Kitchen sink is leaking",
  "budget": 150.00,
  "location": "New York, NY",
  "date": "2025-08-20",
  "created_by": 1,
  "status": "OPEN",
  "accepted_provider_id": null,
  "created_at": "2025-08-18T10:30:00Z",
  "updated_at": "2025-08-18T10:30:00Z"
}
```

#### Get All Tasks
```http
GET /tasks
```

#### Get Task by ID
```http
GET /tasks/{id}
```

#### Update Task Status
```http
PUT /tasks/{task_id}/status
Content-Type: application/x-www-form-urlencoded
```

**Parameters:**
- `status` (string): One of "OPEN", "ACCEPTED", "IN_PROGRESS", "COMPLETED", "CANCELLED"

### 2. Offer Management

#### Create Offer
```http
POST /offers
Content-Type: application/x-www-form-urlencoded
```

**Parameters:**
- `task_id` (string, required): Task ID
- `provider_id` (string, required): Service provider profile ID
- `offered_price` (string, required): Offered price
- `message` (string, optional): Initial message

**Response:**
```json
{
  "id": 1,
  "task_id": 1,
  "provider_id": 2,
  "offered_price": 120.00,
  "message": "I can fix your sink today",
  "status": "PENDING",
  "created_at": "2025-08-18T11:00:00Z",
  "updated_at": "2025-08-18T11:00:00Z"
}
```

**Note:** This automatically creates a chat between customer and provider.

#### Get Task Offers
```http
GET /tasks/{task_id}/offers
```

**Response:**
```json
[
  {
    "id": 1,
    "task_id": 1,
    "provider_id": 2,
    "offered_price": 120.00,
    "message": "I can fix your sink today",
    "status": "PENDING",
    "created_at": "2025-08-18T11:00:00Z",
    "updated_at": "2025-08-18T11:00:00Z",
    "provider_name": "John Doe"
  }
]
```

#### Accept Offer
```http
POST /offers/{offer_id}/accept
```

**Response:**
```json
{
  "message": "Offer accepted successfully",
  "offer_id": 1,
  "task_id": 1
}
```

**Note:** This will:
- Update task status to "ACCEPTED"
- Set the provider as accepted_provider_id
- Reject all other offers for this task
- Deactivate all other chats for this task
- Add a system message to the accepted chat

### 3. Chat Management

#### Get User Chats
```http
GET /users/{user_id}/chats
```

**Response:**
```json
[
  {
    "id": 1,
    "task_id": 1,
    "customer_id": 1,
    "provider_id": 2,
    "offer_id": 1,
    "is_active": true,
    "created_at": "2025-08-18T11:00:00Z",
    "updated_at": "2025-08-18T11:30:00Z",
    "task_title": "Fix Kitchen Sink",
    "partner_name": "John Doe"
  }
]
```

#### Get Chat Messages
```http
GET /chats/{chat_id}/messages
```

**Response:**
```json
[
  {
    "id": 1,
    "chat_id": 1,
    "sender_id": 2,
    "message_text": "I'm interested in your task and would like to offer my services for $120.00. I can fix your sink today",
    "message_type": "OFFER_UPDATE",
    "is_read": false,
    "created_at": "2025-08-18T11:00:00Z",
    "sender_name": "John Doe"
  },
  {
    "id": 2,
    "chat_id": 1,
    "sender_id": 1,
    "message_text": "That sounds good. When can you start?",
    "message_type": "TEXT",
    "is_read": true,
    "created_at": "2025-08-18T11:15:00Z",
    "sender_name": "Jane Smith"
  }
]
```

#### Send Message
```http
POST /messages
Content-Type: application/x-www-form-urlencoded
```

**Parameters:**
- `chat_id` (string, required): Chat ID
- `sender_id` (string, required): Sender profile ID
- `message_text` (string, required): Message content

**Response:**
```json
{
  "id": 3,
  "chat_id": 1,
  "sender_id": 1,
  "message_text": "That sounds good. When can you start?",
  "message_type": "TEXT",
  "is_read": false,
  "created_at": "2025-08-18T11:15:00Z"
}
```

#### Mark Messages as Read
```http
POST /messages/read
Content-Type: application/x-www-form-urlencoded
```

**Parameters:**
- `chat_id` (string, required): Chat ID
- `user_id` (string, required): User ID marking messages as read

#### Get Unread Message Count
```http
GET /users/{user_id}/unread-count
```

**Response:**
```json
{
  "unread_count": 3
}
```

### 4. Profile Management

#### Create Profile
```http
POST /profile
Content-Type: multipart/form-data
```

**Parameters:**
- `full_name` (string, required)
- `email` (string, required)
- `address` (string, optional)
- `phone_number` (string, optional)
- `bio` (string, optional)
- `role` (string, required): "CUSTOMER" or "SERVICE_PROVIDER"
- `photo` (file, required): Profile photo

#### Get Profile by Email
```http
GET /profile/{email}
```

## Database Schema Summary

### Tables Created:
1. **tasks** - Extended with status tracking and owner info
2. **offers** - Store service provider offers
3. **chats** - Individual chat conversations
4. **messages** - Chat messages
5. **profiles** - User profiles (existing)

### Key Relationships:
- Each offer creates a chat
- Each chat belongs to one task, one customer, one provider
- Messages belong to chats
- When offer is accepted, other chats are deactivated

## Message Types:
- `TEXT` - Regular chat message
- `OFFER_UPDATE` - Initial offer message
- `SYSTEM` - System-generated messages (like acceptance confirmation)

## Task Status Flow:
1. `OPEN` - Available for offers
2. `ACCEPTED` - Offer accepted, provider assigned
3. `IN_PROGRESS` - Work started
4. `COMPLETED` - Work finished
5. `CANCELLED` - Task cancelled

## Usage Examples

### Complete Flow Example:

1. **Customer creates task:**
```bash
curl -X POST http://localhost:8080/tasks \
  -F "category=Home Repair" \
  -F "title=Fix Kitchen Sink" \
  -F "description=Kitchen sink is leaking" \
  -F "budget=150" \
  -F "location=New York, NY" \
  -F "date=2025-08-20" \
  -F "created_by=1"
```

2. **Service provider makes offer:**
```bash
curl -X POST http://localhost:8080/offers \
  -d "task_id=1&provider_id=2&offered_price=120&message=I can fix this today"
```

3. **Customer views offers:**
```bash
curl http://localhost:8080/tasks/1/offers
```

4. **Customer and provider chat:**
```bash
curl -X POST http://localhost:8080/messages \
  -d "chat_id=1&sender_id=1&message_text=When can you start?"
```

5. **Customer accepts offer:**
```bash
curl -X POST http://localhost:8080/offers/1/accept
```

## Error Handling

All endpoints return appropriate HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `404` - Not Found
- `409` - Conflict (e.g., duplicate offer)
- `500` - Internal Server Error

Error responses include a descriptive message:
```json
{
  "error": "Task not found"
}
```
