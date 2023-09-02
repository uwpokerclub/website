package testhelpers

import (
	"api/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateSemester(db *gorm.DB, id uuid.UUID, name string, meta string, startDate time.Time, endDate time.Time, startingBudget float64, currentBudget float64, membershipFee int8, discountFee int8, rebuyFee int8) (*models.Semester, error) {
	semester := models.Semester{
		ID:                    id,
		Name:                  name,
		Meta:                  meta,
		StartDate:             startDate,
		EndDate:               endDate,
		StartingBudget:        startingBudget,
		CurrentBudget:         currentBudget,
		MembershipFee:         membershipFee,
		MembershipDiscountFee: discountFee,
		RebuyFee:              rebuyFee,
	}

	res := db.Create(&semester)
	if res.Error != nil {
		return nil, res.Error
	}

	return &semester, nil
}

func CreateUser(db *gorm.DB, id uint64, firstName string, lastName string, email string, faculty string, questId string) (*models.User, error) {
	user := models.User{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Faculty:   faculty,
		QuestID:   questId,
	}

	res := db.Create(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}

func CreateEvent(db *gorm.DB, name string, semesterId uuid.UUID, startDate time.Time) (*models.Event, error) {
	structure := models.Structure{
		Name: "Main Event Structure",
	}
	res := db.Create(&structure)
	if res.Error != nil {
		return nil, res.Error
	}

	event := models.Event{
		Name:        name,
		Format:      "NLHE",
		Notes:       "",
		SemesterID:  semesterId,
		StartDate:   startDate,
		State:       models.EventStateStarted,
		StructureID: structure.ID,
	}

	res = db.Create(&event)
	if res.Error != nil {
		return nil, res.Error
	}

	return &event, nil
}

func CreateMembership(db *gorm.DB, userId uint64, semesterId uuid.UUID, paid bool, discounted bool) (*models.Membership, error) {
	membership := models.Membership{
		UserID:     userId,
		SemesterID: semesterId,
		Paid:       paid,
		Discounted: discounted,
	}

	res := db.Create(&membership)
	if res.Error != nil {
		return nil, res.Error
	}

	return &membership, nil
}

func CreateParticipant(db *gorm.DB, membershipId uuid.UUID, eventId uint64, placement uint32, signedOutAt *time.Time, rebuys uint8) (*models.Participant, error) {
	entry := models.Participant{
		MembershipID: membershipId,
		EventID:      eventId,
		Placement:    placement,
		SignedOutAt:  signedOutAt,
		Rebuys:       rebuys,
	}

	res := db.Create(&entry)
	if res.Error != nil {
		return nil, res.Error
	}

	return &entry, nil
}

type SemesterSetupRet struct {
	Semester    models.Semester
	Users       []models.User
	Memberships []models.Membership
}

func SetupSemester(db *gorm.DB, title string) (*SemesterSetupRet, error) {
	// Create semester
	semesterId := uuid.New()
	semester, err := CreateSemester(
		db,
		semesterId,
		title,
		"",
		time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC),
		100,
		110,
		10,
		7,
		2,
	)
	if err != nil {
		return nil, err
	}

	// Create a few users
	user1, err := CreateUser(db, 20780648, "Adam", "Mahood", "adam@gmail.com", models.FacultyMath, "asmahood")
	if err != nil {
		return nil, err
	}
	user2, err := CreateUser(db, 36459367, "Deep", "Kalra", "deep@gmail.com", models.FacultyEngineering, "d2kal")
	if err != nil {
		return nil, err
	}
	user3, err := CreateUser(db, 13274944, "Jane", "Doe", "jane@gmail.com", models.FacultyArts, "jdoe")
	if err != nil {
		return nil, err
	}

	// Create a few memberships
	membership1, err := CreateMembership(db, user1.ID, semester.ID, true, true)
	if err != nil {
		return nil, err
	}

	membership2, err := CreateMembership(db, user2.ID, semester.ID, true, false)
	if err != nil {
		return nil, err
	}

	membership3, err := CreateMembership(db, user3.ID, semester.ID, false, false)
	if err != nil {
		return nil, err
	}

	return &SemesterSetupRet{
		Semester:    *semester,
		Users:       []models.User{*user1, *user2, *user3},
		Memberships: []models.Membership{*membership1, *membership2, *membership3},
	}, nil
}
