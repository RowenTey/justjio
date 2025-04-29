package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/middleware"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/model/request"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// Shared test secret for JWT generation
const mockJWTSecret = "test-secret"

type BillHandlerTestSuite struct {
	suite.Suite
	app          *fiber.App
	db           *gorm.DB
	ctx          context.Context
	dependencies *tests.TestDependencies

	// Store IDs and tokens for reuse in tests
	testUser1ID    uint
	testUser2ID    uint
	testRoomID     string
	testUser1Token string
	testUser2Token string
}

// Helper function to generate JWT token for testing
func generateTestToken(userID uint, username, email string) (string, error) {
	claims := jwt.MapClaims{
		// Ensure user_id is string in claim like in likely real scenario
		"user_id":  userID,
		"username": username,
		"email":    email,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(mockJWTSecret))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (suite *BillHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error

	// Setup test containers
	suite.dependencies, err = tests.SetupTestDependencies(suite.ctx)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)
	fmt.Println("Test DB Connection String:", pgConnStr) // Log for debugging

	// Initialize database
	suite.db, err = database.InitTestDB(pgConnStr)
	assert.NoError(suite.T(), err)

	// Run migrations
	err = database.Migrate(suite.db)
	assert.NoError(suite.T(), err)

	// Setup Fiber app
	suite.app = fiber.New()

	suite.app.Use(middleware.Authenticated(mockJWTSecret))

	// IMPORTANT: In a real app, JWT middleware would be added here
	// e.g., suite.app.Use(jwtware.New(jwtware.Config{SigningKey: []byte(mockJWTSecret)}))
	// For these tests, we manually add the Auth header, assuming the middleware correctly
	// populates c.Locals("user") based on the token.

	// Register Bill routes
	billRoutes := suite.app.Group("/bills") // Group routes for clarity
	billRoutes.Post("/", CreateBill)
	billRoutes.Get("/", GetBillsByRoom) // Query param: ?roomId=...
	billRoutes.Post("/consolidate", ConsolidateBills)
	billRoutes.Get("/consolidated/:roomId", IsRoomBillConsolidated) // Path param
}

func (suite *BillHandlerTestSuite) TearDownSuite() {
	// Clean up containers
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
	log.Info("Tore down test suite dependencies")
}

func (suite *BillHandlerTestSuite) SetupTest() {
	// Assign the test DB to the global variable used by handlers/services
	database.DB = suite.db
	assert.NotNil(suite.T(), database.DB, "Global DB should be set")

	// Create User 1 (Host)
	hashedPassword1, _ := utils.HashPassword("password123")
	user1 := model.User{
		Username: "hostuser",
		Email:    "host@example.com",
		Password: hashedPassword1,
	}
	result := suite.db.Create(&user1)
	assert.NoError(suite.T(), result.Error)
	suite.testUser1ID = user1.ID
	token1, err := generateTestToken(user1.ID, user1.Username, user1.Email)
	assert.NoError(suite.T(), err)
	suite.testUser1Token = token1

	// Create User 2 (Payer)
	hashedPassword2, _ := utils.HashPassword("password456")
	user2 := model.User{
		Username: "payeruser",
		Email:    "payer@example.com",
		Password: hashedPassword2,
	}
	result = suite.db.Create(&user2)
	assert.NoError(suite.T(), result.Error)
	suite.testUser2ID = user2.ID
	token2, err := generateTestToken(user2.ID, user2.Username, user2.Email)
	assert.NoError(suite.T(), err)
	suite.testUser2Token = token2

	// Create Room hosted by User 1
	room := model.Room{
		Name:   "Test Room",
		HostID: suite.testUser1ID,
		Users:  []model.User{user1, user2}, // Add both users to the room
	}
	result = suite.db.Create(&room)
	assert.NoError(suite.T(), result.Error)
	suite.testRoomID = room.ID

	log.Infof("SetupTest complete: User1 ID=%d, User2 ID=%d, Room ID=%s", suite.testUser1ID, suite.testUser2ID, suite.testRoomID)
}

func (suite *BillHandlerTestSuite) TearDownTest() {
	// Clear database after each test using TRUNCATE for speed and cascade
	// Order matters if FK constraints exist and CASCADE isn't used/working properly
	suite.db.Exec("TRUNCATE TABLE transactions CASCADE")
	suite.db.Exec("TRUNCATE TABLE consolidations CASCADE")
	suite.db.Exec("TRUNCATE TABLE payers CASCADE")
	suite.db.Exec("TRUNCATE TABLE bills CASCADE")
	suite.db.Exec("TRUNCATE TABLE room_users CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms CASCADE")
	suite.db.Exec("TRUNCATE TABLE users CASCADE")

	// Reset the global DB variable
	database.DB = nil
	log.Info("Tore down test data and reset global DB")
}

func TestBillHandlerSuite(t *testing.T) {
	suite.Run(t, new(BillHandlerTestSuite))
}

func (suite *BillHandlerTestSuite) TestCreateBill_Success_WithOwner() {
	createReq := request.CreateBillRequest{
		RoomID:       suite.testRoomID,
		Name:         "Dinner",
		Amount:       100.50,
		IncludeOwner: true,
		Payers:       []uint{suite.testUser2ID}, // User1 (owner) implicitly included
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token) // User1 creates the bill

	resp, err := suite.app.Test(req, -1) // Use -1 timeout for test containers
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode) // Expect 201 Created

	// Verify response body
	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Created bill successfully", responseBody["message"])
	assert.NotNil(suite.T(), responseBody["data"])
	billData := responseBody["data"].(map[string]any)
	assert.NotEmpty(suite.T(), billData["id"])
	assert.Equal(suite.T(), createReq.Name, billData["name"])

	// Verify database
	var bill model.Bill
	err = suite.db.Preload("Payers").First(&bill, "id = ?", billData["id"]).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createReq.Amount, bill.Amount)
	assert.Equal(suite.T(), suite.testUser1ID, bill.OwnerID)
	assert.Equal(suite.T(), suite.testRoomID, bill.RoomID)
	assert.Equal(suite.T(), uint(0), bill.ConsolidationID) // Should be 0 if not consolidated yet

	// Check payers
	assert.Len(suite.T(), bill.Payers, 1)
	payerIDs := []uint{}
	for _, p := range bill.Payers {
		payerIDs = append(payerIDs, p.ID)
	}
	assert.Contains(suite.T(), payerIDs, suite.testUser2ID)
}

func (suite *BillHandlerTestSuite) TestCreateBill_Success_WithoutOwner() {
	createReq := request.CreateBillRequest{
		RoomID:       suite.testRoomID,
		Name:         "Drinks",
		Amount:       50.00,
		IncludeOwner: false,
		Payers:       []uint{suite.testUser2ID}, // Only User2 pays
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	// Verify response body
	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	billData := responseBody["data"].(map[string]any)

	// Verify database
	var bill model.Bill
	err = suite.db.Preload("Payers").First(&bill, "id = ?", billData["id"]).Error
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), bill.Payers, 1)
	assert.Equal(suite.T(), suite.testUser2ID, bill.Payers[0].ID) // Only User2 should be a payer
}

func (suite *BillHandlerTestSuite) TestCreateBill_InvalidInput_BadJSON() {
	reqBody := bytes.NewBuffer([]byte(`{"name": "test", "amount": "not-a-number"}`)) // Invalid JSON structure/type
	req := httptest.NewRequest(http.MethodPost, "/bills", reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Review your input", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestCreateBill_RoomNotFound() {
	nonExistentRoomID := uuid.New() // Generate a new UUID that doesn't exist in the DB
	createReq := request.CreateBillRequest{
		RoomID:       nonExistentRoomID.String(), // Use a UUID that doesn't exist
		Name:         "Ghost Bill",
		Amount:       10.00,
		IncludeOwner: true,
		Payers:       []uint{suite.testUser2ID},
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Room not found", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestCreateBill_PayerNotFound() {
	nonExistentUserID := uint(99999)
	createReq := request.CreateBillRequest{
		RoomID:       suite.testRoomID,
		Name:         "Bill for Nobody",
		Amount:       25.00,
		IncludeOwner: false,
		Payers:       []uint{nonExistentUserID}, // User ID that doesn't exist
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	// Depending on how GetUsersByID handles partial finds, this might be NotFound or potentially InternalError
	// Assuming it errors correctly if *any* user ID is not found.
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Payer(s) not found", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestCreateBill_NoPayersSpecifiedAndOwnerNotIncluded() {
	createReq := request.CreateBillRequest{
		RoomID:       suite.testRoomID,
		Name:         "Empty Bill",
		Amount:       30.00,
		IncludeOwner: false,
		Payers:       []uint{}, // Empty payers list
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Review your input", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestCreateBill_RoomAlreadyConsolidated() {
	// 1. Create a bill first to ensure the room exists and has bills
	bill := model.Bill{
		RoomID: suite.testRoomID, OwnerID: suite.testUser1ID, Name: "Bill 1", Amount: 100.00,
		Payers: []model.User{{ID: suite.testUser1ID}, {ID: suite.testUser2ID}},
	}
	err := suite.db.Create(&bill).Error
	assert.NoError(suite.T(), err)

	// 2. Create a Consolidation record for the room
	consolidation := model.Consolidation{}
	err = suite.db.Create(&consolidation).Error
	assert.NoError(suite.T(), err)

	// 3. Associate the bill with the consolidation
	err = suite.db.Table("bills").
		Where("room_id = ?", suite.testRoomID).
		Update("consolidation_id", consolidation.ID).
		Error
	assert.NoError(suite.T(), err)

	// 4. Attempt to create a new bill for the same room
	createReq := request.CreateBillRequest{
		RoomID:       suite.testRoomID,
		Name:         "Late Bill",
		Amount:       15.00,
		IncludeOwner: true,
		Payers:       []uint{suite.testUser2ID},
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/bills", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Bills for this room have already been consolidated", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestGetBillsByRoom_Success() {
	// Create a bill first
	bill := model.Bill{
		RoomID:  suite.testRoomID,
		OwnerID: suite.testUser1ID,
		Name:    "Existing Bill",
		Amount:  12.34,
	}
	err := suite.db.Create(&bill).Error
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bills?roomId=%s", suite.testRoomID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token) // Any user in the room should be able to view

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved bills successfully", responseBody["message"])
	assert.NotNil(suite.T(), responseBody["data"])

	billsData := responseBody["data"].([]any)
	assert.Len(suite.T(), billsData, 1)
	firstBill := billsData[0].(map[string]any)
	assert.Equal(suite.T(), bill.Name, firstBill["name"])
	assert.Equal(suite.T(), bill.Amount, float32(firstBill["amount"].(float64)))
}

func (suite *BillHandlerTestSuite) TestGetBillsByRoom_NoBills() {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bills?roomId=%s", suite.testRoomID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved bills successfully", responseBody["message"])
	assert.NotNil(suite.T(), responseBody["data"])
	billsData := responseBody["data"].([]any)
	assert.Len(suite.T(), billsData, 0) // Expect empty array
}

func (suite *BillHandlerTestSuite) TestGetBillsByRoom_MissingRoomIDQueryParam() {
	req := httptest.NewRequest(http.MethodGet, "/bills", nil) // No query param
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Review your input", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestConsolidateBills_Success() {
	// Create a couple of bills
	bill1 := model.Bill{
		RoomID: suite.testRoomID, OwnerID: suite.testUser1ID, Name: "Bill 1", Amount: 100.00,
		Payers: []model.User{{ID: suite.testUser1ID}, {ID: suite.testUser2ID}}, // Both pay
	}
	bill2 := model.Bill{
		RoomID: suite.testRoomID, OwnerID: suite.testUser2ID, Name: "Bill 2", Amount: 50.00,
		Payers: []model.User{{ID: suite.testUser1ID}}, // Only User1 pays
	}
	err := suite.db.Create(&bill1).Error
	assert.NoError(suite.T(), err)
	err = suite.db.Create(&bill2).Error
	assert.NoError(suite.T(), err)

	consolidateReq := request.ConsolidateBillsRequest{RoomID: suite.testRoomID}
	reqBody, _ := json.Marshal(consolidateReq)

	req := httptest.NewRequest(http.MethodPost, "/bills/consolidate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token) // Must be host (User1)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Bill consolidated successfully", responseBody["message"])
	assert.Nil(suite.T(), responseBody["data"]) // Handler returns nil data on success

	// Verify database state
	// 1. Bills are marked as consolidated (check associated consolidation ID)
	var bills []model.Bill
	err = suite.db.Find(&bills, "room_id = ?", suite.testRoomID).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), len(bills) == 2, "Should find exactly 2 bills for the room")
	assert.NotEqual(suite.T(), 0, bills[0].ConsolidationID, "Bills should have a consolidation ID")
	consolidationId := bills[0].ConsolidationID
	for _, b := range bills {
		assert.Equal(suite.T(), consolidationId, b.ConsolidationID, "All bills should have the same consolidation ID")
	}

	// 2. Consolidation record exists
	var consolidation model.Consolidation
	err = suite.db.First(&consolidation, "id = ?", consolidationId).Error
	assert.NoError(suite.T(), err, "Consolidation record should exist")

	// 3. Transactions are generated
	var transactions []model.Transaction
	err = suite.db.Find(&transactions, "consolidation_id = ?", consolidationId).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), len(transactions) > 0, "Transactions should be generated")
}

func (suite *BillHandlerTestSuite) TestConsolidateBills_NotHost() {
	consolidateReq := request.ConsolidateBillsRequest{RoomID: suite.testRoomID}
	reqBody, _ := json.Marshal(consolidateReq)

	req := httptest.NewRequest(http.MethodPost, "/bills/consolidate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser2Token) // User2 is NOT the host

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusUnauthorized, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User is not the host of the room", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestConsolidateBills_RoomNotFound() {
	nonExistentRoomID := uuid.New()
	consolidateReq := request.ConsolidateBillsRequest{RoomID: nonExistentRoomID.String()}
	reqBody, _ := json.Marshal(consolidateReq)

	req := httptest.NewRequest(http.MethodPost, "/bills/consolidate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token) // Host attempts, but room is wrong

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusNotFound, resp.StatusCode) // Service checks room existence

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Room not found", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestConsolidateBills_AlreadyConsolidated() {
	bill1 := model.Bill{
		RoomID:  suite.testRoomID,
		OwnerID: suite.testUser1ID,
		Name:    "Bill for Consolidation Test",
		Amount:  20.00,
		Payers:  []model.User{{ID: suite.testUser1ID}, {ID: suite.testUser2ID}},
	}
	err := suite.db.Create(&bill1).Error
	assert.NoError(suite.T(), err, "Failed to create prerequisite bill")

	// Consolidate once
	consolidateReq := request.ConsolidateBillsRequest{RoomID: suite.testRoomID}
	reqBody, _ := json.Marshal(consolidateReq)
	req1 := httptest.NewRequest(http.MethodPost, "/bills/consolidate", bytes.NewBuffer(reqBody))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+suite.testUser1Token)
	resp1, err1 := suite.app.Test(req1, -1)
	assert.NoError(suite.T(), err1)
	assert.Equal(suite.T(), fiber.StatusOK, resp1.StatusCode) // First consolidation should succeed

	// Attempt to consolidate again
	req2 := httptest.NewRequest(http.MethodPost, "/bills/consolidate", bytes.NewBuffer(reqBody)) // Reuse req body
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+suite.testUser1Token)
	resp2, err2 := suite.app.Test(req2, -1)
	assert.NoError(suite.T(), err2)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp2.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp2.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Bills for this room have already been consolidated", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestConsolidateBills_InvalidInput() {
	reqBody := bytes.NewBuffer([]byte(`{"invalid":`)) // Malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/bills/consolidate", reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusBadRequest, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Review your input", responseBody["message"])
}

func (suite *BillHandlerTestSuite) TestIsRoomBillConsolidated_False() {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bills/consolidated/%s", suite.testRoomID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved consolidation status successfully", responseBody["message"])
	data := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), false, data["isConsolidated"])
}

func (suite *BillHandlerTestSuite) TestIsRoomBillConsolidated_True() {
	// Consolidate first
	bill := model.Bill{
		RoomID:  suite.testRoomID,
		OwnerID: suite.testUser1ID,
		Name:    "Bill for Consolidation Test",
		Amount:  20.00,
		Payers:  []model.User{{ID: suite.testUser1ID}, {ID: suite.testUser2ID}},
	}
	err := suite.db.Create(&bill).Error
	assert.NoError(suite.T(), err, "Failed to create prerequisite bill")

	consolidateReq := request.ConsolidateBillsRequest{RoomID: suite.testRoomID}
	reqBody, _ := json.Marshal(consolidateReq)
	req1 := httptest.NewRequest(http.MethodPost, "/bills/consolidate", bytes.NewBuffer(reqBody))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+suite.testUser1Token)
	resp1, err1 := suite.app.Test(req1, -1)
	assert.NoError(suite.T(), err1)
	assert.Equal(suite.T(), fiber.StatusOK, resp1.StatusCode)

	// Check status
	req2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bills/consolidated/%s", suite.testRoomID), nil)
	req2.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp2, err2 := suite.app.Test(req2, -1)
	assert.NoError(suite.T(), err2)
	assert.Equal(suite.T(), fiber.StatusOK, resp2.StatusCode)

	var responseBody map[string]any
	err = json.NewDecoder(resp2.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved consolidation status successfully", responseBody["message"])
	data := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), true, data["isConsolidated"])
}

func (suite *BillHandlerTestSuite) TestIsRoomBillConsolidated_RoomNotFound() {
	// This test checks how the handler/service behaves if the room ID itself is invalid
	// The service function IsRoomBillConsolidated might return false (and no error)
	// if it just checks for a consolidation record by roomID without validating the room exists first.
	// Or it might error. Let's assume it returns false cleanly for a non-existent room's consolidation status.
	nonExistentRoomID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bills/consolidated/%s", nonExistentRoomID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fiber.StatusOK, resp.StatusCode) // Expecting OK because gorm.ErrRecordNotFound is handled

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Retrieved consolidation status successfully", responseBody["message"])
	data := responseBody["data"].(map[string]any)
	assert.Equal(suite.T(), false, data["isConsolidated"]) // Should report false for a room with no consolidation record
}

func (suite *BillHandlerTestSuite) TestIsRoomBillConsolidated_InvalidRoomIDFormat() {
	// Test with a string that is not a valid UUID
	invalidRoomID := "not-a-uuid"
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bills/consolidated/%s", invalidRoomID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser1Token)

	resp, err := suite.app.Test(req, -1)
	assert.NoError(suite.T(), err)
	// GORM might fail trying to query with an invalid UUID format, leading to an internal server error
	// OR Fiber's parameter binding/validation might catch it earlier if typed path parameters are used correctly.
	// If the handler/service tries to parse the UUID and fails, it should return BadRequest.
	// If it passes the invalid string to GORM, it might result in InternalServerError.
	// Let's assume GORM fails internally. Adjust if UUID parsing happens earlier.
	assert.Equal(suite.T(), fiber.StatusInternalServerError, resp.StatusCode) // GORM error likely results in 500

	var responseBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Error occured in server", responseBody["message"]) // Generic 500 message
}
