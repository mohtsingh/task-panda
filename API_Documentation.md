
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

### Accept an Offer  
**POST** `/offers/:offer_id/accept`  

**Example:** `/offers/5/accept`

---

## 🔔 Notification Routes

### Register Device Token  
**POST** `/notifications/fcm/token`  
**Content-Type:** `application/json`  

**JSON Body:**
```json
{
  "profile_id": 1,
  "token": "fcm-token-or-apns-token-here",
  "platform": "android"
}
```

**Parameters:**
- `profile_id`: int (required) - User profile ID  
- `token`: string (required) - FCM/APNs device token  
- `platform`: string (optional) - "android", "ios", or "web"

**Response:**
```json
{
  "message": "Device token registered successfully",
}
```

**Note:** Links device tokens to profiles for push notifications. Updates existing token if profile+platform already exists.

---
