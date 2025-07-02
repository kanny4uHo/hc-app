package entity

type AddUserArgs struct {
	Login    string
	Password string
	Meta     UserMeta
}

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	Meta         UserMeta
}

type UserMeta struct {
	Name  UserName
	Email string
}

type UserName struct {
	First string
	Last  string
}

func (user User) IsEmpty() bool {
	return user.ID == 0
}
