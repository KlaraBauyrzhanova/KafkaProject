package message

type Message struct {
	ID    int    `json:"id" db:"id"`
	FName string `json:"f_name" db:"f_name"`
	LName string `json:"l_name" db:"l_name"`
}
