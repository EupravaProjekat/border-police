package Models

type User struct {
	Email    string
	Role     string
	Requests []Request
}

// SpecialCross or temporaryExport
type Request struct {
	Uuid           string
	RequestType    string
	CarPlateNumber string
	Description    string
}
type GetRequest struct {
	Uuid string
}
