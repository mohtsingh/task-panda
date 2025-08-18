package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

type Task struct {
	ID                 int     `json:"id"`
	Category           string  `json:"category"`
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	Budget             float64 `json:"budget"`
	Location           string  `json:"location"`
	Date               string  `json:"date"`
	CreatedBy          int     `json:"created_by"`
	Status             string  `json:"status"`
	AcceptedProviderID *int    `json:"accepted_provider_id"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type Profile struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	Bio         string `json:"bio"`
	Role        string `json:"role"` // "CUSTOMER" or "SERVICE_PROVIDER"
	Photo       []byte `json:"-"`    // Exclude from JSON output
}

type Offer struct {
	ID           int     `json:"id"`
	TaskID       int     `json:"task_id"`
	ProviderID   int     `json:"provider_id"`
	OfferedPrice float64 `json:"offered_price"`
	Message      string  `json:"message"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	ProviderName string  `json:"provider_name,omitempty"`
}

type Chat struct {
	ID          int    `json:"id"`
	TaskID      int    `json:"task_id"`
	CustomerID  int    `json:"customer_id"`
	ProviderID  int    `json:"provider_id"`
	OfferID     int    `json:"offer_id"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	TaskTitle   string `json:"task_title,omitempty"`
	PartnerName string `json:"partner_name,omitempty"`
}

type Message struct {
	ID          int    `json:"id"`
	ChatID      int    `json:"chat_id"`
	SenderID    int    `json:"sender_id"`
	MessageText string `json:"message_text"`
	MessageType string `json:"message_type"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
	SenderName  string `json:"sender_name,omitempty"`
}

var db *sql.DB

func initDB() {
	var err error
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to the database!")
}

func createTask(c echo.Context) error {
	// Parse form data instead of JSON
	category := c.FormValue("category")
	title := c.FormValue("title")
	description := c.FormValue("description")
	budgetStr := c.FormValue("budget")
	location := c.FormValue("location")
	date := c.FormValue("date")
	createdByStr := c.FormValue("created_by")

	// Validate required fields
	if category == "" || title == "" || description == "" || budgetStr == "" || location == "" || date == "" || createdByStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
	}

	// Convert budget string to float64
	budget, err := strconv.ParseFloat(budgetStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid budget format"})
	}

	// Convert created_by to int
	createdBy, err := strconv.Atoi(createdByStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid created_by format"})
	}

	// Create task object
	newTask := Task{
		Category:    category,
		Title:       title,
		Description: description,
		Budget:      budget,
		Location:    location,
		Date:        date,
		CreatedBy:   createdBy,
		Status:      "OPEN",
	}

	// Handle optional image upload
	file, header, err := c.Request().FormFile("image")
	var imageData []byte
	if err == nil {
		defer file.Close()
		imageData, err = ioutil.ReadAll(file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read image"})
		}
		fmt.Printf("Uploaded image: %s, size: %d bytes\n", header.Filename, len(imageData))
		// You can save imageData to database or file system here
	}

	// Insert task into database
	query := `INSERT INTO tasks (category, title, description, budget, location, date, created_by, status) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	err = db.QueryRow(query, newTask.Category, newTask.Title, newTask.Description, newTask.Budget,
		newTask.Location, newTask.Date, newTask.CreatedBy, newTask.Status).Scan(&newTask.ID, &newTask.CreatedAt, &newTask.UpdatedAt)
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to insert task"})
	}

	fmt.Printf("Task created successfully: %+v\n", newTask)
	return c.JSON(http.StatusCreated, newTask)
}

func getTaskByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID format"})
	}

	var task Task
	query := `SELECT id, category, title, description, budget, location, date, created_by, status, 
	          accepted_provider_id, created_at, updated_at FROM tasks WHERE id = $1`
	err = db.QueryRow(query, id).Scan(&task.ID, &task.Category, &task.Title, &task.Description,
		&task.Budget, &task.Location, &task.Date, &task.CreatedBy, &task.Status,
		&task.AcceptedProviderID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch task"})
	}

	return c.JSON(http.StatusOK, task)
}

func getAllTasks(c echo.Context) error {
	rows, err := db.Query(`SELECT id, category, title, description, budget, location, date, 
	                      created_by, status, accepted_provider_id, created_at, updated_at FROM tasks 
	                      ORDER BY created_at DESC`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch tasks"})
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Category, &t.Title, &t.Description, &t.Budget,
			&t.Location, &t.Date, &t.CreatedBy, &t.Status, &t.AcceptedProviderID,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse task data"})
		}
		tasks = append(tasks, t)
	}

	return c.JSON(http.StatusOK, tasks)
}

// Create an offer for a task
func createOffer(c echo.Context) error {
	taskIDStr := c.FormValue("task_id")
	providerIDStr := c.FormValue("provider_id")
	offeredPriceStr := c.FormValue("offered_price")
	message := c.FormValue("message")

	if taskIDStr == "" || providerIDStr == "" || offeredPriceStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "task_id, provider_id, and offered_price are required"})
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task_id format"})
	}

	providerID, err := strconv.Atoi(providerIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid provider_id format"})
	}

	offeredPrice, err := strconv.ParseFloat(offeredPriceStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid offered_price format"})
	}

	// Check if task exists and is open
	var taskStatus string
	var customerID int
	err = db.QueryRow(`SELECT status, created_by FROM tasks WHERE id = $1`, taskID).Scan(&taskStatus, &customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Task not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to check task"})
	}

	if taskStatus != "OPEN" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Task is not open for offers"})
	}

	// Check if provider already made an offer
	var existingOfferID int
	err = db.QueryRow(`SELECT id FROM offers WHERE task_id = $1 AND provider_id = $2`, taskID, providerID).Scan(&existingOfferID)
	if err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "You have already made an offer for this task"})
	}

	// Create the offer
	var offerID int
	var createdAt, updatedAt time.Time
	query := `INSERT INTO offers (task_id, provider_id, offered_price, message) 
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err = db.QueryRow(query, taskID, providerID, offeredPrice, message).Scan(&offerID, &createdAt, &updatedAt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create offer"})
	}

	// Create a chat for this offer
	var chatID int
	chatQuery := `INSERT INTO chats (task_id, customer_id, provider_id, offer_id) 
	              VALUES ($1, $2, $3, $4) RETURNING id`
	err = db.QueryRow(chatQuery, taskID, customerID, providerID, offerID).Scan(&chatID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create chat"})
	}

	// Send initial message about the offer
	messageText := fmt.Sprintf("I'm interested in your task and would like to offer my services for $%.2f. %s", offeredPrice, message)
	_, err = db.Exec(`INSERT INTO messages (chat_id, sender_id, message_text, message_type) 
	                  VALUES ($1, $2, $3, $4)`, chatID, providerID, messageText, "OFFER_UPDATE")
	if err != nil {
		log.Printf("Failed to create initial message: %v", err)
	}

	offer := Offer{
		ID:           offerID,
		TaskID:       taskID,
		ProviderID:   providerID,
		OfferedPrice: offeredPrice,
		Message:      message,
		Status:       "PENDING",
		CreatedAt:    createdAt.Format(time.RFC3339),
		UpdatedAt:    updatedAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusCreated, offer)
}

// Get all offers for a task
func getTaskOffers(c echo.Context) error {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task_id format"})
	}

	query := `SELECT o.id, o.task_id, o.provider_id, o.offered_price, o.message, o.status, 
	          o.created_at, o.updated_at, p.full_name 
	          FROM offers o 
	          JOIN profiles p ON o.provider_id = p.id 
	          WHERE o.task_id = $1 ORDER BY o.created_at ASC`

	rows, err := db.Query(query, taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch offers"})
	}
	defer rows.Close()

	var offers []Offer
	for rows.Next() {
		var o Offer
		err := rows.Scan(&o.ID, &o.TaskID, &o.ProviderID, &o.OfferedPrice, &o.Message,
			&o.Status, &o.CreatedAt, &o.UpdatedAt, &o.ProviderName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse offer data"})
		}
		offers = append(offers, o)
	}

	return c.JSON(http.StatusOK, offers)
}

// Accept an offer
func acceptOffer(c echo.Context) error {
	offerIDStr := c.Param("offer_id")
	offerID, err := strconv.Atoi(offerIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid offer_id format"})
	}

	// Get offer details
	var offer Offer
	var customerID int
	query := `SELECT o.id, o.task_id, o.provider_id, o.offered_price, o.status, t.created_by
	          FROM offers o 
	          JOIN tasks t ON o.task_id = t.id 
	          WHERE o.id = $1`
	err = db.QueryRow(query, offerID).Scan(&offer.ID, &offer.TaskID, &offer.ProviderID,
		&offer.OfferedPrice, &offer.Status, &customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Offer not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch offer"})
	}

	if offer.Status != "PENDING" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Offer is not in pending status"})
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to start transaction"})
	}
	defer tx.Rollback()

	// Update task status and set accepted provider
	_, err = tx.Exec(`UPDATE tasks SET status = 'ACCEPTED', accepted_provider_id = $1 
	                  WHERE id = $2`, offer.ProviderID, offer.TaskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update task"})
	}

	// Accept this offer
	_, err = tx.Exec(`UPDATE offers SET status = 'ACCEPTED' WHERE id = $1`, offerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to accept offer"})
	}

	// Reject all other offers for this task
	_, err = tx.Exec(`UPDATE offers SET status = 'REJECTED' 
	                  WHERE task_id = $1 AND id != $2`, offer.TaskID, offerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to reject other offers"})
	}

	// Deactivate all other chats for this task
	_, err = tx.Exec(`UPDATE chats SET is_active = false 
	                  WHERE task_id = $1 AND offer_id != $2`, offer.TaskID, offerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to deactivate chats"})
	}

	// Add system message to the accepted chat
	var acceptedChatID int
	err = tx.QueryRow(`SELECT id FROM chats WHERE offer_id = $1`, offerID).Scan(&acceptedChatID)
	if err == nil {
		_, err = tx.Exec(`INSERT INTO messages (chat_id, sender_id, message_text, message_type) 
		                  VALUES ($1, $2, $3, $4)`, acceptedChatID, customerID,
			"Congratulations! Your offer has been accepted. Let's get started!", "SYSTEM")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message":  "Offer accepted successfully",
		"offer_id": offerID,
		"task_id":  offer.TaskID,
	})
}

// Get chats for a user (both as customer and provider)
func getUserChats(c echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid user_id format"})
	}

	query := `SELECT c.id, c.task_id, c.customer_id, c.provider_id, c.offer_id, c.is_active, 
	          c.created_at, c.updated_at, t.title,
	          CASE 
	            WHEN c.customer_id = $1 THEN p_provider.full_name 
	            ELSE p_customer.full_name 
	          END as partner_name
	          FROM chats c
	          JOIN tasks t ON c.task_id = t.id
	          JOIN profiles p_customer ON c.customer_id = p_customer.id
	          JOIN profiles p_provider ON c.provider_id = p_provider.id
	          WHERE c.customer_id = $1 OR c.provider_id = $1
	          ORDER BY c.updated_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch chats"})
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var chat Chat
		err := rows.Scan(&chat.ID, &chat.TaskID, &chat.CustomerID, &chat.ProviderID,
			&chat.OfferID, &chat.IsActive, &chat.CreatedAt, &chat.UpdatedAt,
			&chat.TaskTitle, &chat.PartnerName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse chat data"})
		}
		chats = append(chats, chat)
	}

	return c.JSON(http.StatusOK, chats)
}

// Get messages for a chat
func getChatMessages(c echo.Context) error {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid chat_id format"})
	}

	query := `SELECT m.id, m.chat_id, m.sender_id, m.message_text, m.message_type, 
	          m.is_read, m.created_at, p.full_name
	          FROM messages m
	          JOIN profiles p ON m.sender_id = p.id
	          WHERE m.chat_id = $1
	          ORDER BY m.created_at ASC`

	rows, err := db.Query(query, chatID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch messages"})
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.MessageText,
			&msg.MessageType, &msg.IsRead, &msg.CreatedAt, &msg.SenderName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to parse message data"})
		}
		messages = append(messages, msg)
	}

	return c.JSON(http.StatusOK, messages)
}

// Send a message in a chat
func sendMessage(c echo.Context) error {
	chatIDStr := c.FormValue("chat_id")
	senderIDStr := c.FormValue("sender_id")
	messageText := c.FormValue("message_text")

	if chatIDStr == "" || senderIDStr == "" || messageText == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "chat_id, sender_id, and message_text are required"})
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid chat_id format"})
	}

	senderID, err := strconv.Atoi(senderIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid sender_id format"})
	}

	// Check if chat is active
	var isActive bool
	err = db.QueryRow(`SELECT is_active FROM chats WHERE id = $1`, chatID).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Chat not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to check chat status"})
	}

	if !isActive {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Chat is not active"})
	}

	// Insert message
	var messageID int
	var createdAt time.Time
	query := `INSERT INTO messages (chat_id, sender_id, message_text, message_type) 
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err = db.QueryRow(query, chatID, senderID, messageText, "TEXT").Scan(&messageID, &createdAt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to send message"})
	}

	// Update chat's updated_at timestamp
	_, err = db.Exec(`UPDATE chats SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`, chatID)
	if err != nil {
		log.Printf("Failed to update chat timestamp: %v", err)
	}

	message := Message{
		ID:          messageID,
		ChatID:      chatID,
		SenderID:    senderID,
		MessageText: messageText,
		MessageType: "TEXT",
		IsRead:      false,
		CreatedAt:   createdAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusCreated, message)
}

// Create a new profile with photo upload
func createProfile(c echo.Context) error {
	// Parse form data (including file)
	fullName := c.FormValue("full_name")
	email := c.FormValue("email")
	address := c.FormValue("address")
	phone := c.FormValue("phone_number")
	bio := c.FormValue("bio")
	role := c.FormValue("role")

	if fullName == "" || email == "" || role == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing required fields"})
	}

	// Check if email already exists
	var existingEmail string
	err := db.QueryRow(`SELECT email FROM profiles WHERE email = $1`, email).Scan(&existingEmail)
	if err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "Email already exists"})
	}

	// Retrieve the photo file from the request
	file, _, err := c.Request().FormFile("photo")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Photo is required"})
	}
	defer file.Close()

	// Read the photo file into a byte slice (binary data)
	photoBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to read photo"})
	}

	// Insert profile and photo into the database
	query := `INSERT INTO profiles (full_name, email, address, phone_number, bio, role, photo)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var id int
	err = db.QueryRow(query, fullName, email, address, phone, bio, role, photoBytes).Scan(&id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create profile"})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "Profile created",
		"profile": echo.Map{
			"id":       id,
			"fullName": fullName,
			"email":    email,
			"role":     role,
		},
	})
}

// Get profile by email
func getProfileByEmail(c echo.Context) error {
	email := c.Param("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Email is required"})
	}

	var profile struct {
		ID          int
		FullName    string
		Email       string
		Address     string
		PhoneNumber string
		Bio         string
		Role        string
		Photo       []byte
	}

	query := `SELECT id, full_name, email, address, phone_number, bio, role, photo 
              FROM profiles WHERE email = $1`
	err := db.QueryRow(query, email).Scan(&profile.ID, &profile.FullName, &profile.Email, &profile.Address,
		&profile.PhoneNumber, &profile.Bio, &profile.Role, &profile.Photo)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Profile not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch profile"})
	}

	return c.JSON(http.StatusOK, profile)
}

// Mark messages as read
func markMessagesAsRead(c echo.Context) error {
	chatIDStr := c.FormValue("chat_id")
	userIDStr := c.FormValue("user_id")

	if chatIDStr == "" || userIDStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "chat_id and user_id are required"})
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid chat_id format"})
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid user_id format"})
	}

	// Mark all messages in this chat as read for messages not sent by this user
	_, err = db.Exec(`UPDATE messages SET is_read = true 
	                  WHERE chat_id = $1 AND sender_id != $2 AND is_read = false`, chatID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to mark messages as read"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Messages marked as read"})
}

// Get unread message count for a user
func getUnreadCount(c echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid user_id format"})
	}

	var count int
	query := `SELECT COUNT(*) FROM messages m
	          JOIN chats c ON m.chat_id = c.id
	          WHERE (c.customer_id = $1 OR c.provider_id = $1)
	          AND m.sender_id != $1 
	          AND m.is_read = false
	          AND c.is_active = true`

	err = db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to get unread count"})
	}

	return c.JSON(http.StatusOK, echo.Map{"unread_count": count})
}

// Update task status (for completing tasks, etc.)
func updateTaskStatus(c echo.Context) error {
	taskIDStr := c.Param("task_id")
	status := c.FormValue("status")

	if status == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Status is required"})
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid task_id format"})
	}

	// Validate status
	validStatuses := map[string]bool{
		"OPEN": true, "ACCEPTED": true, "IN_PROGRESS": true,
		"COMPLETED": true, "CANCELLED": true,
	}
	if !validStatuses[status] {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid status"})
	}

	_, err = db.Exec(`UPDATE tasks SET status = $1 WHERE id = $2`, status, taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update task status"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Task status updated successfully"})
}

func main() {
	initDB()
	defer db.Close()

	e := echo.New()

	// Enable CORS with more specific settings
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},                                                                // Specify allowed origins
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}, // Allowed methods
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},                   // Allowed headers
		AllowCredentials: true,                                                                         // Allow credentials (cookies, etc.)
	}))

	// Task routes
	e.POST("/tasks", createTask)
	e.GET("/tasks/:id", getTaskByID)
	e.GET("/tasks", getAllTasks)
	e.PUT("/tasks/:task_id/status", updateTaskStatus)

	// Profile routes
	e.POST("/profile", createProfile)
	e.GET("/profile/:email", getProfileByEmail)

	// Offer routes
	e.POST("/offers", createOffer)
	e.GET("/tasks/:task_id/offers", getTaskOffers)
	e.POST("/offers/:offer_id/accept", acceptOffer)

	// Chat routes
	e.GET("/users/:user_id/chats", getUserChats)
	e.GET("/chats/:chat_id/messages", getChatMessages)
	e.POST("/messages", sendMessage)
	e.POST("/messages/read", markMessagesAsRead)
	e.GET("/users/:user_id/unread-count", getUnreadCount)

	e.Logger.Fatal(e.Start(":8080"))
}
