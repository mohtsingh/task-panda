# Task Panda API Documentation

_Last Updated: 2025-08-19_

## Base URL

```
https://task-panda.onrender.com
```

---

## 📦 Task Routes

### Create Task
**POST** `/tasks`  
**Form Data:**  
- `category`: string (required)  
- `title`: string (required)  
- `description`: string (required)  
- `budget`: float (required)  
- `location`: string (required)  
- `date`: string (required)  
- `created_by`: int (required)  
- `image`: file (optional)  

**Example (form-data)**:
```
category=Plumbing
title=Fix leaking pipe
description=Pipe leaking in kitchen
budget=150.50
location=New Delhi
date=2025-08-21
created_by=1
image=file.jpg
```

---

### Get Task by ID  
**GET** `/tasks/:id`

**Example:** `/tasks/1`

---

### Get All Tasks  
**GET** `/tasks`

---

### Update Task Status  
**PUT** `/tasks/:task_id/status`  
**Form Data:**  
- `status`: OPEN, ACCEPTED, IN_PROGRESS, COMPLETED, CANCELLED

**Example:** `/tasks/1/status`  
```
status=COMPLETED
```

---

## 👤 Profile Routes

### Create Profile  
**POST** `/profile`  
**Form Data:**  
- `full_name`: string (required)  
- `email`: string (required)  
- `address`: string  
- `phone_number`: string  
- `bio`: string  
- `role`: CUSTOMER or SERVICE_PROVIDER (required)  
- `photo`: file (required)

**Example (form-data)**:
```
full_name=John Doe
email=john@example.com
address=123 Main St
phone_number=1234567890
bio=Experienced plumber
role=SERVICE_PROVIDER
photo=file.jpg
```

---

### Get Profile by Email  
**GET** `/profile/:email`

**Example:** `/profile/john@example.com`

---

## 💼 Offer Routes

### Create Offer  
**POST** `/offers`  
**Form Data:**  
- `task_id`: int (required)  
- `provider_id`: int (required)  
- `offered_price`: float (required)  
- `message`: string  

**Example (form-data)**:
```
task_id=1
provider_id=2
offered_price=200
message=I can complete it by tomorrow.
```

---

### Get Offers for a Task  
**GET** `/tasks/:task_id/offers`  

**Example:** `/tasks/1/offers`

---

### Update Offer  
**PUT** `/offers/:offer_id`  
**Content-Type:** `appliction/json`  
**JSON Body:**
```
{
  "offered_price": 175.50,
  "message": "Updated offer with better timeline"
}
```
**Parameters:**
- `offered_price`: float (optional)  
- `message`: string (optional)  

**Example:** `/offers/1`  

**Note:** Only pending offers can be updated. You can update either price, message, or both fields.

---

### Accept an Offer  
**POST** `/offers/:offer_id/accept`  

**Example:** `/offers/5/accept`

---
