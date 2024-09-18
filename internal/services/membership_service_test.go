package services

import (
	"api/internal/database"
	"api/internal/models"
	"api/internal/testhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMembershipService_CreateMembership_NotPaid(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		100.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membershipService := NewMembershipService(db)

	req := models.CreateMembershipRequest{
		UserID:     user1.ID,
		SemesterID: semester1.ID.String(),
		Paid:       false,
		Discounted: false,
	}
	membership, err := membershipService.CreateMembership(&req)
	if err != nil {
		t.Errorf("CreateMembership() error = %v", err)
		return
	}

	if membership.ID.String() == "" {
		t.Errorf("ID is empty")
		return
	}

	if membership.SemesterID != semesterId {
		t.Errorf("SemesterID = %v, want = %v", membership.SemesterID, semester1.ID)
		return
	}

	if membership.UserID != user1.ID {
		t.Errorf("UserID = %v, want = %v", membership.UserID, user1.ID)
		return
	}

	if membership.Paid != false {
		t.Errorf("Paid = %v, want = %v", membership.Paid, false)
		return
	}

	if membership.Discounted != false {
		t.Errorf("Discounted = %v, want = %v", membership.Discounted, false)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}
	res := db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}

	if !almostEqual(sem.CurrentBudget, 100.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 100.0)
		return
	}
}

func TestMembershipService_CreateMembership_PaidNotDiscounted(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		100.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membershipService := NewMembershipService(db)

	req := models.CreateMembershipRequest{
		UserID:     user1.ID,
		SemesterID: semester1.ID.String(),
		Paid:       true,
		Discounted: false,
	}
	membership, err := membershipService.CreateMembership(&req)
	if err != nil {
		t.Errorf("CreateMembership() error = %v", err)
		return
	}

	if membership.SemesterID != semesterId {
		t.Errorf("SemesterID = %v, want = %v", membership.SemesterID, semester1.ID)
		return
	}

	if membership.UserID != user1.ID {
		t.Errorf("UserID = %v, want = %v", membership.UserID, user1.ID)
		return
	}

	if membership.Paid != true {
		t.Errorf("Paid = %v, want = %v", membership.Paid, true)
		return
	}

	if membership.Discounted != false {
		t.Errorf("Discounted = %v, want = %v", membership.Discounted, false)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}
	res := db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}

	if !almostEqual(sem.CurrentBudget, 110.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 110.0)
		return
	}
}

func TestMembershipService_CreateMembership_PaidDiscounted(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		100.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membershipService := NewMembershipService(db)

	req := models.CreateMembershipRequest{
		UserID:     user1.ID,
		SemesterID: semester1.ID.String(),
		Paid:       true,
		Discounted: true,
	}
	membership, err := membershipService.CreateMembership(&req)
	if err != nil {
		t.Errorf("CreateMembership() error = %v", err)
		return
	}

	if membership.SemesterID != semesterId {
		t.Errorf("SemesterID = %v, want = %v", membership.SemesterID, semester1.ID)
		return
	}

	if membership.UserID != user1.ID {
		t.Errorf("UserID = %v, want = %v", membership.UserID, user1.ID)
		return
	}

	if membership.Paid != true {
		t.Errorf("Paid = %v, want = %v", membership.Paid, true)
		return
	}

	if membership.Discounted != true {
		t.Errorf("Discounted = %v, want = %v", membership.Discounted, true)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}
	res := db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}

	if !almostEqual(sem.CurrentBudget, 107.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 107.0)
		return
	}
}

func TestMembershipService_GetMembership(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semester1, err := testhelpers.CreateSemester(
		db,
		uuid.New(),
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		110.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: false,
	}
	res := db.Create(&membership)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)

	found, err := membershipService.GetMembership(membership.ID)
	if err != nil {
		t.Errorf("GetMembership() error = %v", err)
		return
	}

	if found.ID != membership.ID {
		t.Errorf("ID = %v, want = %v", found.ID, membership.ID)
		return
	}
}

func TestMembershipService_ListMemberships(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}
	user2, err := testhelpers.CreateUser(db, 46372894, "deep", "kalra", "deep@gmail.com", models.FacultyArts, "dkal")
	if err != nil {
		t.Fatalf(err.Error())
	}
	user3, err := testhelpers.CreateUser(db, 41694873, "jane", "doe", "jane@gmail.com", models.FacultyMath, "jdane")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semester1, err := testhelpers.CreateSemester(
		db,
		uuid.New(),
		"Fall 2022",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		110.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
	semester2, err := testhelpers.CreateSemester(
		db,
		uuid.New(),
		"Spring 2023",
		"",
		time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		107.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership1 := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       false,
		Discounted: false,
	}
	res := db.Create(&membership1)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membership2 := models.Membership{
		UserID:     user2.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: false,
	}
	res = db.Create(&membership2)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membership3 := models.Membership{
		UserID:     user3.ID,
		SemesterID: semester2.ID,
		Paid:       true,
		Discounted: true,
	}
	res = db.Create(&membership3)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)

	filter := models.ListMembershipsFilter{
		SemesterID: &semester1.ID,
	}

	memberships, err := membershipService.ListMemberships(&filter)
	if err != nil {
		t.Errorf("ListMemberships() error = %v", err)
		return
	}

	if len(memberships) != 2 {
		t.Errorf("length = %v, want = %v", len(memberships), 2)
		return
	}

	if memberships[0].UserID != user1.ID && memberships[1].UserID != user2.ID {
		t.Errorf("result = %v, want = %v", memberships, []uint64{user1.ID, user2.ID})
		return
	}

	if memberships[0].Attendance != 0 || memberships[1].Attendance != 0 {
		t.Errorf("ListMemberships attendance incorrect: %v, expected 0", memberships)
		return
	}
}

func TestMembershipService_ListMemberships_Limit(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}
	user2, err := testhelpers.CreateUser(db, 46372894, "deep", "kalra", "deep@gmail.com", models.FacultyArts, "dkal")
	if err != nil {
		t.Fatalf(err.Error())
	}
	user3, err := testhelpers.CreateUser(db, 41694873, "jane", "doe", "jane@gmail.com", models.FacultyMath, "jdane")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semester1, err := testhelpers.CreateSemester(
		db,
		uuid.New(),
		"Fall 2022",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		110.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership1 := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       false,
		Discounted: false,
	}
	res := db.Create(&membership1)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membership2 := models.Membership{
		UserID:     user2.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: false,
	}
	res = db.Create(&membership2)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membership3 := models.Membership{
		UserID:     user3.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: true,
	}
	res = db.Create(&membership3)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)

	limit := 1
	filter := models.ListMembershipsFilter{
		SemesterID: &semester1.ID,
		Limit:      &limit,
	}

	memberships, err := membershipService.ListMemberships(&filter)
	if err != nil {
		t.Errorf("ListMemberships() error = %v", err)
		return
	}

	if len(memberships) != 1 {
		t.Errorf("length = %v, want = %v", len(memberships), 2)
		return
	}

	if memberships[0].UserID != user1.ID {
		t.Errorf("result = %v, want = %v", memberships, []uint64{user1.ID})
		return
	}

	if memberships[0].Attendance != 0 {
		t.Errorf("ListMemberships attendance incorrect: %v, expected 0", memberships)
		return
	}
}

func TestMembershipService_ListMemberships_Offset(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}
	user2, err := testhelpers.CreateUser(db, 46372894, "deep", "kalra", "deep@gmail.com", models.FacultyArts, "dkal")
	if err != nil {
		t.Fatalf(err.Error())
	}
	user3, err := testhelpers.CreateUser(db, 41694873, "jane", "doe", "jane@gmail.com", models.FacultyMath, "jdane")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semester1, err := testhelpers.CreateSemester(
		db,
		uuid.New(),
		"Fall 2022",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		110.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership1 := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       false,
		Discounted: false,
	}
	res := db.Create(&membership1)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membership2 := models.Membership{
		UserID:     user2.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: false,
	}
	res = db.Create(&membership2)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membership3 := models.Membership{
		UserID:     user3.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: true,
	}
	res = db.Create(&membership3)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)

	offset := 1
	filter := models.ListMembershipsFilter{
		SemesterID: &semester1.ID,
		Offset:     &offset,
	}

	memberships, err := membershipService.ListMemberships(&filter)
	if err != nil {
		t.Errorf("ListMemberships() error = %v", err)
		return
	}

	if len(memberships) != 2 {
		t.Errorf("length = %v, want = %v", len(memberships), 2)
		return
	}

	if memberships[0].UserID != user2.ID && memberships[1].UserID != user3.ID {
		t.Errorf("result = %v, want = %v", memberships, []uint64{user2.ID, user3.ID})
		return
	}

	if memberships[0].Attendance != 0 || memberships[1].Attendance != 0 {
		t.Errorf("ListMemberships attendance incorrect: %v, expected 0", memberships)
		return
	}
}

func TestMembershipService_ListMemberships_LimitOffset(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}
	user2, err := testhelpers.CreateUser(db, 46372894, "deep", "kalra", "deep@gmail.com", models.FacultyArts, "dkal")
	if err != nil {
		t.Fatalf(err.Error())
	}
	user3, err := testhelpers.CreateUser(db, 41694873, "jane", "doe", "jane@gmail.com", models.FacultyMath, "jdane")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semester1, err := testhelpers.CreateSemester(
		db,
		uuid.New(),
		"Fall 2022",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		110.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership1 := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       false,
		Discounted: false,
	}
	res := db.Create(&membership1)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membership2 := models.Membership{
		UserID:     user2.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: false,
	}
	res = db.Create(&membership2)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membership3 := models.Membership{
		UserID:     user3.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: true,
	}
	res = db.Create(&membership3)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)

	limit := 1
	offset := 1
	filter := models.ListMembershipsFilter{
		SemesterID: &semester1.ID,
		Offset:     &offset,
		Limit:      &limit,
	}

	memberships, err := membershipService.ListMemberships(&filter)
	if err != nil {
		t.Errorf("ListMemberships() error = %v", err)
		return
	}

	if len(memberships) != 1 {
		t.Errorf("length = %v, want = %v", len(memberships), 2)
		return
	}

	if memberships[0].UserID != user2.ID {
		t.Errorf("result = %v, want = %v", memberships, []uint64{user2.ID})
		return
	}

	if memberships[0].Attendance != 0 {
		t.Errorf("ListMemberships attendance incorrect: %v, expected 0", memberships)
		return
	}
}

func TestMembershipService_UpdateMembership_InvalidUpdate(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		100.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       false,
		Discounted: false,
	}
	res := db.Create(&membership)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)

	_, err = membershipService.UpdateMembership(&models.UpdateMembershipRequest{
		ID:         membership.ID,
		Paid:       false,
		Discounted: true,
	})
	if err == nil {
		t.Errorf("UpdateMembership did not error, want error")
		return
	}
}

func TestMembershipService_UpdateMembership_UnpaidToPaid_NoDiscount(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		100.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       false,
		Discounted: false,
	}
	res := db.Create(&membership)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)
	updated, err := membershipService.UpdateMembership(&models.UpdateMembershipRequest{
		ID:         membership.ID,
		Paid:       true,
		Discounted: false,
	})

	if err != nil {
		t.Errorf("UpdateMembership() error = %v", err)
		return
	}

	if !updated.Paid && updated.Discounted {
		t.Errorf("got: Paid = %v, Discounted = %v | want: Paid = %v, Discounted = %v", updated.Paid, updated.Discounted, true, false)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}

	res = db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}
	if sem.MembershipFee != 10 {
		t.Fatalf("MegaFail: %v", sem.MembershipFee)
	}

	if !almostEqual(sem.CurrentBudget, 110.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 110.0)
		return
	}
}

func TestMembershipService_UpdateMembership_UnpaidToPaid_Discount(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		100.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       false,
		Discounted: false,
	}
	res := db.Create(&membership)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)
	updated, err := membershipService.UpdateMembership(&models.UpdateMembershipRequest{
		ID:         membership.ID,
		Paid:       true,
		Discounted: true,
	})

	if err != nil {
		t.Errorf("UpdateMembership() error = %v", err)
		return
	}

	if !updated.Paid && !updated.Discounted {
		t.Errorf("got: Paid = %v, Discounted = %v | want: Paid = %v, Discounted = %v", updated.Paid, updated.Discounted, true, true)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}
	res = db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}

	if !almostEqual(sem.CurrentBudget, 107.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 107.0)
		return
	}
}

func TestMembershipService_UpdateMembership_PaidToUnpaid_NotDiscounted(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		110.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: false,
	}
	res := db.Create(&membership)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)
	updated, err := membershipService.UpdateMembership(&models.UpdateMembershipRequest{
		ID:         membership.ID,
		Paid:       false,
		Discounted: false,
	})

	if err != nil {
		t.Errorf("UpdateMembership() error = %v", err)
		return
	}

	if updated.Paid && updated.Discounted {
		t.Errorf("got: Paid = %v, Discounted = %v | want: Paid = %v, Discounted = %v", updated.Paid, updated.Discounted, false, false)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}
	res = db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}

	if !almostEqual(sem.CurrentBudget, 100.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 100.0)
		return
	}
}

func TestMembershipService_UpdateMembership_PaidToUnpaid_Discounted(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		107.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: true,
	}
	res := db.Create(&membership)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)
	updated, err := membershipService.UpdateMembership(&models.UpdateMembershipRequest{
		ID:         membership.ID,
		Paid:       false,
		Discounted: false,
	})

	if err != nil {
		t.Errorf("UpdateMembership() error = %v", err)
		return
	}

	if updated.Paid && updated.Discounted {
		t.Errorf("got: Paid = %v, Discounted = %v | want: Paid = %v, Discounted = %v", updated.Paid, updated.Discounted, false, false)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}
	res = db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}

	if !almostEqual(sem.CurrentBudget, 100.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 100.0)
		return
	}
}

func TestMembershipService_UpdateMembership_Paid_NoDiscountToDiscount(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		110.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: false,
	}
	res := db.Create(&membership)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)
	updated, err := membershipService.UpdateMembership(&models.UpdateMembershipRequest{
		ID:         membership.ID,
		Paid:       true,
		Discounted: true,
	})

	if err != nil {
		t.Errorf("UpdateMembership() error = %v", err)
		return
	}

	if !updated.Paid && !updated.Discounted {
		t.Errorf("got: Paid = %v, Discounted = %v | want: Paid = %v, Discounted = %v", updated.Paid, updated.Discounted, true, true)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}
	res = db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}

	if !almostEqual(sem.CurrentBudget, 107.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 107.0)
		return
	}
}

func TestMembershipService_UpdateMembership_Paid_DiscountToNoDiscount(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer database.WipeDB(db)

	user1, err := testhelpers.CreateUser(db, 20780648, "adam", "mahood", "adam@gmail.com", models.FacultyAHS, "amahood")
	if err != nil {
		t.Fatalf(err.Error())
	}

	semesterId := uuid.New()
	semester1, err := testhelpers.CreateSemester(
		db,
		semesterId,
		"test",
		"",
		time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 1, 12, 0, 0, 0, time.UTC),
		100.0,
		107.0,
		10,
		7,
		2,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	membership := models.Membership{
		UserID:     user1.ID,
		SemesterID: semester1.ID,
		Paid:       true,
		Discounted: true,
	}
	res := db.Create(&membership)
	if res.Error != nil {
		t.Fatalf(res.Error.Error())
	}

	membershipService := NewMembershipService(db)
	updated, err := membershipService.UpdateMembership(&models.UpdateMembershipRequest{
		ID:         membership.ID,
		Paid:       true,
		Discounted: false,
	})

	if err != nil {
		t.Errorf("UpdateMembership() error = %v", err)
		return
	}

	if !updated.Paid && updated.Discounted {
		t.Errorf("got: Paid = %v, Discounted = %v | want: Paid = %v, Discounted = %v", updated.Paid, updated.Discounted, true, false)
		return
	}

	// Check that semester budget was updated
	sem := models.Semester{ID: semester1.ID}
	res = db.First(&sem)
	if res.Error != nil {
		t.Fatalf("Error when retrieving semester: %v", res.Error)
		return
	}

	if !almostEqual(sem.CurrentBudget, 110.0) {
		t.Errorf("CurrentBudget = %v, expected = %v", sem.CurrentBudget, 110.0)
		return
	}
}
